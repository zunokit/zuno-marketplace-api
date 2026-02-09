# Phase 04: Testing

## Overview
Write tests for cached repository and webhook cache invalidation.

## Test Files to Create/Modify

### 1. cached_collection_repository_test.go (NEW)

Test cases:
- GetByID cache hit
- GetByID cache miss populates cache
- GetByContractAddress cache hit
- GetByContractAddress cache miss populates cache
- InvalidateCollection deletes both keys
- Non-blocking behavior on cache errors

### 2. webhook_handler_test.go (MODIFY)

Add test cases:
- handleCollectionCreated invalidates cache
- handleCollectionMinted invalidates cache
- handleCollectionBatchMinted invalidates cache

## Success Criteria
- [ ] Unit tests for CachedCollectionRepository
- [ ] Integration tests for webhook cache invalidation
- [ ] All tests pass
- [ ] Coverage for cache hit, miss, and invalidation scenarios
