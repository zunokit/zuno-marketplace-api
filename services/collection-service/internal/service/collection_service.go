package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/quangdang46/NFT-Marketplace/services/collection-service/internal/models"
	"github.com/quangdang46/NFT-Marketplace/services/collection-service/internal/repository"
)

var (
	ErrUnauthorized        = errors.New("unauthorized: only owner can perform this action")
	ErrInvalidStatusChange = errors.New("invalid status change")
	ErrContractRequired    = errors.New("contract address required for deployed status")
)

type CollectionService struct {
	collectionRepo repository.CollectionRepository
	allowlistRepo  repository.AllowlistRepository
	metadataRepo   repository.MetadataRepository
}

func NewCollectionService(
	collectionRepo repository.CollectionRepository,
	allowlistRepo repository.AllowlistRepository,
	metadataRepo repository.MetadataRepository,
) *CollectionService {
	return &CollectionService{
		collectionRepo: collectionRepo,
		allowlistRepo:  allowlistRepo,
		metadataRepo:   metadataRepo,
	}
}

func (s *CollectionService) CreateCollection(ctx context.Context, collection *models.Collection) (*models.Collection, error) {
	// Set defaults
	collection.Status = models.CollectionStatusPending
	collection.IndexStatus = models.IndexStatusNotStarted

	// Create in DB
	if err := s.collectionRepo.Create(ctx, collection); err != nil {
		return nil, err
	}

	// Retrieve created collection with relations
	return s.collectionRepo.GetByID(ctx, collection.ID)
}

func (s *CollectionService) GetCollection(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
	return s.collectionRepo.GetByID(ctx, id)
}

func (s *CollectionService) GetCollectionByContract(ctx context.Context, address, chainID string) (*models.Collection, error) {
	return s.collectionRepo.GetByContractAddress(ctx, address, chainID)
}

func (s *CollectionService) UpdateCollection(ctx context.Context, id uuid.UUID, userID uuid.UUID, updates map[string]interface{}) (*models.Collection, error) {
	// 1. Get existing collection
	collection, err := s.collectionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Verify ownership
	if collection.UserID != userID {
		return nil, ErrUnauthorized
	}

	// 3. Validate status change
	if newStatus, ok := updates["status"].(models.CollectionStatus); ok {
		if collection.Status == models.CollectionStatusDeployed && newStatus == models.CollectionStatusPending {
			return nil, ErrInvalidStatusChange
		}
		if newStatus == models.CollectionStatusDeployed {
			if contractAddr, ok := updates["contract_address"].(string); !ok || contractAddr == "" {
				// Check if already has contract address
				if collection.ContractAddress == nil || *collection.ContractAddress == "" {
					return nil, ErrContractRequired
				}
			}
		}
	}

	// 4. Apply updates
	if err := s.collectionRepo.Update(ctx, id, updates); err != nil {
		return nil, err
	}

	// 5. Return updated collection
	return s.collectionRepo.GetByID(ctx, id)
}

func (s *CollectionService) UpdateMetadata(ctx context.Context, collectionID uuid.UUID, userID uuid.UUID, metadataUpdates map[string]interface{}) error {
	// 1. Verify ownership
	collection, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return err
	}
	if collection.UserID != userID {
		return ErrUnauthorized
	}

	// 2. Update metadata
	if len(metadataUpdates) > 0 {
		return s.metadataRepo.Update(ctx, collectionID, metadataUpdates)
	}

	return nil
}

func (s *CollectionService) ListCollectionsByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.Collection, int64, error) {
	return s.collectionRepo.ListByUser(ctx, userID, page, limit)
}

func (s *CollectionService) ListCollections(ctx context.Context, filters *repository.ListFilters, page, limit int) ([]*models.Collection, int64, error) {
	return s.collectionRepo.List(ctx, filters, page, limit)
}

func (s *CollectionService) AddToAllowlist(ctx context.Context, collectionID uuid.UUID, userID uuid.UUID, wallets []string, maxMintAmount int) (int, error) {
	// 1. Verify ownership
	collection, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return 0, err
	}
	if collection.UserID != userID {
		return 0, ErrUnauthorized
	}

	// 2. Prepare records
	var allowlistEntries []*models.CollectionAllowlist
	for _, addr := range wallets {
		allowlistEntries = append(allowlistEntries, &models.CollectionAllowlist{
			CollectionID:  collectionID,
			WalletAddress: addr,
			MaxMintAmount: maxMintAmount,
			AddedByUserID: &userID,
			AddedAt:       time.Now(),
		})
	}

	// 3. Add to DB
	if err := s.allowlistRepo.AddWallets(ctx, allowlistEntries); err != nil {
		return 0, err
	}

	return len(wallets), nil
}

func (s *CollectionService) DeleteCollection(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// 1. Verify ownership
	collection, err := s.collectionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if collection.UserID != userID {
		return ErrUnauthorized
	}

	// 2. Delete
	return s.collectionRepo.Delete(ctx, id)
}
