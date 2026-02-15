-- ============================================================================
-- PROCESSED EVENTS TABLE - WEBHOOK IDEMPOTENCY
-- Purpose: Track processed webhook events to prevent duplicate processing
-- Event ID format: "{tx_hash}:{log_index}"
-- ============================================================================

-- Table: processed_events
CREATE TABLE IF NOT EXISTS processed_events (
    -- Identity
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Event Identifier (unique per blockchain event)
    event_id VARCHAR(200) UNIQUE NOT NULL, -- Format: "0xabc...:0"

    -- Event Classification
    event_type VARCHAR(50) NOT NULL, -- collection.created, collection.minted, etc.
    chain_id VARCHAR(50) NOT NULL,   -- eip155:1, eip155:11155111, etc.

    -- Blockchain Metadata
    block_number BIGINT NOT NULL,
    tx_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,

    -- Collection Reference (optional, for filtering)
    collection_address VARCHAR(42),

    -- Processing Metadata
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT check_tx_hash_format CHECK (tx_hash ~ '^0x[a-f0-9]{64}$'),
    CONSTRAINT check_collection_address_format CHECK (
        collection_address IS NULL OR
        collection_address ~ '^0x[a-f0-9]{40}$'
    )
);

-- Indexes for performance
CREATE INDEX idx_processed_events_event_id ON processed_events(event_id);
CREATE INDEX idx_processed_events_event_type ON processed_events(event_type);
CREATE INDEX idx_processed_events_chain_id ON processed_events(chain_id);
CREATE INDEX idx_processed_events_collection_address ON processed_events(collection_address) WHERE collection_address IS NOT NULL;
CREATE INDEX idx_processed_events_processed_at ON processed_events(processed_at DESC);

-- Comments
COMMENT ON TABLE processed_events IS 'Track processed webhook events for idempotency';
COMMENT ON COLUMN processed_events.event_id IS 'Unique identifier: tx_hash:log_index';
COMMENT ON COLUMN processed_events.event_type IS 'Event type: collection.created, collection.minted, etc.';
COMMENT ON COLUMN processed_events.chain_id IS 'Chain ID in CAIP-2 format: eip155:{chain_id}';
COMMENT ON COLUMN processed_events.tx_hash IS 'Transaction hash with 0x prefix';
COMMENT ON COLUMN processed_events.log_index IS 'Log index within the transaction';
COMMENT ON COLUMN processed_events.collection_address IS 'Collection contract address (if applicable)';
