package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupProcessedEventTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Create table
	err = db.AutoMigrate(&models.ProcessedEvent{})
	require.NoError(t, err)

	return db
}

func TestProcessedEventRepository_IsEventProcessed(t *testing.T) {
	db := setupProcessedEventTestDB(t)
	repo := NewProcessedEventRepository(db)
	ctx := context.Background()

	t.Run("Event not processed initially", func(t *testing.T) {
		processed, err := repo.IsEventProcessed(ctx, "0xabc123:0")
		assert.NoError(t, err)
		assert.False(t, processed)
	})

	t.Run("Event processed after creation", func(t *testing.T) {
		event := models.NewProcessedEvent(
			"0xabc123:0",
			"collection.created",
			"eip155:1",
			12345,
			"0xabc123",
			0,
			"0x1234567890123456789012345678901234567890",
		)

		err := repo.CreateProcessedEvent(ctx, event)
		require.NoError(t, err)

		processed, err := repo.IsEventProcessed(ctx, "0xabc123:0")
		assert.NoError(t, err)
		assert.True(t, processed)
	})
}

func TestProcessedEventRepository_CreateProcessedEvent(t *testing.T) {
	db := setupProcessedEventTestDB(t)
	repo := NewProcessedEventRepository(db)
	ctx := context.Background()

	t.Run("Create new event successfully", func(t *testing.T) {
		event := models.NewProcessedEvent(
			"0xdef456:1",
			"collection.minted",
			"eip155:1",
			12346,
			"0xdef456",
			1,
			"0x0987654321098765432109876543210987654321",
		)

		err := repo.CreateProcessedEvent(ctx, event)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, event.ID)
		assert.WithinDuration(t, time.Now(), event.ProcessedAt, time.Second)
	})

	t.Run("Duplicate event returns error", func(t *testing.T) {
		event := models.NewProcessedEvent(
			"0xduplicate:0",
			"collection.created",
			"eip155:1",
			12347,
			"0xduplicate",
			0,
			"0x1111111111111111111111111111111111111111",
		)

		// First creation should succeed
		err := repo.CreateProcessedEvent(ctx, event)
		assert.NoError(t, err)

		// Second creation should fail
		err = repo.CreateProcessedEvent(ctx, event)
		assert.Error(t, err)
		assert.Equal(t, ErrEventAlreadyProcessed, err)
	})
}

func TestProcessedEventRepository_GetProcessedEvent(t *testing.T) {
	db := setupProcessedEventTestDB(t)
	repo := NewProcessedEventRepository(db)
	ctx := context.Background()

	t.Run("Get existing event", func(t *testing.T) {
		originalEvent := models.NewProcessedEvent(
			"0xgettest:0",
			"collection.created",
			"eip155:1",
			12348,
			"0xgettest",
			0,
			"0x2222222222222222222222222222222222222222",
		)

		err := repo.CreateProcessedEvent(ctx, originalEvent)
		require.NoError(t, err)

		retrievedEvent, err := repo.GetProcessedEvent(ctx, "0xgettest:0")
		assert.NoError(t, err)
		assert.NotNil(t, retrievedEvent)
		assert.Equal(t, originalEvent.EventID, retrievedEvent.EventID)
		assert.Equal(t, originalEvent.EventType, retrievedEvent.EventType)
		assert.Equal(t, originalEvent.ChainID, retrievedEvent.ChainID)
	})

	t.Run("Get non-existent event returns nil", func(t *testing.T) {
		event, err := repo.GetProcessedEvent(ctx, "0xnonexistent:0")
		assert.NoError(t, err)
		assert.Nil(t, event)
	})
}

func TestProcessedEventRepository_DeleteProcessedEventsBefore(t *testing.T) {
	db := setupProcessedEventTestDB(t)
	repo := NewProcessedEventRepository(db)
	ctx := context.Background()

	// Create events with different timestamps
	oldTime := time.Now().Add(-48 * time.Hour)
	recentTime := time.Now().Add(-1 * time.Hour)

	oldEvent := &models.ProcessedEvent{
		EventID:     "0xoldevent:0",
		EventType:   "collection.created",
		ChainID:     "eip155:1",
		BlockNumber: 12349,
		TxHash:      "0xoldevent",
		LogIndex:    0,
		ProcessedAt: oldTime,
	}

	recentEvent := &models.ProcessedEvent{
		EventID:     "0xrecentevent:0",
		EventType:   "collection.created",
		ChainID:     "eip155:1",
		BlockNumber: 12350,
		TxHash:      "0xrecentevent",
		LogIndex:    0,
		ProcessedAt: recentTime,
	}

	err := repo.CreateProcessedEvent(ctx, oldEvent)
	require.NoError(t, err)
	err = repo.CreateProcessedEvent(ctx, recentEvent)
	require.NoError(t, err)

	t.Run("Delete events before cutoff", func(t *testing.T) {
		cutoff := time.Now().Add(-24 * time.Hour)
		deletedCount, err := repo.DeleteProcessedEventsBefore(ctx, cutoff)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), deletedCount)

		// Verify old event is deleted
		_, err = repo.GetProcessedEvent(ctx, "0xoldevent:0")
		assert.Error(t, err)

		// Verify recent event still exists
		event, err := repo.GetProcessedEvent(ctx, "0xrecentevent:0")
		assert.NoError(t, err)
		assert.NotNil(t, event)
	})
}
