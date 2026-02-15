package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Custom Types
// ============================================================================

// JSONB represents a JSONB column in PostgreSQL
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &j)
}

// ============================================================================
// Enums
// ============================================================================

type CollectionStatus string

const (
	CollectionStatusPending  CollectionStatus = "PENDING"
	CollectionStatusDeployed CollectionStatus = "DEPLOYED"
	CollectionStatusFailed   CollectionStatus = "FAILED"
	CollectionStatusArchived CollectionStatus = "ARCHIVED"
)

type IndexStatus string

const (
	IndexStatusNotStarted IndexStatus = "NOT_STARTED"
	IndexStatusSyncing    IndexStatus = "SYNCING"
	IndexStatusSynced     IndexStatus = "SYNCED"
	IndexStatusFailed     IndexStatus = "FAILED"
)

type TokenStandard string

const (
	TokenStandardERC721  TokenStandard = "ERC721"
	TokenStandardERC1155 TokenStandard = "ERC1155"
)

// ============================================================================
// Models
// ============================================================================

// Collection represents the main collection entity
type Collection struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Slug        *string   `gorm:"size:100;uniqueIndex" json:"slug,omitempty"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Symbol      string    `gorm:"size:10;not null" json:"symbol"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	Category    *string   `gorm:"size:50" json:"category,omitempty"`

	// Blockchain Binding
	ContractAddress *string        `gorm:"size:42;index" json:"contract_address,omitempty"`
	ChainID         *string        `gorm:"size:50;index" json:"chain_id,omitempty"`
	TokenStandard   *TokenStandard `gorm:"size:10;index" json:"token_standard,omitempty"`
	DeployerAddress string         `gorm:"size:42;not null;index" json:"deployer_address"`
	DeployedBlock   *int64         `json:"deployed_block,omitempty"`

	// Lifecycle
	Status     CollectionStatus `gorm:"size:20;not null;default:'PENDING';index" json:"status"`
	DeployedAt *time.Time       `json:"deployed_at,omitempty"`

	// Indexing
	IndexStatus IndexStatus `gorm:"size:20;default:'NOT_STARTED'" json:"index_status"`

	// Moderation
	IsVerified bool   `gorm:"default:false;index" json:"is_verified"`
	IsHidden   bool   `gorm:"default:false" json:"is_hidden"`
	Source     string `gorm:"size:50;default:'USER_CREATED';index" json:"source"`

	// Media
	ImageURL         string  `gorm:"type:text;not null" json:"image_url"`
	BannerURL        *string `gorm:"type:text" json:"banner_url,omitempty"`
	FeaturedImageURL *string `gorm:"type:text" json:"featured_image_url,omitempty"`
	WebsiteURL       *string `gorm:"type:text" json:"website_url,omitempty"`

	// Social Links
	SocialLinksJSON JSONB `gorm:"type:jsonb" json:"social_links_json,omitempty"`

	// Minting Configuration
	BaseURI            *string    `gorm:"type:text" json:"base_uri,omitempty"`
	MaxSupply          *int64     `json:"max_supply,omitempty"`
	MintPriceAllowlist *string    `gorm:"type:numeric(78,0)" json:"mint_price_allowlist,omitempty"` // Wei
	MintPricePublic    *string    `gorm:"type:numeric(78,0)" json:"mint_price_public,omitempty"`    // Wei
	MintStartTime      *time.Time `json:"mint_start_time,omitempty"`
	AllowlistStageEnd  *time.Time `json:"allowlist_stage_end,omitempty"`
	MintLimitPerWallet *int       `json:"mint_limit_per_wallet,omitempty"`

	// Royalty
	RoyaltyFeeBPS    *int    `json:"royalty_fee_bps,omitempty"`
	RoyaltyRecipient *string `gorm:"size:42" json:"royalty_recipient,omitempty"`

	// Stats (synced)
	TotalSupply int `gorm:"default:0" json:"total_supply"`
	TotalMinted int `gorm:"default:0" json:"total_minted"`

	// Metadata
	MetadataStandard *string `gorm:"size:20" json:"metadata_standard,omitempty"`
	TagsJSON         JSONB   `gorm:"type:jsonb" json:"tags_json,omitempty"`
	SettingsJSON     JSONB   `gorm:"type:jsonb" json:"settings_json,omitempty"`

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:now();index" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`

	// Relations
	Metadata  *CollectionMetadata   `gorm:"foreignKey:CollectionID;constraint:OnDelete:CASCADE" json:"metadata,omitempty"`
	Stats     *CollectionStats      `gorm:"foreignKey:CollectionID;constraint:OnDelete:CASCADE" json:"stats,omitempty"`
	Allowlist []CollectionAllowlist `gorm:"foreignKey:CollectionID;constraint:OnDelete:CASCADE" json:"allowlist,omitempty"`
}

func (Collection) TableName() string {
	return "collections"
}

// CollectionMetadata represents extended metadata
type CollectionMetadata struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CollectionID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"collection_id"`

	// IPFS
	MetadataURI *string `gorm:"type:text" json:"metadata_uri,omitempty"`
	IPFSHash    *string `gorm:"column:ipfs_hash;size:100;index" json:"ipfs_hash,omitempty"`
	IPFSURL     *string `gorm:"column:ipfs_url;type:text" json:"ipfs_url,omitempty"`

	// Social Media
	DiscordURL   *string `gorm:"type:text" json:"discord_url,omitempty"`
	TwitterURL   *string `gorm:"type:text" json:"twitter_url,omitempty"`
	InstagramURL *string `gorm:"type:text" json:"instagram_url,omitempty"`
	MediumURL    *string `gorm:"type:text" json:"medium_url,omitempty"`
	TelegramURL  *string `gorm:"type:text" json:"telegram_url,omitempty"`

	// Styling
	BackgroundColor *string `gorm:"size:6" json:"background_color,omitempty"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (CollectionMetadata) TableName() string {
	return "collection_metadata"
}

// CollectionStats represents aggregated statistics
type CollectionStats struct {
	CollectionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"collection_id"`

	TotalItems  int `gorm:"not null;default:0;index" json:"total_items"`
	TotalOwners int `gorm:"not null;default:0" json:"total_owners"`
	TotalSales  int `gorm:"not null;default:0" json:"total_sales"`

	FloorPriceWei   *string `gorm:"type:numeric(78,0);index" json:"floor_price_wei,omitempty"`
	TotalVolumeWei  string  `gorm:"type:numeric(78,0);not null;default:0;index" json:"total_volume_wei"`
	AveragePriceWei *string `gorm:"type:numeric(78,0)" json:"average_price_wei,omitempty"`

	Volume24hWei string `gorm:"type:numeric(78,0);default:0;index" json:"volume_24h_wei"`
	Sales24h     int    `gorm:"default:0" json:"sales_24h"`

	LastSaleAt *time.Time `json:"last_sale_at,omitempty"`
	LastMintAt *time.Time `json:"last_mint_at,omitempty"`

	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (CollectionStats) TableName() string {
	return "collection_stats"
}

// CollectionActivity represents audit logs
type CollectionActivity struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CollectionID uuid.UUID  `gorm:"type:uuid;not null;index" json:"collection_id"`
	UserID       *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`

	ActivityType string `gorm:"size:50;not null;index" json:"activity_type"`
	Details      JSONB  `gorm:"type:jsonb" json:"details,omitempty"`

	IPAddress *string `gorm:"type:inet" json:"ip_address,omitempty"`
	UserAgent *string `gorm:"type:text" json:"user_agent,omitempty"`

	CreatedAt time.Time `gorm:"not null;default:now();index" json:"created_at"`
}

func (CollectionActivity) TableName() string {
	return "collection_activity"
}

// CollectionAllowlist represents whitelist entries
type CollectionAllowlist struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CollectionID uuid.UUID `gorm:"type:uuid;not null;index" json:"collection_id"`

	WalletAddress string `gorm:"size:42;not null;index" json:"wallet_address"`
	MaxMintAmount int    `gorm:"default:1" json:"max_mint_amount"`

	AddedByUserID *uuid.UUID `gorm:"type:uuid" json:"added_by_user_id,omitempty"`
	AddedAt       time.Time  `gorm:"not null;default:now()" json:"added_at"`
	CreatedAt     time.Time  `gorm:"not null;default:now()" json:"created_at"`
}

func (CollectionAllowlist) TableName() string {
	return "collection_allowlist"
}
