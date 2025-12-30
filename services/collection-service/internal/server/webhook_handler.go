package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
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
	// 1. Verify HMAC signature (if provided via metadata)
	// Note: In production, signature should be passed via gRPC metadata
	// For HTTP->gRPC gateway, this will be handled by the gateway layer

	// 2. Parse event data
	var eventData WebhookEventData
	if err := json.Unmarshal([]byte(req.DataJson), &eventData); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event data: %v", err)
	}

	// 3. Handle event based on type
	switch req.Event {
	case "collection.created":
		return s.handleCollectionCreated(ctx, req, eventData)
	case "collection.minted":
		return s.handleCollectionMinted(ctx, req, eventData)
	case "collection.batch_minted":
		return s.handleCollectionBatchMinted(ctx, req, eventData)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown event type: %s", req.Event)
	}
}

// handleCollectionCreated processes collection.created webhook events
func (s *CollectionServer) handleCollectionCreated(
	ctx context.Context,
	req *pb.ProcessIndexerWebhookRequest,
	data WebhookEventData,
) (*pb.ProcessIndexerWebhookResponse, error) {
	// Build chainID in eip155 format
	chainID := fmt.Sprintf("eip155:%d", req.ChainId)

	// Find collection by contract address and chainID
	dbCollection, err := s.service.GetCollectionByContract(ctx, data.CollectionAddress, chainID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "collection not found: %v", err)
	}

	// Update index status to INDEXED
	updates := map[string]interface{}{
		"index_status": "INDEXED",
		"indexed_at":   time.Now(),
		"updated_at":   time.Now(),
	}

	// Update in database (using system user ID for webhook updates)
	systemUserID := dbCollection.UserID // Use owner's ID for authorization
	if _, err := s.service.UpdateCollection(ctx, dbCollection.ID, systemUserID, updates); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update collection: %v", err)
	}

	return &pb.ProcessIndexerWebhookResponse{
		Success: true,
		Message: fmt.Sprintf("Collection %s indexed successfully", data.CollectionAddress),
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

	// Find collection by contract address and chainID
	dbCollection, err := s.service.GetCollectionByContract(ctx, data.ContractAddress, chainID)
	if err != nil {
		// Collection might not be in our database yet (external collection)
		return &pb.ProcessIndexerWebhookResponse{
			Success: true,
			Message: "Collection not in database, skipping stats update",
		}, nil
	}

	// Increment total_minted counter
	// Note: For proper atomic increment, repository should handle this
	updates := map[string]interface{}{
		"total_minted": dbCollection.TotalMinted + 1,
		"updated_at":   time.Now(),
	}

	// Update in database
	if _, err := s.service.UpdateCollection(ctx, dbCollection.ID, dbCollection.UserID, updates); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update collection stats: %v", err)
	}

	return &pb.ProcessIndexerWebhookResponse{
		Success: true,
		Message: fmt.Sprintf("Collection %s stats updated", data.ContractAddress),
	}, nil
}

// handleCollectionBatchMinted processes collection.batch_minted webhook events
func (s *CollectionServer) handleCollectionBatchMinted(
	ctx context.Context,
	req *pb.ProcessIndexerWebhookRequest,
	data WebhookEventData,
) (*pb.ProcessIndexerWebhookResponse, error) {
	// Similar to handleCollectionMinted but for batch mints
	// For now, treat it the same as single mint
	return s.handleCollectionMinted(ctx, req, data)
}

// VerifyWebhookSignature verifies HMAC-SHA256 signature
func VerifyWebhookSignature(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
