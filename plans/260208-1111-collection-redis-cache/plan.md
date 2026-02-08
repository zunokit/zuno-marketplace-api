# Plan: Collection Redis Cache Implementation

## Overview

Implement Redis caching for the collection service using cache-aside pattern with event-based invalidation. This plan follows the brainstorm report at `plans/reports/brainstorm-260208-1102-collection-redis-cache.md`.

**Status:** Pending
**Priority:** High
**Estimated Effort:** 2-3 days

---

## Architecture

```
┌─────────────────┐     ┌──────────────────────┐     ┌─────────────┐
│  gRPC Server    │────▶│ CachedCollectionRepo │────▶│    Redis    │
│                 │     │   (new wrapper)      │     │   Cache     │
└─────────────────┘     └──────────┬───────────┘     └─────────────┘
                                   │
                    Cache miss ────┘
                                   │
                                   ▼
                          ┌──────────────────┐
                          │ CollectionRepo   │
                          │   (existing)     │
                          └────────┬─────────┘
                                   │
                                   ▼
                          ┌──────────────────┐
                          │   PostgreSQL     │
                          └──────────────────┘
```

---

## Phases

| Phase | Name | Status | Description |
|-------|------|--------|-------------|
| 1 | [Setup Redis Configuration](./phase-01-setup-redis.md) | **Completed** | Add Redis config and initialization |
| 2 | [Cache Key Utilities](./phase-02-cache-keys.md) | **Completed** | Create cache key builders |
| 3 | [Cached Repository](./phase-03-cached-repository.md) | **Completed** | Implement cache-aside wrapper |
| 4 | [Stats Separation](./phase-04-stats-separation.md) | **Completed** | Separate stats caching |
| 5 | [Wire Up Main](./phase-05-wire-up-main.md) | **Completed** | Integrate into service |
| 6 | [Testing](./phase-06-testing.md) | **Completed** | Unit and integration tests |

---

## Key Decisions

| Decision | Value | Rationale |
|----------|-------|-----------|
| Pattern | Cache-aside decorator | Follows auth-service pattern, non-invasive |
| Single item TTL | Event-based (no TTL) | Invalidated on update/delete |
| Stats TTL | 5 minutes | Frequent changes during trading |
| List TTL | 10 minutes | Balance freshness/performance |
| List invalidation | TTL-based (MVP) | Simple, add tag-based later if needed |
| Cache errors | Non-blocking | Service continues if Redis fails |
| Redis Mode | Dual (Upstash + Docker) | Upstash for dev, Docker for production |

---

## Cache Key Patterns

| Data Type | Pattern | Example |
|-----------|---------|---------|
| By ID | `collection:{id}` | `collection:550e8400...` |
| By contract | `collection:contract:{chain}:{addr}` | `collection:contract:1:0x1234...` |
| Stats | `collection:{id}:stats` | `collection:550e8400:stats` |
| User lists | `collections:user:{id}:p{p}:l{l}` | `collections:user:550e...:p1:l20` |
| Filtered lists | `collections:list:{hash}` | `collections:list:a3f7...` |

---

## Files to Create/Modify

### New Files
- `services/collection-service/internal/repository/cached_collection_repository.go`
- `services/collection-service/internal/cache/cache_keys.go`
- `services/collection-service/internal/cache/invalidator.go`

### Modified Files
- `services/collection-service/cmd/main.go`
- `services/collection-service/internal/config/config.go`

---

## Success Criteria

- [ ] Cache hit rate > 80% for GetCollection calls
- [ ] P95 latency reduced by 50% for cached endpoints
- [ ] All existing tests pass
- [ ] New cache tests cover hit/miss/invalidation scenarios
- [ ] Redis failures don't break the service

---

## Risks

| Risk | Mitigation |
|------|------------|
| Stale data | Event-based invalidation; short TTL for stats |
| Redis failure | Non-blocking errors; fallback to DB |
| Memory exhaustion | Redis LRU; 24h safety TTL |

---

## Open Questions

1. Need cache metrics for Prometheus?
2. Need cache bypass header for debugging?
3. Need singleflight for cache stampede protection?

---

## Next Steps

Start with [Phase 1: Setup Redis Configuration](./phase-01-setup-redis.md)
