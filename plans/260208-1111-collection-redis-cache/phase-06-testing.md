# Phase 6: Testing

## Overview
Write comprehensive tests for cache functionality.

**Status:** Pending
**Blocked By:** Phase 5
**Blocks:** None

---

## Test Files to Create

### 1. `services/collection-service/internal/cache/cache_keys_test.go`

```go
package cache

import (
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/repository"
)

func TestCollectionKey(t *testing.T) {
    id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
    key := CollectionKey(id)
    assert.Equal(t, "collection:550e8400-e29b-41d4-a716-446655440000", key)
}

func TestCollectionByContractKey(t *testing.T) {
    key := CollectionByContractKey("1", "0x1234567890abcdef")
    assert.Equal(t, "collection:contract:1:0x1234567890abcdef", key)
}

func TestCollectionsListKey_Deterministic(t *testing.T) {
    filters := &repository.ListFilters{
        SortBy:      "created_at",
        SortOrder:   "desc",
        Category:    "art",
        ChainID:     "1",
        SearchQuery: "test",
    }

    key1 := CollectionsListKey(filters, 1, 20)
    key2 := CollectionsListKey(filters, 1, 20)

    assert.Equal(t, key1, key2, "Same filters should produce same key")
}

func TestCollectionsListKey_DifferentFilters(t *testing.T) {
    filters1 := &repository.ListFilters{Category: "art"}
    filters2 := &repository.ListFilters{Category: "music"}

    key1 := CollectionsListKey(filters1, 1, 20)
    key2 := CollectionsListKey(filters2, 1, 20)

    assert.NotEqual(t, key1, key2, "Different filters should produce different keys")
}
```

### 2. `services/collection-service/internal/repository/cached_collection_repository_test.go`

```go
package repository

import (
    "context"
    "errors"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
    sharedredis "github.com/zunokit/zuno-marketplace-api/shared/redis"
)

// MockCache implements cache interface for testing
type MockCache struct {
    mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
    args := m.Called(ctx, key, dest)
    return args.Error(0)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    args := m.Called(ctx, key, value, ttl)
    return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, keys ...string) error {
    args := m.Called(ctx, keys)
    return args.Error(0)
}

func (m *MockCache) Exists(ctx context.Context, keys ...string) (int64, error) {
    args := m.Called(ctx, keys)
    return args.Get(0).(int64), args.Error(1)
}

// MockCollectionRepository mocks the base repository
type MockCollectionRepository struct {
    mock.Mock
}

func (m *MockCollectionRepository) Create(ctx context.Context, collection *models.Collection) error {
    args := m.Called(ctx, collection)
    return args.Error(0)
}

func (m *MockCollectionRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
    args := m.Called(ctx, id, updates)
    return args.Error(0)
}

func (m *MockCollectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.Collection), args.Error(1)
}

func (m *MockCollectionRepository) GetByContractAddress(ctx context.Context, address, chainID string) (*models.Collection, error) {
    args := m.Called(ctx, address, chainID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.Collection), args.Error(1)
}

func (m *MockCollectionRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.Collection, int64, error) {
    args := m.Called(ctx, userID, page, limit)
    if args.Get(0) == nil {
        return nil, 0, args.Error(2)
    }
    return args.Get(0).([]*models.Collection), args.Get(1).(int64), args.Error(2)
}

func (m *MockCollectionRepository) List(ctx context.Context, filters *ListFilters, page, limit int) ([]*models.Collection, int64, error) {
    args := m.Called(ctx, filters, page, limit)
    if args.Get(0) == nil {
        return nil, 0, args.Error(2)
    }
    return args.Get(0).([]*models.Collection), args.Get(1).(int64), args.Error(2)
}

func (m *MockCollectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockCollectionRepository) IncrementTotalMinted(ctx context.Context, id uuid.UUID, increment int64) error {
    args := m.Called(ctx, id, increment)
    return args.Error(0)
}

func TestCachedCollectionRepository_GetByID_CacheHit(t *testing.T) {
    mockRepo := new(MockCollectionRepository)
    mockCache := new(MockCache)
    repo := NewCachedCollectionRepository(mockRepo, mockCache)

    id := uuid.New()
    expected := &models.Collection{ID: id, Name: "Test Collection"}

    // Cache hit
    mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
        Run(func(args mock.Arguments) {
            dest := args.Get(2).(*models.Collection)
            *dest = *expected
        }).
        Return(nil)

    result, err := repo.GetByID(context.Background(), id)

    assert.NoError(t, err)
    assert.Equal(t, expected, result)
    mockRepo.AssertNotCalled(t, "GetByID") // Should not hit DB
    mockCache.AssertExpectations(t)
}

func TestCachedCollectionRepository_GetByID_CacheMiss(t *testing.T) {
    mockRepo := new(MockCollectionRepository)
    mockCache := new(MockCache)
    repo := NewCachedCollectionRepository(mockRepo, mockCache)

    id := uuid.New()
    expected := &models.Collection{ID: id, Name: "Test Collection"}

    // Cache miss
    mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
        Return(errors.New("cache miss"))

    // DB hit
    mockRepo.On("GetByID", mock.Anything, id).Return(expected, nil)

    // Cache populate
    mockCache.On("Set", mock.Anything, mock.Anything, expected, mock.Anything).
        Return(nil)

    result, err := repo.GetByID(context.Background(), id)

    assert.NoError(t, err)
    assert.Equal(t, expected, result)
    mockRepo.AssertExpectations(t)
    mockCache.AssertExpectations(t)
}

func TestCachedCollectionRepository_Update_InvalidatesCache(t *testing.T) {
    mockRepo := new(MockCollectionRepository)
    mockCache := new(MockCache)
    repo := NewCachedCollectionRepository(mockRepo, mockCache)

    id := uuid.New()
    updates := map[string]interface{}{"name": "Updated Name"}

    mockRepo.On("Update", mock.Anything, id, updates).Return(nil)
    mockCache.On("Delete", mock.Anything, mock.Anything).Return(nil)

    err := repo.Update(context.Background(), id, updates)

    assert.NoError(t, err)
    mockCache.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestCachedCollectionRepository_CacheError_NonBlocking(t *testing.T) {
    mockRepo := new(MockCollectionRepository)
    mockCache := new(MockCache)
    repo := NewCachedCollectionRepository(mockRepo, mockCache)

    id := uuid.New()
    expected := &models.Collection{ID: id, Name: "Test Collection"}

    // Cache error
    mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
        Return(errors.New("redis unavailable"))

    // Should fall back to DB
    mockRepo.On("GetByID", mock.Anything, id).Return(expected, nil)

    // Set also errors (non-blocking)
    mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(errors.New("redis unavailable"))

    result, err := repo.GetByID(context.Background(), id)

    assert.NoError(t, err) // Should not fail
    assert.Equal(t, expected, result)
}
```

---

## Integration Tests

Create `services/collection-service/tests/cache_integration_test.go`:

```go
package tests

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/cache"
    "github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
    "github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/repository"
    sharedredis "github.com/zunokit/zuno-marketplace-api/shared/redis"
)

func TestCacheIntegration_GetByID(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup real Redis connection
    redisCache := sharedredis.NewCache()

    // Setup repository with real DB (use test database)
    // ... setup code ...

    baseRepo := repository.NewCollectionRepository(db)
    cachedRepo := repository.NewCachedCollectionRepository(baseRepo, redisCache)

    ctx := context.Background()

    // First call - cache miss
    start := time.Now()
    collection1, err := cachedRepo.GetByID(ctx, collectionID)
    missDuration := time.Since(start)
    require.NoError(t, err)

    // Second call - cache hit
    start = time.Now()
    collection2, err := cachedRepo.GetByID(ctx, collectionID)
    hitDuration := time.Since(start)
    require.NoError(t, err)

    // Cache hit should be faster
    assert.Less(t, hitDuration, missDuration/2)
    assert.Equal(t, collection1.ID, collection2.ID)
}
```

---

## Testing Checklist

- [ ] Unit tests for cache key generation
- [ ] Unit tests for cache-aside pattern
- [ ] Unit tests for cache invalidation
- [ ] Unit tests for non-blocking error handling
- [ ] Integration tests with real Redis
- [ ] Benchmark tests for cache hit vs miss performance
- [ ] All existing tests still pass

---

## Success Criteria

- [ ] > 80% test coverage for cache package
- [ ] All cache operations tested (hit, miss, invalidate)
- [ ] Error scenarios tested (Redis unavailable)
- [ ] Integration tests verify actual cache behavior

---

## Completion

After completing all phases:
1. Run full test suite
2. Verify cache hit rates in logs
3. Deploy to staging
4. Monitor metrics

**Plan Complete!**
