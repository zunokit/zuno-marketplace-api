package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
	sharedredis "github.com/zunokit/zuno-marketplace-api/shared/redis"
)

const (
	defaultCollectionTTL = 5 * time.Minute
	keyCollectionByID    = "collection:id:%s"
	keyCollectionByAddr  = "collection:addr:%s:%s"
)

// CachedCollectionRepository wraps CollectionRepository with Redis caching
type CachedCollectionRepository struct {
	repo  CollectionRepository
	cache *sharedredis.Cache
	ttl   time.Duration
}

// NewCachedCollectionRepository creates a new cached collection repository
func NewCachedCollectionRepository(repo CollectionRepository) CollectionRepository {
	return &CachedCollectionRepository{
		repo:  repo,
		cache: sharedredis.NewCache(),
		ttl:   defaultCollectionTTL,
	}
}

// GetByID retrieves a collection by ID with caching
func (r *CachedCollectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
	key := r.keyByID(id)

	// Try cache first
	var cached models.Collection
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return &cached, nil
	}

	// Cache miss - fetch from DB
	collection, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Populate cache
	if err := r.cache.Set(ctx, key, collection, r.ttl); err != nil {
		// Non-blocking: log cache error but don't fail the operation
	}

	return collection, nil
}

// GetByContractAddress retrieves a collection by contract address with caching
func (r *CachedCollectionRepository) GetByContractAddress(ctx context.Context, address, chainID string) (*models.Collection, error) {
	key := r.keyByAddress(address, chainID)

	// Try cache first
	var cached models.Collection
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return &cached, nil
	}

	// Cache miss - fetch from DB
	collection, err := r.repo.GetByContractAddress(ctx, address, chainID)
	if err != nil {
		return nil, err
	}

	// Populate cache
	if err := r.cache.Set(ctx, key, collection, r.ttl); err != nil {
		// Non-blocking: log cache error but don't fail the operation
	}

	return collection, nil
}

// InvalidateCollection removes a collection from cache
func (r *CachedCollectionRepository) InvalidateCollection(ctx context.Context, id uuid.UUID, address, chainID string) {
	keys := []string{r.keyByID(id)}
	if address != "" && chainID != "" {
		keys = append(keys, r.keyByAddress(address, chainID))
	}
	r.cache.Delete(ctx, keys...)
}

// keyByID generates cache key for collection ID
func (r *CachedCollectionRepository) keyByID(id uuid.UUID) string {
	return fmt.Sprintf(keyCollectionByID, id.String())
}

// keyByAddress generates cache key for contract address
func (r *CachedCollectionRepository) keyByAddress(address, chainID string) string {
	return fmt.Sprintf(keyCollectionByAddr, chainID, address)
}

// Pass-through methods (no caching)

// Create creates a new collection (no caching)
func (r *CachedCollectionRepository) Create(ctx context.Context, collection *models.Collection) error {
	return r.repo.Create(ctx, collection)
}

// Update updates a collection (no caching)
func (r *CachedCollectionRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return r.repo.Update(ctx, id, updates)
}

// ListByUser lists collections by user (no caching)
func (r *CachedCollectionRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.Collection, int64, error) {
	return r.repo.ListByUser(ctx, userID, page, limit)
}

// List lists collections with filters (no caching)
func (r *CachedCollectionRepository) List(ctx context.Context, filters *ListFilters, page, limit int) ([]*models.Collection, int64, error) {
	return r.repo.List(ctx, filters, page, limit)
}

// Delete deletes a collection (no caching)
func (r *CachedCollectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.repo.Delete(ctx, id)
}

// IncrementTotalMinted atomically increments total_minted (no caching)
func (r *CachedCollectionRepository) IncrementTotalMinted(ctx context.Context, id uuid.UUID, increment int64) error {
	return r.repo.IncrementTotalMinted(ctx, id, increment)
}
