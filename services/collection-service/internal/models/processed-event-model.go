package models

import "time"

// ProcessedEvent represents a processed webhook event for idempotency
type ProcessedEvent struct {
	// Identity
	ID string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	// Event Identifier (unique per blockchain event)
	EventID string `gorm:"column:event_id;type:varchar(200);uniqueIndex;not null" json:"event_id"` // Format: "0xabc...:0"

	// Event Classification
	EventType string `gorm:"column:event_type;type:varchar(50);not null;index" json:"event_type"` // collection.created, collection.minted, etc.
	ChainID   string `gorm:"column:chain_id;type:varchar(50);not null;index" json:"chain_id"`     // eip155:1, eip155:11155111, etc.

	// Blockchain Metadata
	BlockNumber       int64  `gorm:"column:block_number;type:bigint;not null" json:"block_number"`
	TxHash            string `gorm:"column:tx_hash;type:varchar(66);not null" json:"tx_hash"`
	LogIndex          int    `gorm:"column:log_index;type:int;not null" json:"log_index"`
	CollectionAddress string `gorm:"column:collection_address;type:varchar(42);index" json:"collection_address"` // Optional

	// Processing Metadata
	ProcessedAt time.Time `gorm:"column:processed_at;type:timestamptz;not null;default:now()" json:"processed_at"`
}

// TableName specifies the table name for ProcessedEvent
func (ProcessedEvent) TableName() string {
	return "processed_events"
}

// NewProcessedEvent creates a new ProcessedEvent instance
func NewProcessedEvent(eventID, eventType, chainID string, blockNumber int64, txHash string, logIndex int, collectionAddress string) *ProcessedEvent {
	return &ProcessedEvent{
		EventID:           eventID,
		EventType:         eventType,
		ChainID:           chainID,
		BlockNumber:       blockNumber,
		TxHash:            txHash,
		LogIndex:          logIndex,
		CollectionAddress: collectionAddress,
		ProcessedAt:       time.Now(),
	}
}
