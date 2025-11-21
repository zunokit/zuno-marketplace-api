package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/quangdang46/NFT-Marketplace/services/collection-service/internal/models"
	"gorm.io/gorm"
)

var (
	ErrCollectionNotFound = errors.New("collection not found")
	ErrSlugTaken          = errors.New("slug already taken")
)

type ListFilters struct {
	SortBy      string
	SortOrder   string
	Category    string
	ChainID     string
	IsVerified  *bool
	SearchQuery string
}

type CollectionRepository interface {
	Create(ctx context.Context, collection *models.Collection) error
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Collection, error)
	GetByContractAddress(ctx context.Context, address, chainID string) (*models.Collection, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.Collection, int64, error)
	List(ctx context.Context, filters *ListFilters, page, limit int) ([]*models.Collection, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type collectionRepository struct {
	db *gorm.DB
}

func NewCollectionRepository(db *gorm.DB) CollectionRepository {
	return &collectionRepository{db: db}
}

func (r *collectionRepository) Create(ctx context.Context, collection *models.Collection) error {
	return r.db.WithContext(ctx).Create(collection).Error
}

func (r *collectionRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Model(&models.Collection{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrCollectionNotFound
	}

	return nil
}

func (r *collectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
	var collection models.Collection
	err := r.db.WithContext(ctx).
		Preload("Metadata").
		Preload("Stats").
		First(&collection, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCollectionNotFound
		}
		return nil, err
	}

	return &collection, nil
}

func (r *collectionRepository) GetByContractAddress(ctx context.Context, address, chainID string) (*models.Collection, error) {
	var collection models.Collection
	err := r.db.WithContext(ctx).
		Preload("Metadata").
		Preload("Stats").
		Where("contract_address = ? AND chain_id = ?", address, chainID).
		First(&collection).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCollectionNotFound
		}
		return nil, err
	}

	return &collection, nil
}

func (r *collectionRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.Collection, int64, error) {
	var collections []*models.Collection
	var total int64

	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&models.Collection{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Metadata").
		Preload("Stats").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&collections).Error

	if err != nil {
		return nil, 0, err
	}

	return collections, total, nil
}

func (r *collectionRepository) List(ctx context.Context, filters *ListFilters, page, limit int) ([]*models.Collection, int64, error) {
	var collections []*models.Collection
	var total int64

	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&models.Collection{})

	// Apply filters
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}
	if filters.ChainID != "" {
		query = query.Where("chain_id = ?", filters.ChainID)
	}
	if filters.IsVerified != nil {
		query = query.Where("is_verified = ?", *filters.IsVerified)
	}
	if filters.SearchQuery != "" {
		query = query.Where("name ILIKE ? OR symbol ILIKE ?", "%"+filters.SearchQuery+"%", "%"+filters.SearchQuery+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	sortOrder := "DESC"
	if filters.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// Handle special sort fields that might be in joined tables (stats)
	if sortBy == "volume" || sortBy == "floor_price" {
		query = query.Joins("JOIN collection_stats ON collection_stats.collection_id = collections.id")
		if sortBy == "volume" {
			query = query.Order("collection_stats.total_volume_wei " + sortOrder)
		} else if sortBy == "floor_price" {
			query = query.Order("collection_stats.floor_price_wei " + sortOrder)
		}
	} else {
		query = query.Order(sortBy + " " + sortOrder)
	}

	err := query.
		Preload("Metadata").
		Preload("Stats").
		Offset(offset).
		Limit(limit).
		Find(&collections).Error

	if err != nil {
		return nil, 0, err
	}

	return collections, total, nil
}

func (r *collectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.Collection{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCollectionNotFound
	}
	return nil
}
