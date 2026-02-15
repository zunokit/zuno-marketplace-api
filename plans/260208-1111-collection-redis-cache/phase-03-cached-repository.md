# Phase 3: Cached Repository Implementation

## Overview
Implement the cache-aside wrapper repository following the auth-service pattern.

**Status:** Pending
**Blocked By:** Phase 2
**Blocks:** Phase 4, 5

---

## Files to Create

### `services/collection-service/internal/repository/cached_collection_repository.go`

```go
package repository

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/cache"
    "github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
    sharedredis "github.com/zunokit/zuno-marketplace-api/shared/redis"
)

// CachedCollectionRepository wraps CollectionRepository with Redis caching
type CachedCollectionRepository struct {
    repo  CollectionRepository
    cache *sharedredis.Cache
}

// NewCachedCollectionRepository creates a new cached repository
func NewCachedCollectionRepository(repo CollectionRepository, cache *sharedredis.Cache) CollectionRepository {
    return &CachedCollectionRepository{
        repo:  repo,
        cache: cache,
    }
}

// GetByID retrieves collection by ID with caching (no TTL - event-based)
func (r *CachedCollectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
    key := cache.CollectionKey(id)

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

    // Populate cache with safety TTL (event-based invalidation primary)
    if err := r.cache.Set(ctx, key, collection, time.Duration(cache.SafetyTTL)*time.Second); err != nil {
        // Non-blocking: log but don't fail
    }

    return collection, nil
}

// GetByContractAddress retrieves by contract with caching
func (r *CachedCollectionRepository) GetByContractAddress(ctx context.Context, address, chainID string) (*models.Collection, error) {
    // First try to get by contract key
    key := cache.CollectionByContractKey(chainID, address)

    var cached models.Collection
    if err := r.cache.Get(ctx, key, &cached); err == nil {
        return &cached, nil
    }

    // Cache miss
    collection, err := r.repo.GetByContractAddress(ctx, address, chainID)
    if err != nil {
        return nil, err
    }

    // Cache both keys for consistency
    if err := r.cache.Set(ctx, key, collection, time.Duration(cache.SafetyTTL)*time.Second); err != nil {
        // Non-blocking
    }

    // Also cache by ID for direct lookups
    idKey := cache.CollectionKey(collection.ID)
    if err := r.cache.Set(ctx, idKey, collection, time.Duration(cache.SafetyTTL)*time.Second); err != nil {
        // Non-blocking
    }

    return collection, nil
}

// ListByUser retrieves user's collections with caching (10 min TTL)
func (r *CachedCollectionRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.Collection, int64, error) {
    key := cache.CollectionsByUserKey(userID, page, limit)

    // Try cache
    var cached struct {
        Collections []*models.Collection
        Total       int64
    }
    if err := r.cache.Get(ctx, key, &cached); err == nil {
        return cached.Collections, cached.Total, nil
    }

    // Cache miss
    collections, total, err := r.repo.ListByUser(ctx, userID, page, limit)
    if err != nil {
        return nil, 0, err
    }

    // Populate cache
    cached.Collections = collections
    cached.Total = total
    if err := r.cache.Set(ctx, key, cached, time.Duration(cache.ListTTL)*time.Second); err != nil {
        // Non-blocking
    }

    return collections, total, nil
}

// List retrieves filtered collections with caching (10 min TTL)
func (r *CachedCollectionRepository) List(ctx context.Context, filters *ListFilters, page, limit int) ([]*models.Collection, int64, error) {
    key := cache.CollectionsListKey(filters, page, limit)

    // Try cache
    var cached struct {
        Collections []*models.Collection
        Total       int64
    }
    if err := r.cache.Get(ctx, key, &cached); err == nil {
        return cached.Collections, cached.Total, nil
    }

    // Cache miss
    collections, total, err := r.repo.List(ctx, filters, page, limit)
    if err != nil {
        return nil, 0, err
    }

    // Populate cache
    cached.Collections = collections
    cached.Total = total
    if err := r.cache.Set(ctx, key, cached, time.Duration(cache.ListTTL)*time.Second); err != nil {
        // Non-blocking
    }

    return collections, total, nil
}

// Create creates collection and invalidates relevant caches
func (r *CachedCollectionRepository) Create(ctx context.Context, collection *models.Collection) error {
    if err := r.repo.Create(ctx, collection); err != nil {
        return err
    }

    // Invalidate user's list caches (new collection affects user's lists)
    // Note: We can't know all page combinations, so rely on TTL for lists
    // Or implement pattern-based deletion if needed

    return nil
}

// Update updates collection and invalidates cache
func (r *CachedCollectionRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
    if err := r.repo.Update(ctx, id, updates); err != nil {
        return err
    }

    // Invalidate collection caches
    keys := []string{
        cache.CollectionKey(id),
        cache.CollectionStatsKey(id),
    }

    // Also invalidate by contract if address in updates
    if contractAddr, ok := updates["contract_address"].(string); ok && contractAddr != "" {
        if chainID, ok := updates["chain_id"].(string); ok && chainID != "" {
            keys = append(keys, cache.CollectionByContractKey(chainID, contractAddr))
        }
    }

    if err := r.cache.Delete(ctx, keys...); err != nil {
        // Non-blocking: log but don't fail
    }

    // Note: List caches use TTL-based invalidation for simplicity

    return nil
}

// Delete deletes collection and invalidates cache
func (r *CachedCollectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
    // Get collection first to find contract key for invalidation
    collection, err := r.repo.GetByID(ctx, id)
    if err != nil {
        return err
    }

    if err := r.repo.Delete(ctx, id); err != nil {
        return err
    }

    // Invalidate all related caches
    keys := []string{
        cache.CollectionKey(id),
        cache.CollectionStatsKey(id),
    }

    if collection.ContractAddress != nil && *collection.ContractAddress != "" {
        keys = append(keys, cache.CollectionByContractKey(collection.ChainID, *collection.ContractAddress))
    }

    if err := r.cache.Delete(ctx, keys...); err != nil {
        // Non-blocking
    }

    return nil
}

// IncrementTotalMinted increments counter and invalidates stats cache
func (r *CachedCollectionRepository) IncrementTotalMinted(ctx context.Context, id uuid.UUID, increment int64) error {
    if err := r.repo.IncrementTotalMinted(ctx, id, increment); err != nil {
        return err
    }

    // Invalidate stats cache (will be refreshed on next read)
    key := cache.CollectionStatsKey(id)
    if err := r.cache.Delete(ctx, key); err != nil {
        // Non-blocking
    }

    return nil
}
```

---

## Implementation Steps

1. **Create cached_collection_repository.go**
   - Implement all CollectionRepository interface methods
   - Use cache-aside pattern for reads
   - Invalidate on writes
   - Non-blocking cache errors

2. **Handle edge cases**
   - Contract address lookups cache both keys
   - Delete fetches collection first for contract key
   - Updates invalidate both ID and contract keys

3. **Add logging for cache operations**
   - Log cache hits/misses at debug level
   - Log cache errors at warn level (non-blocking)

---

## Testing Checklist

- [ ] GetByID returns cached data on hit
- [ ] GetByID populates cache on miss
- [ ] Update invalidates collection cache
- [ ] Delete invalidates all related keys
- [ ] List queries cache with TTL
- [ ] Cache errors don't fail operations

---

## Next Phase

[Phase 4: Stats Separation](./phase-04-stats-separation.md)
