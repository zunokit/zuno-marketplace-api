# Brainstorm Report: Collection Redis Cache Implementation

## Problem Statement

The collection service currently has **no caching layer**. Every request hits PostgreSQL directly, causing:
- Unnecessary database load for frequently accessed collections
- Slower response times for browse pages (ListCollections)
- No protection against traffic spikes on popular collection pages
- Repeated expensive queries for collection stats that change frequently

## Requirements Summary

Based on discussion:
1. **Event-based invalidation** - Cache until explicitly invalidated on updates
2. **List query caching** - Cache ListCollections and ListCollectionsByUser with filters
3. **Separate stats cache** - Collection stats cached separately with shorter TTL

## Evaluated Approaches

### Approach 1: Cache-Aside with Decorator Pattern (RECOMMENDED)

Create a `CachedCollectionRepository` that wraps the existing repository, similar to `CachedSessionRepository` in auth-service.

**Pros:**
- Non-invasive: no changes to existing repository or service code
- Follows existing codebase patterns (DRY principle)
- Easy to disable caching by swapping repository implementation
- Cache errors are non-blocking (operation continues if Redis fails)
- Simple to test - can mock cache layer

**Cons:**
- Requires careful cache invalidation on all mutation paths
- List cache invalidation is complex (many filter combinations)

**Implementation:**
```go
type CachedCollectionRepository struct {
    repo  CollectionRepository
    cache *sharedredis.Cache
}

// Cache-aside pattern for GetByID
func (r *CachedCollectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
    key := fmt.Sprintf("collection:%s", id)

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

    // Populate cache (no TTL - event-based invalidation)
    r.cache.Set(ctx, key, collection, 0)

    return collection, nil
}
```

### Approach 2: Service-Level Caching

Add caching logic directly in `CollectionService` methods.

**Pros:**
- Can cache at the business logic level (e.g., computed values)
- Single place for cache logic per operation

**Cons:**
- Violates SRP - service now handles business logic AND caching
- Harder to test - cache logic mixed with business logic
- More invasive changes required
- Not consistent with existing auth-service pattern

**Verdict:** Rejected - breaks separation of concerns.

### Approach 3: gRPC Interceptor Caching

Implement caching at the gRPC server level using interceptors.

**Pros:**
- Transparent to service and repository layers
- Can cache raw protobuf responses
- Centralized cache control

**Cons:**
- Cannot leverage existing Redis client in shared/redis
- Harder to implement fine-grained invalidation
- Over-caching risk (caches error responses unless careful)
- More complex to debug cache issues

**Verdict:** Rejected - too complex for current needs, harder to maintain.

## Final Recommended Solution

### Architecture Overview

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────┐
│  gRPC Server    │────▶│ CachedCollection │────▶│    Redis    │
│                 │     │   Repository     │     │   Cache     │
└─────────────────┘     └────────┬─────────┘     └─────────────┘
                                 │
                                 ▼ (cache miss)
                        ┌──────────────────┐
                        │  Collection      │
                        │  Repository      │
                        │  (existing)      │
                        └────────┬─────────┘
                                 │
                                 ▼
                        ┌──────────────────┐
                        │   PostgreSQL     │
                        └──────────────────┘
```

### Cache Key Strategy

| Data Type | Cache Key Pattern | Example |
|-----------|------------------|---------|
| Collection by ID | `collection:{id}` | `collection:550e8400-e29b-41d4-a716-446655440000` |
| Collection by contract | `collection:contract:{chain}:{address}` | `collection:contract:1:0x1234...` |
| Collection stats | `collection:{id}:stats` | `collection:550e8400:stats` |
| List by user | `collections:user:{user_id}:p{page}:l{limit}` | `collections:user:550e...:p1:l20` |
| Filtered list | `collections:list:{hash}` | `collections:list:a3f7...` |

### TTL Strategy

| Data Type | TTL | Rationale |
|-----------|-----|-----------|
| Collection metadata | No TTL (event-based) | Stable data, invalidated on update |
| Collection stats | 5 minutes | Changes frequently during trading |
| List queries | 10 minutes | Balance between freshness and performance |

### Invalidation Strategy

**On Collection Update:**
```go
func (r *CachedCollectionRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
    // 1. Update in DB
    if err := r.repo.Update(ctx, id, updates); err != nil {
        return err
    }

    // 2. Invalidate collection cache
    r.cache.Delete(ctx,
        fmt.Sprintf("collection:%s", id),
        fmt.Sprintf("collection:%s:stats", id),
    )

    // 3. Invalidate list caches (pattern-based or tags)
    // Option A: Use Redis SCAN to find and delete list keys
    // Option B: Use Redis tags (requires RedisJSON or custom implementation)
    // Option C: Accept stale lists for TTL duration (simplest)

    return nil
}
```

**List Cache Invalidation Options:**

1. **TTL-based (Recommended for MVP):** Lists expire naturally after 10 min. Simple, no complex invalidation logic.

2. **Tag-based:** Use Redis sets to track which list keys contain a collection. On update, find and invalidate all tagged lists. More complex but immediate consistency.

3. **Version-based:** Include a version number in list cache keys. On any collection update, increment global version (all lists invalidated). Nuclear option but simple.

**Recommendation:** Start with TTL-based for lists, event-based for single items. Add tag-based invalidation later if stale lists become problematic.

### Implementation Files

**New Files:**
1. `services/collection-service/internal/repository/cached_collection_repository.go` - Cache wrapper
2. `services/collection-service/internal/cache/cache_keys.go` - Cache key builders
3. `services/collection-service/internal/cache/invalidator.go` - Cache invalidation helpers

**Modified Files:**
1. `services/collection-service/cmd/main.go` - Wire up cached repository
2. `services/collection-service/internal/config/config.go` - Add Redis config

### Code Structure

```go
// cached_collection_repository.go
package repository

type CachedCollectionRepository struct {
    repo  CollectionRepository
    cache *sharedredis.Cache
}

func NewCachedCollectionRepository(repo CollectionRepository, cache *sharedredis.Cache) CollectionRepository {
    return &CachedCollectionRepository{repo: repo, cache: cache}
}

// Implement all CollectionRepository methods with caching
// - GetByID, GetByContractAddress: cache-aside, no TTL
// - List, ListByUser: cache-aside, 10 min TTL
// - Create, Update, Delete: write-through + invalidate
// - Stats methods: separate 5 min TTL
```

### Stats Separation Strategy

Since stats change frequently but metadata is stable:

```go
// In service layer or repository
type CollectionWithStats struct {
    Collection *models.Collection
    Stats      *models.CollectionStats
}

func (r *CachedCollectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
    // Fetch collection (cached, no TTL)
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
```

## Implementation Considerations

### Race Conditions
- **Cache stampede:** Multiple requests for expired cache key hit DB simultaneously.
  - **Mitigation:** Use singleflight pattern or acceptable for MVP (low traffic)

### Memory Usage
- **Redis memory:** Unlimited growth if no TTL on collections.
  - **Mitigation:** Add 24h TTL as safety net, or implement LRU eviction policy in Redis config

### Cache Warming
- New deployments start with cold cache.
  - **Mitigation:** Not critical for MVP, consider background warming for popular collections later

### Monitoring
- Cache hit/miss rates
- Redis memory usage
- Cache invalidation events
- Consider adding Prometheus metrics

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Stale data served | Medium | Event-based invalidation on all mutations; short TTL for stats |
| Redis failure | Low | Non-blocking cache errors; fallback to DB always works |
| Memory exhaustion | Medium | Redis LRU eviction; monitor memory usage |
| Complex invalidation | Medium | Start with TTL for lists; add tag-based if needed |
| Cache stampede | Low | Acceptable for current scale; add singleflight if needed |

## Success Metrics

1. **Cache hit rate > 80%** for GetCollection calls
2. **P95 latency reduction** by 50% for cached endpoints
3. **Database CPU** reduction by 30% during normal load
4. **Zero cache-related errors** in production (non-blocking design)

## Next Steps

1. **Create implementation plan** with detailed TODOs
2. **Implement CachedCollectionRepository** following auth-service pattern
3. **Add Redis configuration** to collection service
4. **Write tests** for cache hit/miss scenarios
5. **Add metrics** for cache performance monitoring
6. **Deploy and monitor** cache hit rates

## Open Questions

1. Should we implement cache warming for featured/popular collections on startup?
2. Do we need cache metrics exposed via Prometheus for monitoring?
3. Should we add a cache bypass header for debugging (e.g., `X-Cache-Bypass: true`)?
4. What's the expected peak traffic - do we need singleflight for cache stampede protection?
