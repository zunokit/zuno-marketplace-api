# Cache Invalidation Strategy for Webhook Events

## Problem Statement

When webhook events from the indexer update collection data (status, total_minted, etc.), the system currently has **no cache invalidation logic**. This causes:

1. **Stale data in GraphQL queries** - Users see outdated collection stats
2. **Inconsistent state** - Database updated but cache still holds old values
3. **Poor UX** - Mint counts don't reflect actual blockchain state

## Current Architecture

```
Indexer → Webhook Handler (gateway) → Collection Service (gRPC) → PostgreSQL
                                              ↓
                                       No cache invalidation!
                                              ↓
GraphQL Gateway ← Collection Service ← (stale data if cached)
```

### Webhook Events That Modify Data

| Event | Data Modified | Cache Keys Affected |
|-------|---------------|---------------------|
| `collection.created` | status, index_status, indexed_at | `collection:{id}`, `collection:contract:{address}` |
| `collection.minted` | total_minted | `collection:{id}`, `collection:contract:{address}`, list caches |
| `collection.batch_minted` | total_minted | `collection:{id}`, `collection:contract:{address}`, list caches |

## Evaluated Approaches

### Approach 1: Direct Cache Deletion in Webhook Handler (Simplest)

**Implementation:**
- Add Redis cache client to `CollectionServer`
- Delete cache keys immediately after successful DB update in webhook handler

**Pros:**
- Simple to implement and understand
- Immediate consistency
- No additional infrastructure

**Cons:**
- Cache logic mixed with business logic
- Hard to maintain if cache keys change
- No retry mechanism if cache deletion fails

**Code Pattern:**
```go
// In handleCollectionMinted, after DB update:
if err := s.cache.Delete(ctx,
    fmt.Sprintf("collection:%s", dbCollection.ID),
    fmt.Sprintf("collection:contract:%s:%s", chainID, normalizedAddress),
); err != nil {
    s.logger.Warn("Failed to invalidate cache", zap.Error(err))
    // Non-blocking: don't fail the webhook
}
```

---

### Approach 2: Cache-Aside Repository Wrapper (Recommended)

**Implementation:**
- Create `CachedCollectionRepository` similar to `CachedSessionRepository`
- Wrap all DB operations with cache invalidation
- Use cache-aside pattern (check cache → DB → populate cache)

**Pros:**
- Clean separation of concerns
- Reusable pattern (already used in auth-service)
- Centralized cache key management
- Consistent with existing codebase

**Cons:**
- More boilerplate code
- Slightly more complex

**Code Pattern:**
```go
type CachedCollectionRepository struct {
    repo  CollectionRepository
    cache *redis.Cache
}

func (r *CachedCollectionRepository) IncrementTotalMinted(ctx context.Context, id uuid.UUID, increment int64) error {
    // Update DB first
    if err := r.repo.IncrementTotalMinted(ctx, id, increment); err != nil {
        return err
    }
    // Invalidate cache
    return r.invalidateCollectionCache(ctx, id)
}
```

---

### Approach 3: Event-Driven Cache Invalidation (Most Scalable)

**Implementation:**
- Publish cache invalidation events to RabbitMQ
- Separate consumer handles cache deletion
- Supports cross-service cache invalidation

**Pros:**
- Decoupled from webhook processing
- Can batch invalidations
- Supports multiple consumers
- Better for high-throughput scenarios

**Cons:**
- Adds latency (eventual consistency)
- More infrastructure complexity
- Overkill for current scale

**Flow:**
```
Webhook Handler → DB Update → Publish "cache.invalidate" event
                                        ↓
                              RabbitMQ Queue
                                        ↓
                              Cache Invalidation Consumer → Redis Delete
```

---

### Approach 4: TTL-Based Cache with Short Expiry (Simplest but Least Consistent)

**Implementation:**
- Set short TTL (e.g., 30-60 seconds) on collection cache
- No explicit invalidation needed

**Pros:**
- Zero code changes to webhook handler
- Simplest implementation

**Cons:**
- Users see stale data for TTL duration
- Not true consistency
- Higher DB load from cache misses

**Not recommended** for financial/NFT data where accuracy matters.

---

## Recommended Solution: Hybrid Approach (Approach 2 + Selective TTL)

### Phase 1: Cache-Aside Repository Wrapper

Implement `CachedCollectionRepository` with these cache keys:

```
collection:{id}                    # Single collection by ID
collection:contract:{chain}:{addr} # Single collection by contract
collections:user:{userID}:{page}   # User's collections (paginated)
collections:list:{hash}            # Public list with filters
```

### Phase 2: Webhook Handler Integration

Modify webhook handlers to invalidate specific keys:

```go
func (s *CollectionServer) handleCollectionMinted(...) {
    // ... existing DB update logic ...

    // Invalidate specific collection caches
    s.cache.Delete(ctx,
        fmt.Sprintf("collection:%s", dbCollection.ID),
        fmt.Sprintf("collection:contract:%s:%s", chainID, normalizedAddress),
    )

    // Invalidate list caches (pattern-based or tracked)
    s.invalidateListCaches(ctx, dbCollection.ID)
}
```

### Phase 3: List Cache Strategy

**Option A: Pattern-Based Invalidation**
- Use Redis SCAN to find and delete matching keys
- Simpler but slower on large datasets

**Option B: Tag-Based Invalidation**
- Add tags to cached entries
- Invalidate by tag
- Requires more complex cache implementation

**Option C: Skip List Caching (Recommended for MVP)**
- Only cache single collection lookups
- List queries always hit DB
- Simpler, no list invalidation complexity

## Implementation Plan

### Step 1: Create CachedCollectionRepository

```go
// services/collection-service/internal/repository/cached_collection_repository.go

type CachedCollectionRepository struct {
    repo  CollectionRepository
    cache *redis.Cache
    ttl   time.Duration
}

func NewCachedCollectionRepository(repo CollectionRepository, cache *redis.Cache) CollectionRepository {
    return &CachedCollectionRepository{
        repo: repo,
        cache: cache,
        ttl:   5 * time.Minute,
    }
}
```

### Step 2: Update CollectionServer

```go
type CollectionServer struct {
    pb.UnimplementedCollectionServiceServer
    service            *service.CollectionService
    processedEventRepo repository.ProcessedEventRepository
    cache              *redis.Cache  // Add this
    logger             *zap.Logger
}
```

### Step 3: Add Cache Invalidation to Webhook Handlers

After each successful DB operation:

```go
// Invalidate collection cache
keys := []string{
    fmt.Sprintf("collection:%s", dbCollection.ID),
    fmt.Sprintf("collection:contract:%s:%s", chainID, normalizedAddress),
}
if err := s.cache.Delete(ctx, keys...); err != nil {
    s.logger.Warn("Cache invalidation failed", zap.Error(err))
    // Non-blocking - don't fail webhook
}
```

## Cache Key Design

```go
// Key patterns
const (
    KeyCollectionByID       = "collection:%s"
    KeyCollectionByContract = "collection:contract:%s:%s" // chain:address
    KeyUserCollections      = "collections:user:%s:%d"    // userID:page
    KeyListCollections      = "collections:list:%s"       // filter hash
)

// Helper functions
func CollectionKey(id uuid.UUID) string {
    return fmt.Sprintf(KeyCollectionByID, id.String())
}

func CollectionContractKey(chainID, address string) string {
    return fmt.Sprintf(KeyCollectionByContract, chainID, address)
}
```

## Success Criteria

1. **Immediate Consistency:** After webhook processing, subsequent reads return updated data
2. **Non-blocking:** Cache failures don't fail webhook processing
3. **Observability:** Cache invalidation events are logged
4. **Test Coverage:** Unit tests verify cache invalidation behavior

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Cache deletion fails | Medium | Log warning, don't fail webhook, rely on TTL fallback |
| Race condition (cache repopulated with stale data) | Low | Use cache-aside pattern, delete after DB commit |
| Redis unavailable | Low | Circuit breaker pattern, fallback to DB |
| Memory pressure from cache | Low | Set appropriate TTLs, monitor Redis memory |

## Next Steps

1. **Decision:** Choose approach (recommend Approach 2)
2. **Implementation:** Create cached repository layer
3. **Integration:** Add cache invalidation to webhook handlers
4. **Testing:** Unit + integration tests
5. **Monitoring:** Add metrics for cache hit/miss/invalidation rates

## Unresolved Questions

1. Should we cache list queries or only single collection lookups?
2. What TTL values are appropriate? (suggest: 5 min for collections, 1 min for lists)
3. Do we need distributed cache invalidation across multiple gateway instances?
