package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/quangdang46/NFT-Marketplace/services/collection-service/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AllowlistRepository interface {
	AddWallets(ctx context.Context, wallets []*models.CollectionAllowlist) error
	GetByCollection(ctx context.Context, collectionID uuid.UUID) ([]*models.CollectionAllowlist, error)
	RemoveWallet(ctx context.Context, collectionID uuid.UUID, walletAddress string) error
}

type allowlistRepository struct {
	db *gorm.DB
}

func NewAllowlistRepository(db *gorm.DB) AllowlistRepository {
	return &allowlistRepository{db: db}
}

func (r *allowlistRepository) AddWallets(ctx context.Context, wallets []*models.CollectionAllowlist) error {
	if len(wallets) == 0 {
		return nil
	}

	// Use OnConflict to do nothing if wallet already exists for collection
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "collection_id"}, {Name: "wallet_address"}},
			DoUpdates: clause.AssignmentColumns([]string{"max_mint_amount", "added_at"}),
		}).
		Create(&wallets).Error
}

func (r *allowlistRepository) GetByCollection(ctx context.Context, collectionID uuid.UUID) ([]*models.CollectionAllowlist, error) {
	var allowlist []*models.CollectionAllowlist
	err := r.db.WithContext(ctx).
		Where("collection_id = ?", collectionID).
		Order("created_at DESC").
		Find(&allowlist).Error
	return allowlist, err
}

func (r *allowlistRepository) RemoveWallet(ctx context.Context, collectionID uuid.UUID, walletAddress string) error {
	return r.db.WithContext(ctx).
		Where("collection_id = ? AND wallet_address = ?", collectionID, walletAddress).
		Delete(&models.CollectionAllowlist{}).Error
}
