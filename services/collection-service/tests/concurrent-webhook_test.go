package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/repository"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/server"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/service"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func getTestDSN(t *testing.T) string {
	// Hard opt-in to prevent accidentally running destructive DB setup against a non-test database.
	switch os.Getenv("RUN_COLLECTION_SERVICE_INTEGRATION_TESTS") {
	case "1", "true", "TRUE", "yes", "YES":
		// ok
	default:
		t.Skip("skipping: set RUN_COLLECTION_SERVICE_INTEGRATION_TESTS=1 to enable integration tests")
	}

	// Prefer a collection-service specific DSN; fall back to a shared test DSN.
	if dsn := os.Getenv("COLLECTION_SERVICE_TEST_DATABASE_DSN"); dsn != "" {
		return dsn
	}
	if dsn := os.Getenv("TEST_DATABASE_DSN"); dsn != "" {
		return dsn
	}

	t.Skip("skipping: set COLLECTION_SERVICE_TEST_DATABASE_DSN (PostgreSQL) to run integration tests")
	return ""
}

func setupPostgres(t *testing.T) *gorm.DB {
	dsn := getTestDSN(t)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Ensure uuid extension exists for DEFAULT uuid_generate_v4().
	// If the test DB doesn't allow creating extensions, skip.
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		t.Skipf("skipping: test DB cannot create uuid-ossp extension: %v", err)
	}

	// Minimal schema needed for webhook idempotency.
	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS processed_events (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			event_id VARCHAR(200) NOT NULL UNIQUE,
			event_type VARCHAR(50) NOT NULL,
			chain_id VARCHAR(50) NOT NULL,
			block_number BIGINT NOT NULL,
			tx_hash VARCHAR(66) NOT NULL,
			log_index INTEGER NOT NULL,
			collection_address VARCHAR(42),
			processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`).Error)

	// Clean slate per test file run.
	require.NoError(t, db.Exec("DELETE FROM processed_events").Error)

	return db
}

func createWebhookRequest(t *testing.T, eventType string, txHash string, logIndex int64) *pb.ProcessIndexerWebhookRequest {
	eventData := server.WebhookEventData{
		ContractAddress: "0x1234567890123456789012345678901234567890",
		Creator:         "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		TokenType:       "ERC721",
		BlockNumber:     12345,
		TxHash:          txHash,
		LogIndex:        logIndex,
	}

	dataJSON, err := json.Marshal(eventData)
	require.NoError(t, err)

	return &pb.ProcessIndexerWebhookRequest{
		Event:     eventType,
		ChainId:   1,
		Timestamp: time.Now().Unix(),
		DataJson:  string(dataJSON),
	}
}

func TestConcurrentWebhookProcessing_AdvisoryLockSerializesDuplicates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short")
	}

	db := setupPostgres(t)

	// Real repos/services; the handler path used below does not require collections table.
	collectionRepo := repository.NewCollectionRepository(db)
	allowlistRepo := repository.NewAllowlistRepository(db)
	metadataRepo := repository.NewMetadataRepository(db)
	processedEventRepo := repository.NewProcessedEventRepository(db)

	collectionSvc := service.NewCollectionService(collectionRepo, allowlistRepo, metadataRepo)
	zapLogger := zap.NewNop()
	srv := server.NewCollectionServer(collectionSvc, processedEventRepo, zapLogger)

	// Use an event type that succeeds even when the collection is not present in DB.
	// Only one request should actually handle the event; others should be either Aborted (lock contention)
	// or return "Event already processed" (arrived after the first commit).
	req := createWebhookRequest(t, "collection.minted", fmt.Sprintf("0x%064x", 1), 0)

	const n = 100
	start := make(chan struct{})
	var wg sync.WaitGroup

	var handledCount int32
	var alreadyProcessedCount int32
	var abortedCount int32
	var canceledCount int32
	var otherErrCount int32

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start

			resp, err := srv.ProcessIndexerWebhook(context.Background(), req)
			if err != nil {
				switch status.Code(err) {
				case codes.Aborted:
					atomic.AddInt32(&abortedCount, 1)
				case codes.Canceled:
					atomic.AddInt32(&canceledCount, 1)
				default:
					atomic.AddInt32(&otherErrCount, 1)
				}
				return
			}

			if resp == nil {
				atomic.AddInt32(&otherErrCount, 1)
				return
			}

			if resp.Message == "Event already processed" {
				atomic.AddInt32(&alreadyProcessedCount, 1)
				return
			}
			atomic.AddInt32(&handledCount, 1)
		}()
	}

	close(start)
	wg.Wait()

	require.Equal(t, int32(0), canceledCount, "no requests should fail with context cancellation")
	require.Equal(t, int32(0), otherErrCount, "no unexpected errors")
	require.Equal(t, int32(1), handledCount, "exactly one request should handle the event")

	// Verify only 1 processed_event record exists.
	var count int64
	require.NoError(t, db.Table("processed_events").Count(&count).Error)
	require.Equal(t, int64(1), count)

	// Sanity: the remaining requests should be explained by either lock contention or already-processed.
	require.Equal(t, int32(n)-handledCount, abortedCount+alreadyProcessedCount)
}

func TestConcurrentWebhookProcessing_DifferentEventsDoNotBlock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short")
	}

	db := setupPostgres(t)

	collectionRepo := repository.NewCollectionRepository(db)
	allowlistRepo := repository.NewAllowlistRepository(db)
	metadataRepo := repository.NewMetadataRepository(db)
	processedEventRepo := repository.NewProcessedEventRepository(db)

	collectionSvc := service.NewCollectionService(collectionRepo, allowlistRepo, metadataRepo)
	zapLogger := zap.NewNop()
	srv := server.NewCollectionServer(collectionSvc, processedEventRepo, zapLogger)

	const n = 25
	start := make(chan struct{})
	var wg sync.WaitGroup
	var okCount int32
	var errCount int32

	for i := 0; i < n; i++ {
		wg.Add(1)
		req := createWebhookRequest(t, "collection.minted", fmt.Sprintf("0x%064x", i+100), int64(i))

		go func(r *pb.ProcessIndexerWebhookRequest) {
			defer wg.Done()
			<-start
			resp, err := srv.ProcessIndexerWebhook(context.Background(), r)
			if err != nil || resp == nil || !resp.Success {
				atomic.AddInt32(&errCount, 1)
				return
			}
			atomic.AddInt32(&okCount, 1)
		}(req)
	}

	close(start)
	wg.Wait()

	require.Equal(t, int32(0), errCount)
	require.Equal(t, int32(n), okCount)

	var count int64
	require.NoError(t, db.Table("processed_events").Count(&count).Error)
	require.Equal(t, int64(n), count)
}
