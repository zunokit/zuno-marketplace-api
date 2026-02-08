package cache

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

func TestCollectionStatsKey(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	key := CollectionStatsKey(id)
	assert.Equal(t, "collection:550e8400-e29b-41d4-a716-446655440000:stats", key)
}

func TestCollectionsByUserKey(t *testing.T) {
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	key := CollectionsByUserKey(userID, 1, 20)
	assert.Equal(t, "collections:user:550e8400-e29b-41d4-a716-446655440000:p1:l20", key)
}

func TestCollectionsListKey_Deterministic(t *testing.T) {
	filters := &ListFilters{
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
	filters1 := &ListFilters{Category: "art"}
	filters2 := &ListFilters{Category: "music"}

	key1 := CollectionsListKey(filters1, 1, 20)
	key2 := CollectionsListKey(filters2, 1, 20)

	assert.NotEqual(t, key1, key2, "Different filters should produce different keys")
}

func TestCollectionsListKey_DifferentPages(t *testing.T) {
	filters := &ListFilters{Category: "art"}

	key1 := CollectionsListKey(filters, 1, 20)
	key2 := CollectionsListKey(filters, 2, 20)

	assert.NotEqual(t, key1, key2, "Different pages should produce different keys")
}

func TestTTLConstants(t *testing.T) {
	assert.Equal(t, CacheTTL(0), NoTTL)
	assert.Equal(t, CacheTTL(300), StatsTTL)   // 5 minutes
	assert.Equal(t, CacheTTL(600), ListTTL)    // 10 minutes
	assert.Equal(t, CacheTTL(86400), SafetyTTL) // 24 hours
}
