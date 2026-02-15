# Phase 5: Wire Up Main

## Overview
Integrate the cached repository into the service initialization.

**Status:** Pending
**Blocked By:** Phase 3, 4
**Blocks:** Phase 6

---

## Files to Modify

### `services/collection-service/cmd/main.go`

Current initialization:
```go
collectionRepo := repository.NewCollectionRepository(db)
collectionService := service.NewCollectionService(
    collectionRepo,
    allowlistRepo,
    metadataRepo,
)
```

New initialization with caching:
```go
import (
    // ... existing imports ...
    sharedredis "github.com/zunokit/zuno-marketplace-api/shared/redis"
)

func main() {
    // ... existing setup ...

    // Initialize Redis cache
    cache := sharedredis.NewCache()

    // Create base repository
    baseRepo := repository.NewCollectionRepository(db)

    // Wrap with caching
    cachedRepo := repository.NewCachedCollectionRepository(baseRepo, cache)

    // Use cached repository in service
    collectionService := service.NewCollectionService(
        cachedRepo,
        allowlistRepo,
        metadataRepo,
    )

    // ... rest of initialization ...
}
```

---

## Implementation Steps

1. **Import shared/redis package**
   ```go
   sharedredis "github.com/zunokit/zuno-marketplace-api/shared/redis"
   ```

2. **Initialize Redis client**
   ```go
   cache := sharedredis.NewCache()
   ```

3. **Wrap repository with cache**
   ```go
   baseRepo := repository.NewCollectionRepository(db)
   cachedRepo := repository.NewCachedCollectionRepository(baseRepo, cache)
   ```

4. **Pass to service**
   ```go
   collectionService := service.NewCollectionService(
       cachedRepo,  // Use cached version
       allowlistRepo,
       metadataRepo,
   )
   ```

---

## Verification

After wiring up:

1. **Service starts successfully**
   - No compilation errors
   - Redis connection established (or graceful fallback)

2. **Cache is active**
   - First request populates cache
   - Subsequent requests hit cache

3. **Invalidation works**
   - Update collection → cache invalidated
   - Next request fetches fresh data

---

## Testing Checklist

- [ ] Service compiles and starts
- [ ] Redis connection logged on startup
- [ ] Cache operations logged (debug level)
- [ ] First request is cache miss
- [ ] Second request is cache hit
- [ ] Update invalidates cache

---

## Next Phase

[Phase 6: Testing](./phase-06-testing.md)
