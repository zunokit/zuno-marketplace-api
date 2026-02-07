package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/repository"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"github.com/zunokit/zuno-marketplace-api/shared/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WebhookEventData represents the data from indexer webhook
type WebhookEventData struct {
	CollectionAddress string `json:"collectionAddress"`
	Creator           string `json:"creator"`
	TokenType         string `json:"tokenType"`
	BlockNumber       int64  `json:"blockNumber"`
	TxHash            string `json:"txHash"`
	LogIndex          int64  `json:"logIndex"`
	ContractAddress   string `json:"contractAddress"`
}

// ProcessIndexerWebhook handles webhook events from the indexer
func (s *CollectionServer) ProcessIndexerWebhook(
	ctx context.Context,
	req *pb.ProcessIndexerWebhookRequest,
) (*pb.ProcessIndexerWebhookResponse, error) {
	// Build chain ID in CAIP-2 format
	chainID := fmt.Sprintf("eip155:%d", req.ChainId)

	// Log incoming webhook
	s.logger.Debug("Processing webhook",
		zap.String("event_type", req.Event),
		zap.String("chain_id", chainID),
		zap.Int64("timestamp", req.Timestamp),
	)

	// 1. Parse event data
	var eventData WebhookEventData
	if err := json.Unmarshal([]byte(req.DataJson), &eventData); err != nil {
		s.logger.Error("Failed to parse event data",
			zap.String("event_type", req.Event),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.InvalidArgument, "invalid event data: %v", err)
	}

	// 2. Validate payload
	validator := NewWebhookValidator(true) // Allow external collections
	if err := validator.ValidateRequest(req.Event, req.ChainId, req.Timestamp, eventDataToMap(eventData)); err != nil {
		s.logger.Error("Webhook validation failed",
			zap.String("event_type", req.Event),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	// 3. Build event ID for idempotency
	eventID := BuildEventID(eventData.TxHash, eventData.LogIndex)

	// 4-6. Serialize idempotency check + marking using a transaction-scoped advisory lock.
	// The transaction is held open for the duration of processing to keep the xact lock alive.
	var response *pb.ProcessIndexerWebhookResponse
	err := s.processedEventRepo.WithTransaction(ctx, func(txRepo repository.ProcessedEventRepository) error {
		// Acquire advisory lock BEFORE idempotency check to prevent race conditions.
		if err := txRepo.AcquireEventLock(ctx, eventID); err != nil {
			if errors.Is(err, repository.ErrLockNotAcquired) {
				s.logger.Info("Event already being processed by another transaction",
					zap.String("event_id", eventID),
					zap.String("event_type", req.Event),
				)
				return status.Error(codes.Aborted, "event lock not acquired: another transaction processing this event")
			}
			return status.Errorf(codes.Internal, "failed to acquire lock: %v", err)
		}

		processed, err := txRepo.IsEventProcessed(ctx, eventID)
		if err != nil {
			s.logger.Error("Failed to check if event was processed",
				zap.String("event_id", eventID),
				zap.Error(err),
			)
			return status.Errorf(codes.Internal, "failed to check event status: %v", err)
		}

		if processed {
			s.logger.Info("Event already processed, skipping",
				zap.String("event_id", eventID),
				zap.String("event_type", req.Event),
			)
			response = &pb.ProcessIndexerWebhookResponse{Success: true, Message: "Event already processed"}
			return nil
		}

		// Handle event based on type
		var handleErr error
		switch req.Event {
		case "collection.created":
			response, handleErr = s.handleCollectionCreated(ctx, req, eventData)
		case "collection.minted":
			response, handleErr = s.handleCollectionMinted(ctx, req, eventData)
		case "collection.batch_minted":
			response, handleErr = s.handleCollectionBatchMinted(ctx, req, eventData)
		default:
			return status.Errorf(codes.InvalidArgument, "unknown event type: %s", req.Event)
		}
		if handleErr != nil {
			return handleErr
		}

		// Mark event as processed if handling succeeded
		if response != nil && response.Success {
			processedEvent := models.NewProcessedEvent(
				eventID,
				req.Event,
				chainID,
				eventData.BlockNumber,
				eventData.TxHash,
				int(eventData.LogIndex),
				getCollectionAddress(eventData),
			)

			if err := txRepo.CreateProcessedEvent(ctx, processedEvent); err != nil {
				// Log error but don't fail the request (keep existing behavior)
				s.logger.Error("Failed to mark event as processed",
					zap.String("event_id", eventID),
					zap.Error(err),
				)
			} else {
				s.logger.Info("Event marked as processed",
					zap.String("event_id", eventID),
					zap.String("event_type", req.Event),
				)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

// handleCollectionCreated processes collection.created webhook events
func (s *CollectionServer) handleCollectionCreated(
	ctx context.Context,
	req *pb.ProcessIndexerWebhookRequest,
	data WebhookEventData,
) (*pb.ProcessIndexerWebhookResponse, error) {
	// Build chainID in eip155 format
	chainID := fmt.Sprintf("eip155:%d", req.ChainId)

	// Normalize address for case-insensitive lookup
	normalizedAddress := utils.NormalizeAddress(data.CollectionAddress)

	s.logger.Info("Processing collection.created event",
		zap.String("collection_address", normalizedAddress),
		zap.String("original_address", data.CollectionAddress),
		zap.String("chain_id", chainID),
		zap.String("creator", data.Creator),
		zap.String("token_type", data.TokenType),
		zap.Int64("block_number", data.BlockNumber),
	)

	// Find collection by contract address and chainID
	dbCollection, err := s.service.GetCollectionByContract(ctx, normalizedAddress, chainID)
	if err != nil {
		s.logger.Warn("Collection not found in database",
			zap.String("collection_address", normalizedAddress),
			zap.String("chain_id", chainID),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.NotFound, "collection not found: %v", err)
	}

	s.logger.Debug("Collection found, updating status",
		zap.String("collection_id", dbCollection.ID.String()),
		zap.String("collection_address", data.CollectionAddress),
		zap.String("current_status", string(dbCollection.Status)),
	)

	// Update status to DEPLOYED and index_status to INDEXED
	updates := map[string]interface{}{
		"status":       models.CollectionStatusDeployed,
		"index_status": models.IndexStatusSynced,
		"indexed_at":   time.Now(),
		"updated_at":   time.Now(),
	}

	// Update in database (using owner's ID for authorization)
	if _, err := s.service.UpdateCollection(ctx, dbCollection.ID, dbCollection.UserID, updates); err != nil {
		s.logger.Error("Failed to update collection",
			zap.String("collection_id", dbCollection.ID.String()),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "failed to update collection: %v", err)
	}

	s.logger.Info("Collection indexed successfully",
		zap.String("collection_id", dbCollection.ID.String()),
		zap.String("collection_address", normalizedAddress),
		zap.String("status", string(models.CollectionStatusDeployed)),
	)

	return &pb.ProcessIndexerWebhookResponse{
		Success: true,
		Message: fmt.Sprintf("Collection %s indexed successfully", normalizedAddress),
	}, nil
}

// handleCollectionMinted processes collection.minted webhook events
func (s *CollectionServer) handleCollectionMinted(
	ctx context.Context,
	req *pb.ProcessIndexerWebhookRequest,
	data WebhookEventData,
) (*pb.ProcessIndexerWebhookResponse, error) {
	// Build chainID in eip155 format
	chainID := fmt.Sprintf("eip155:%d", req.ChainId)

	// Normalize address for case-insensitive lookup
	normalizedAddress := utils.NormalizeAddress(data.ContractAddress)

	s.logger.Debug("Processing collection.minted event",
		zap.String("contract_address", normalizedAddress),
		zap.String("chain_id", chainID),
	)

	// Find collection by contract address and chainID
	dbCollection, err := s.service.GetCollectionByContract(ctx, normalizedAddress, chainID)
	if err != nil {
		// Collection might not be in our database yet (external collection)
		s.logger.Info("External collection not in database, skipping stats update",
			zap.String("contract_address", normalizedAddress),
			zap.String("chain_id", chainID),
		)
		return &pb.ProcessIndexerWebhookResponse{
			Success: true,
			Message: "Collection not in database, skipping stats update",
		}, nil
	}

	// Increment total_minted counter atomically
	if err := s.service.IncrementTotalMinted(ctx, dbCollection.ID, 1); err != nil {
		s.logger.Error("Failed to update collection stats",
			zap.String("collection_id", dbCollection.ID.String()),
			zap.String("contract_address", normalizedAddress),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "failed to update collection stats: %v", err)
	}

	s.logger.Debug("Collection stats updated",
		zap.String("collection_id", dbCollection.ID.String()),
		zap.String("contract_address", normalizedAddress),
	)

	return &pb.ProcessIndexerWebhookResponse{
		Success: true,
		Message: fmt.Sprintf("Collection %s stats updated", normalizedAddress),
	}, nil
}

// handleCollectionBatchMinted processes collection.batch_minted webhook events
func (s *CollectionServer) handleCollectionBatchMinted(
	ctx context.Context,
	req *pb.ProcessIndexerWebhookRequest,
	data WebhookEventData,
) (*pb.ProcessIndexerWebhookResponse, error) {
	// Build chainID in eip155 format
	chainID := fmt.Sprintf("eip155:%d", req.ChainId)

	// Normalize address for case-insensitive lookup
	normalizedAddress := utils.NormalizeAddress(data.ContractAddress)

	s.logger.Debug("Processing collection.batch_minted event",
		zap.String("contract_address", normalizedAddress),
		zap.String("chain_id", chainID),
	)

	// Find collection by contract address and chainID
	dbCollection, err := s.service.GetCollectionByContract(ctx, normalizedAddress, chainID)
	if err != nil {
		// Collection might not be in our database yet (external collection)
		s.logger.Info("External collection not in database, skipping batch stats update",
			zap.String("contract_address", normalizedAddress),
			zap.String("chain_id", chainID),
		)
		return &pb.ProcessIndexerWebhookResponse{
			Success: true,
			Message: "Collection not in database, skipping stats update",
		}, nil
	}

	// For batch_minted, we'd need to extract the batch size from the event data
	// For now, we'll increment by 1 (this can be enhanced later)
	// TODO: Extract batch size from event data when available
	batchSize := int64(1)

	// Increment total_minted counter atomically
	if err := s.service.IncrementTotalMinted(ctx, dbCollection.ID, batchSize); err != nil {
		s.logger.Error("Failed to update collection stats",
			zap.String("collection_id", dbCollection.ID.String()),
			zap.String("contract_address", normalizedAddress),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "failed to update collection stats: %v", err)
	}

	s.logger.Debug("Collection stats updated for batch mint",
		zap.String("collection_id", dbCollection.ID.String()),
		zap.String("contract_address", normalizedAddress),
		zap.Int64("batch_size", batchSize),
	)

	return &pb.ProcessIndexerWebhookResponse{
		Success: true,
		Message: fmt.Sprintf("Collection %s stats updated for batch mint", normalizedAddress),
	}, nil
}

// Helper functions

// eventDataToMap converts WebhookEventData to map for validation
func eventDataToMap(data WebhookEventData) map[string]interface{} {
	return map[string]interface{}{
		"collectionAddress": data.CollectionAddress,
		"creator":           data.Creator,
		"tokenType":         data.TokenType,
		"blockNumber":       float64(data.BlockNumber),
		"txHash":            data.TxHash,
		"logIndex":          float64(data.LogIndex),
		"contractAddress":   data.ContractAddress,
	}
}

// getCollectionAddress extracts the collection address from event data
func getCollectionAddress(data WebhookEventData) string {
	if data.CollectionAddress != "" {
		return data.CollectionAddress
	}
	return data.ContractAddress
}

// VerifyWebhookSignature verifies HMAC-SHA256 signature
func VerifyWebhookSignature(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
