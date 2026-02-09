package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

// ListFilters defines filters for collection list queries
// Defined here to avoid import cycle with repository package
type ListFilters struct {
	SortBy      string
	SortOrder   string
	Category    string
	ChainID     string
	IsVerified  *bool
	SearchQuery string
}

// Key prefixes for different cache types
const (
	PrefixCollection         = "collection"
	PrefixCollectionContract = "collection:contract"
	PrefixCollectionStats    = "collection:stats"
	PrefixCollectionsUser    = "collections:user"
	PrefixCollectionsList    = "collections:list"
)

// CacheTTL represents cache TTL in seconds
type CacheTTL int

const (
	// NoTTL means cache until explicitly invalidated
	NoTTL CacheTTL = 0
	// StatsTTL is for frequently changing stats (5 minutes)
	StatsTTL CacheTTL = 5 * 60
	// ListTTL is for list queries (10 minutes)
	ListTTL CacheTTL = 10 * 60
	// SafetyTTL is a 24h fallback to prevent unlimited growth
	SafetyTTL CacheTTL = 24 * 60 * 60
)

// CollectionKey returns cache key for collection by ID
func CollectionKey(id uuid.UUID) string {
	return fmt.Sprintf("%s:%s", PrefixCollection, id.String())
}

// CollectionByContractKey returns cache key for collection by contract address
func CollectionByContractKey(chainID, address string) string {
	return fmt.Sprintf("%s:%s:%s", PrefixCollectionContract, chainID, address)
}

// CollectionStatsKey returns cache key for collection stats
func CollectionStatsKey(id uuid.UUID) string {
	return fmt.Sprintf("%s:%s:stats", PrefixCollection, id.String())
}

// CollectionsByUserKey returns cache key for user's collections list
func CollectionsByUserKey(userID uuid.UUID, page, limit int) string {
	return fmt.Sprintf("%s:%s:p%d:l%d", PrefixCollectionsUser, userID.String(), page, limit)
}

// CollectionsListKey returns cache key for filtered list
func CollectionsListKey(filters *ListFilters, page, limit int) string {
	// Create deterministic hash from filters
	filterStr := fmt.Sprintf("%s:%s:%s:%s:%v:%d:%d",
		filters.SortBy,
		filters.SortOrder,
		filters.Category,
		filters.ChainID,
		filters.IsVerified,
		page,
		limit,
	)
	if filters.SearchQuery != "" {
		filterStr += ":" + filters.SearchQuery
	}

	hash := sha256.Sum256([]byte(filterStr))
	hashStr := hex.EncodeToString(hash[:8]) // Use first 8 bytes for brevity

	return fmt.Sprintf("%s:%s", PrefixCollectionsList, hashStr)
}
