# Cache Invalidation for Webhook Events - Implementation Plan

## Overview

Implement cache-aside pattern for collection repository with cache invalidation in webhook handlers. When webhook events from the indexer update collection data, the cache will be invalidated to ensure subsequent reads return fresh data.

## Architecture

```
Indexer → Webhook Handler → Cached Repo → DB
                │                │
                └────────────────┘
                     Invalidate
```

## Design Decisions

1. **Cache-aside pattern**: Follow `CachedSessionRepository` from auth-service
2. **Cache keys**: `collection:id:{uuid}`, `collection:addr:{chain}:{address}`
3. **Non-blocking**: Log warnings, don't fail webhook processing
4. **TTL**: 5 minutes for collections
5. **Only cache single lookups**: GetByID, GetByContractAddress

## Phases

| Phase | Description | Status |
|-------|-------------|--------|
| 01 | Create CachedCollectionRepository | completed |
| 02 | Add cache invalidation to webhook handlers | completed |
| 03 | Wire dependencies in main.go and server | completed |
| 04 | Testing | completed |

## Key Files

- `services/collection-service/internal/repository/cached_collection_repository.go` (NEW)
- `services/collection-service/internal/server/webhook_handler.go` (MODIFY)
- `services/collection-service/internal/server/collection_server.go` (MODIFY)
- `services/collection-service/cmd/main.go` (MODIFY)

## Cache Invalidation Triggers

| Event | Data Modified | Cache Invalidated |
|-------|---------------|-------------------|
| collection.created | status, index_status | collection:id, collection:addr |
| collection.minted | total_minted | collection:id, collection:addr |
| collection.batch_minted | total_minted | collection:id, collection:addr |
