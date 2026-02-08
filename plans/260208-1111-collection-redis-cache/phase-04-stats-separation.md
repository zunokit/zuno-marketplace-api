# Phase 4: Stats Separation

## Overview
Separate collection stats caching with shorter TTL since stats change frequently.

**Status:** Pending
**Blocked By:** Phase 3
**Blocks:** Phase 5

---

## Problem

Collection stats (floor price, volume, listed count) change frequently during trading, but metadata (name, description, image) is stable. Caching them together causes:
- Either serving stale stats (if long TTL)
- Or unnecessary DB hits for metadata (if short TTL)

## Solution

Fetch and cache stats separately with 5-minute TTL, while metadata uses event-based invalidation.

---

## Implementation

### Option A: Repository-Level Separation (Recommended)

Modify `GetByID` to fetch stats separately:

```go
func (r *CachedCollectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
    // Fetch collection metadata (cached, event-based)
    collection, err := r.getCollectionFromCache(ctx, id)
    if err != nil {
        return nil, err
    }

    // Fetch stats separately (cached, 5 min TTL)
    stats, err := r.getStatsFromCache(ctx, id)
    if err != nil {
        // Stats non-critical, continue without
        stats = nil
    }

    collection.Stats = stats
    return collection, nil
}

func (r *CachedCollectionRepository) getStatsFromCache(ctx context.Context, id uuid.UUID) (*models.CollectionStats, error) {
    key := cache.CollectionStatsKey(id)

    // Try cache
    var cached models.CollectionStats
    if err := r.cache.Get(ctx, key, &cached); err == nil {
        return &cached, nil
    }

    // Cache miss - fetch from DB
    // Note: This requires a new method in base repo or use GetByID and extract
    collection, err := r.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if collection.Stats == nil {
        return nil, nil
    }

    // Cache with 5 min TTL
    if err := r.cache.Set(ctx, key, collection.Stats, time.Duration(cache.StatsTTL)*time.Second); err != nil {
        // Non-blocking
    }

    return collection.Stats, nil
}
```

### Option B: Separate Stats Repository

Create a dedicated stats repository if stats access pattern differs significantly:

```go
type CollectionStatsRepository interface {
    GetByCollectionID(ctx context.Context, collectionID uuid.UUID) (*models.CollectionStats, error)
    UpdateStats(ctx context.Context, collectionID uuid.UUID, updates map[string]interface{}) error
}
```

**Verdict:** Option A is simpler and sufficient for current needs.

---

## Files to Modify

### `cached_collection_repository.go`

Add private helper methods:

```go
// getCollectionFromCache fetches collection without stats (metadata only)
func (r *CachedCollectionRepository) getCollectionFromCache(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
    key := cache.CollectionKey(id)

    var cached models.Collection
    if err := r.cache.Get(ctx, key, &cached); err == nil {
        return &cached, nil
    }

    collection, err := r.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Store without stats for cleaner separation
    collectionWithoutStats := *collection
    collectionWithoutStats.Stats = nil

    if err := r.cache.Set(ctx, key, collectionWithoutStats, time.Duration(cache.SafetyTTL)*time.Second); err != nil {
        // Non-blocking
    }

    return collection, nil
}

// getStatsFromCache fetches only stats with short TTL
func (r *CachedCollectionRepository) getStatsFromCache(ctx context.Context, id uuid.UUID) (*models.CollectionStats, error) {
    key := cache.CollectionStatsKey(id)

    var cached models.CollectionStats
    if err := r.cache.Get(ctx, key, &cached); err == nil {
        return &cached, nil
    }

    // Fetch from DB
    collection, err := r.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if collection.Stats == nil {
        return nil, nil
    }

    // Cache with 5 min TTL
    if err := r.cache.Set(ctx, key, *collection.Stats, time.Duration(cache.StatsTTL)*time.Second); err != nil {
        // Non-blocking
    }

    return collection.Stats, nil
}
```

---

## Implementation Steps

1. **Modify GetByID** to use helper methods
2. **Store metadata without stats** in collection cache
3. **Cache stats separately** with 5 min TTL
4. **Invalidate both** on collection update

---

## Testing Checklist

- [ ] Stats cached separately from metadata
- [ ] Stats TTL is 5 minutes
- [ ] Metadata TTL is event-based (safety 24h)
- [ ] Update invalidates both caches
- [ ] Stats cache miss doesn't fail request

---

## Next Phase

[Phase 5: Wire Up Main](./phase-05-wire-up-main.md)
