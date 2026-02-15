package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
	"gorm.io/gorm"
)

type MetadataRepository interface {
	Update(ctx context.Context, collectionID uuid.UUID, updates map[string]interface{}) error
	GetByCollectionID(ctx context.Context, collectionID uuid.UUID) (*models.CollectionMetadata, error)
}

type metadataRepository struct {
	db *gorm.DB
}

func NewMetadataRepository(db *gorm.DB) MetadataRepository {
	return &metadataRepository{db: db}
}

func (r *metadataRepository) Update(ctx context.Context, collectionID uuid.UUID, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Model(&models.CollectionMetadata{}).
		Where("collection_id = ?", collectionID).
		Updates(updates)

	return result.Error
}

func (r *metadataRepository) GetByCollectionID(ctx context.Context, collectionID uuid.UUID) (*models.CollectionMetadata, error) {
	var metadata models.CollectionMetadata
	err := r.db.WithContext(ctx).
		Where("collection_id = ?", collectionID).
		First(&metadata).Error

	if err != nil {
		return nil, err
	}

	return &metadata, nil
}
