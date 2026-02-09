# Phase 02: Add Cache Invalidation to Webhook Handlers

## Overview
Modify webhook handlers to invalidate cache after successful DB updates.

## File to Modify
`services/collection-service/internal/server/webhook_handler.go`

## Changes Required

### 1. Add cache invalidation in handleCollectionCreated

After successful collection update (around line 207), add:

```go
// Invalidate cache after successful update
if s.cacheRepo != nil {
	s.cacheRepo.InvalidateCollection(ctx, dbCollection.ID, normalizedAddress, chainID)
}
```

### 2. Add cache invalidation in handleCollectionMinted

After successful increment (around line 260), add:

```go
// Invalidate cache after successful increment
if s.cacheRepo != nil {
	s.cacheRepo.InvalidateCollection(ctx, dbCollection.ID, normalizedAddress, chainID)
}
```

### 3. Add cache invalidation in handleCollectionBatchMinted

After successful increment (around line 316), add:

```go
// Invalidate cache after successful increment
if s.cacheRepo != nil {
	s.cacheRepo.InvalidateCollection(ctx, dbCollection.ID, normalizedAddress, chainID)
}
```

## Implementation Notes

- Check `s.cacheRepo != nil` for safety during transition
- Non-blocking: don't fail webhook if cache invalidation fails
- Use normalized address and chainID from handler context

## Success Criteria
- [ ] All three webhook handlers invalidate cache after DB updates
- [ ] Cache invalidation is non-blocking
- [ ] Uses normalized address and chainID
