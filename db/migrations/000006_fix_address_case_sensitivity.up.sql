-- ============================================================================
-- Fix Ethereum Address Case Sensitivity Issues
-- ============================================================================

-- 1. Normalize existing contract addresses to lowercase
UPDATE collections 
SET contract_address = LOWER(contract_address) 
WHERE contract_address IS NOT NULL;

UPDATE collection_allowlist
SET wallet_address = LOWER(wallet_address);

-- 2. Add function to automatically lowercase addresses on insert/update
CREATE OR REPLACE FUNCTION normalize_ethereum_addresses()
RETURNS TRIGGER AS $$
BEGIN
    -- Normalize contract_address
    IF NEW.contract_address IS NOT NULL THEN
        NEW.contract_address = LOWER(NEW.contract_address);
    END IF;
    
    -- Normalize deployer_address
    IF NEW.deployer_address IS NOT NULL THEN
        NEW.deployer_address = LOWER(NEW.deployer_address);
    END IF;
    
    -- Normalize royalty_recipient
    IF NEW.royalty_recipient IS NOT NULL THEN
        NEW.royalty_recipient = LOWER(NEW.royalty_recipient);
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 3. Create trigger to normalize addresses before insert/update
CREATE TRIGGER trigger_normalize_collection_addresses
    BEFORE INSERT OR UPDATE ON collections
    FOR EACH ROW
    EXECUTE FUNCTION normalize_ethereum_addresses();

-- 4. Add trigger for allowlist addresses
CREATE OR REPLACE FUNCTION normalize_allowlist_address()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.wallet_address IS NOT NULL THEN
        NEW.wallet_address = LOWER(NEW.wallet_address);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_normalize_allowlist_addresses
    BEFORE INSERT OR UPDATE ON collection_allowlist
    FOR EACH ROW
    EXECUTE FUNCTION normalize_allowlist_address();

-- 5. Drop existing index and recreate as composite index for better performance
DROP INDEX IF EXISTS idx_collections_contract_address;

-- Create composite index on (contract_address, chain_id) for faster lookups
CREATE UNIQUE INDEX idx_collections_contract_chain 
ON collections(contract_address, chain_id) 
WHERE contract_address IS NOT NULL;

-- 6. Add index on deployer_address for faster queries
DROP INDEX IF EXISTS idx_collections_deployer;
CREATE INDEX idx_collections_deployer_address ON collections(deployer_address);

-- 7. Add composite index for common query patterns
CREATE INDEX idx_collections_status_created ON collections(status, created_at DESC);
CREATE INDEX idx_collections_user_created ON collections(user_id, created_at DESC);

-- 8. Add index on processed_events for faster duplicate checks
CREATE INDEX IF NOT EXISTS idx_processed_events_event_id ON processed_events(event_id);

-- Comments
COMMENT ON FUNCTION normalize_ethereum_addresses() IS 'Automatically converts Ethereum addresses to lowercase for case-insensitive comparisons';
COMMENT ON INDEX idx_collections_contract_chain IS 'Composite index for fast contract address lookups with chain_id';
