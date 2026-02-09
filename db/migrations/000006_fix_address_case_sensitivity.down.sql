-- ============================================================================
-- Rollback Address Case Sensitivity Fixes
-- ============================================================================

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_normalize_collection_addresses ON collections;
DROP TRIGGER IF EXISTS trigger_normalize_allowlist_addresses ON collection_allowlist;

-- Drop functions
DROP FUNCTION IF EXISTS normalize_ethereum_addresses();
DROP FUNCTION IF EXISTS normalize_allowlist_address();

-- Drop new indexes
DROP INDEX IF EXISTS idx_collections_contract_chain;
DROP INDEX IF EXISTS idx_collections_status_created;
DROP INDEX IF EXISTS idx_collections_user_created;
DROP INDEX IF EXISTS idx_collections_deployer_address;
DROP INDEX IF EXISTS idx_processed_events_event_id;

-- Recreate original indexes
CREATE INDEX idx_collections_contract_address ON collections(contract_address) WHERE contract_address IS NOT NULL;
CREATE INDEX idx_collections_deployer ON collections(deployer_address);
