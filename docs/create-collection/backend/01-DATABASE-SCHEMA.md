# PHASE 1: DATABASE SCHEMA DESIGN

**Phase**: 1 of 6
**Dependencies**: None (foundation phase)
**Estimated Effort**: ~500 lines SQL
**Output**: Migration files ready, schema tested

---

## 🎯 Overview

Create database schema cho Collection feature. Schema đã được định nghĩa trong **main schema file** (`database-schema.md`) và phase này implement migration để apply schema vào database.

**Schema Reference**: See [`E:\zuno-marketplace-api\database-schema.md`](../../database-schema.md)

---

## 📊 Schema Summary

### Collections Tables (5 tables - ✨ NEW)

1. **collections** - Main collection records
   - Lifecycle tracking (PENDING → DEPLOYED → FAILED → ARCHIVED)
   - User ownership (user_id FK)
   - Blockchain binding (contract_address, chain_id, token_standard)
   - Minting configuration (prices, allowlist timing, limits)
   - Catalog features (slug, is_verified, is_hidden, source)
   - Royalty (EIP-2981 compatible)

2. **collection_metadata** - Extended metadata
   - IPFS links (metadata_uri, ipfs_hash, ipfs_url)
   - Social media URLs (twitter, discord, instagram, etc.)
   - Styling (background_color)

3. **collection_stats** - Aggregated statistics
   - Counts (total_items, total_owners, total_sales)
   - Pricing (floor_price_wei, total_volume_wei, average_price_wei)
   - 24h metrics (volume_24h_wei, sales_24h)
   - Last activity timestamps

4. **collection_activity** - Audit log
   - Activity types (CREATED, DEPLOYED, STATUS_CHANGED, METADATA_UPDATED, etc.)
   - User attribution (user_id FK)
   - Request context (ip_address, user_agent)
   - Details (JSONB flexible storage)

5. **collection_allowlist** - Whitelist management
   - Wallet addresses for ALLOWLIST mint stage
   - Per-address mint limits (max_mint_amount)
   - Audit trail (added_by_user_id, added_at)

---

## 📋 Prerequisites

- [ ] PostgreSQL 15+ running
- [ ] Database migrations tool (golang-migrate) installed
- [ ] Database connection configured in `.env`
- [ ] Existing tables: `users`, `user_stats` (required for FK)
- [ ] Main schema file reviewed: [`database-schema.md`](../../database-schema.md)

---

## 🗂️ Step 1: Create Migration Files

### Command

```bash
cd E:\zuno-marketplace-api
make migrate-create NAME=add_collections_tables
```

**Output**:
```
Created: db/migrations/000003_add_collections_tables.up.sql
Created: db/migrations/000003_add_collections_tables.down.sql
```

---

## 📝 Step 2: Implement UP Migration

**File**: `E:\zuno-marketplace-api\db\migrations\000003_add_collections_tables.up.sql`

**Source**: Derived from [`database-schema.md`](../../database-schema.md)

```sql
-- ============================================================================
-- COLLECTIONS SCHEMA
-- Based on: database-schema.md (COLLECTIONS, COLLECTION_METADATA, etc.)
-- ============================================================================

-- Table 1: collections
CREATE TABLE collections (
    -- Identity
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(100) UNIQUE,

    -- Ownership
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Basic Info
    name VARCHAR(100) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    description TEXT,
    category VARCHAR(50),

    -- Blockchain Binding
    contract_address VARCHAR(42),
    chain_id VARCHAR(50), -- CAIP-2: eip155:1
    token_standard VARCHAR(10) CHECK (token_standard IN ('ERC721', 'ERC1155')),
    deployer_address VARCHAR(42) NOT NULL,
    deployed_block BIGINT,

    -- Lifecycle Status
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'DEPLOYED', 'FAILED', 'ARCHIVED')),
    deployed_at TIMESTAMPTZ,

    -- Indexing Status
    index_status VARCHAR(20) DEFAULT 'NOT_STARTED'
        CHECK (index_status IN ('NOT_STARTED', 'SYNCING', 'SYNCED', 'FAILED')),

    -- Moderation
    is_verified BOOLEAN DEFAULT FALSE,
    is_hidden BOOLEAN DEFAULT FALSE,
    source VARCHAR(50) DEFAULT 'USER_CREATED', -- USER_CREATED, INDEXED, IMPORTED

    -- Media
    image_url TEXT NOT NULL,
    banner_url TEXT,
    featured_image_url TEXT,
    website_url TEXT,

    -- Social Links (JSONB for flexibility)
    social_links_json JSONB,

    -- Minting Configuration
    base_uri TEXT,
    max_supply BIGINT,
    mint_price_allowlist NUMERIC(78, 0), -- Wei
    mint_price_public NUMERIC(78, 0), -- Wei
    mint_start_time TIMESTAMPTZ,
    allowlist_stage_end TIMESTAMPTZ,
    mint_limit_per_wallet INT,

    -- Royalty (EIP-2981)
    royalty_fee_bps INT CHECK (royalty_fee_bps >= 0 AND royalty_fee_bps <= 10000),
    royalty_recipient VARCHAR(42),

    -- Stats (synced from blockchain)
    total_supply INT DEFAULT 0,
    total_minted INT DEFAULT 0,

    -- Metadata
    metadata_standard VARCHAR(20),
    tags_json JSONB,
    settings_json JSONB,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT unique_contract_per_chain UNIQUE (contract_address, chain_id),
    CONSTRAINT check_deployed_status CHECK (
        (status = 'DEPLOYED' AND contract_address IS NOT NULL) OR
        (status != 'DEPLOYED')
    )
);

-- Indexes
CREATE UNIQUE INDEX idx_collections_slug ON collections(slug);
CREATE INDEX idx_collections_user_id ON collections(user_id);
CREATE INDEX idx_collections_status ON collections(status);
CREATE INDEX idx_collections_contract_address ON collections(contract_address) WHERE contract_address IS NOT NULL;
CREATE INDEX idx_collections_chain_id ON collections(chain_id);
CREATE INDEX idx_collections_deployer ON collections(deployer_address);
CREATE INDEX idx_collections_created_at ON collections(created_at DESC);
CREATE INDEX idx_collections_token_standard ON collections(token_standard);
CREATE INDEX idx_collections_is_verified ON collections(is_verified) WHERE is_verified = TRUE;
CREATE INDEX idx_collections_source ON collections(source);

COMMENT ON TABLE collections IS 'NFT collections with lifecycle tracking and catalog features';

-- ============================================================================

-- Table 2: collection_metadata
CREATE TABLE collection_metadata (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    collection_id UUID NOT NULL UNIQUE REFERENCES collections(id) ON DELETE CASCADE,

    -- IPFS
    metadata_uri TEXT,
    ipfs_hash VARCHAR(100),
    ipfs_url TEXT,

    -- Social Media
    discord_url TEXT,
    twitter_url TEXT,
    instagram_url TEXT,
    medium_url TEXT,
    telegram_url TEXT,

    -- Styling
    background_color VARCHAR(6),

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_collection_metadata_collection_id ON collection_metadata(collection_id);
CREATE INDEX idx_collection_metadata_ipfs_hash ON collection_metadata(ipfs_hash) WHERE ipfs_hash IS NOT NULL;

-- ============================================================================

-- Table 3: collection_stats
CREATE TABLE collection_stats (
    collection_id UUID PRIMARY KEY REFERENCES collections(id) ON DELETE CASCADE,

    -- Counts
    total_items INT NOT NULL DEFAULT 0,
    total_owners INT NOT NULL DEFAULT 0,
    total_sales INT NOT NULL DEFAULT 0,

    -- Pricing (Wei)
    floor_price_wei NUMERIC(78, 0),
    total_volume_wei NUMERIC(78, 0) NOT NULL DEFAULT 0,
    average_price_wei NUMERIC(78, 0),

    -- 24h Stats
    volume_24h_wei NUMERIC(78, 0) DEFAULT 0,
    sales_24h INT DEFAULT 0,

    -- Last Activity
    last_sale_at TIMESTAMPTZ,
    last_mint_at TIMESTAMPTZ,

    -- Timestamp
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_collection_stats_floor_price ON collection_stats(floor_price_wei DESC NULLS LAST);
CREATE INDEX idx_collection_stats_volume ON collection_stats(total_volume_wei DESC);
CREATE INDEX idx_collection_stats_items ON collection_stats(total_items DESC);
CREATE INDEX idx_collection_stats_volume_24h ON collection_stats(volume_24h_wei DESC);

-- ============================================================================

-- Table 4: collection_activity
CREATE TABLE collection_activity (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Activity Info
    activity_type VARCHAR(50) NOT NULL,
    details JSONB,

    -- Request Context
    ip_address INET,
    user_agent TEXT,

    -- Timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_collection_activity_collection_id ON collection_activity(collection_id);
CREATE INDEX idx_collection_activity_user_id ON collection_activity(user_id);
CREATE INDEX idx_collection_activity_type ON collection_activity(activity_type);
CREATE INDEX idx_collection_activity_created_at ON collection_activity(created_at DESC);
CREATE INDEX idx_collection_activity_details ON collection_activity USING GIN (details);

-- ============================================================================

-- Table 5: collection_allowlist
CREATE TABLE collection_allowlist (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,

    -- Wallet Info
    wallet_address VARCHAR(42) NOT NULL,
    max_mint_amount INT DEFAULT 1,

    -- Audit
    added_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT unique_wallet_per_collection UNIQUE (collection_id, wallet_address)
);

CREATE INDEX idx_collection_allowlist_collection_id ON collection_allowlist(collection_id);
CREATE INDEX idx_collection_allowlist_wallet ON collection_allowlist(wallet_address);

-- ============================================================================
-- TRIGGERS & FUNCTIONS
-- ============================================================================

-- Function: Update updated_at
CREATE OR REPLACE FUNCTION update_collection_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_collection_updated_at
    BEFORE UPDATE ON collections
    FOR EACH ROW
    EXECUTE FUNCTION update_collection_updated_at();

CREATE TRIGGER trigger_update_collection_metadata_updated_at
    BEFORE UPDATE ON collection_metadata
    FOR EACH ROW
    EXECUTE FUNCTION update_collection_updated_at();

-- Function: Increment user collections count
CREATE OR REPLACE FUNCTION increment_user_collections_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE user_stats
    SET collections_count = collections_count + 1
    WHERE user_id = NEW.user_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_increment_user_collections_count
    AFTER INSERT ON collections
    FOR EACH ROW
    EXECUTE FUNCTION increment_user_collections_count();

-- Function: Decrement user collections count
CREATE OR REPLACE FUNCTION decrement_user_collections_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE user_stats
    SET collections_count = collections_count - 1
    WHERE user_id = OLD.user_id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_decrement_user_collections_count
    AFTER DELETE ON collections
    FOR EACH ROW
    EXECUTE FUNCTION decrement_user_collections_count();

-- Function: Create collection defaults
CREATE OR REPLACE FUNCTION create_collection_defaults()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO collection_stats (collection_id) VALUES (NEW.id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_create_collection_defaults
    AFTER INSERT ON collections
    FOR EACH ROW
    EXECUTE FUNCTION create_collection_defaults();

-- Function: Log collection activity
CREATE OR REPLACE FUNCTION log_collection_activity()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO collection_activity (collection_id, activity_type, details)
        VALUES (NEW.id, 'CREATED', jsonb_build_object(
            'name', NEW.name,
            'symbol', NEW.symbol,
            'token_standard', NEW.token_standard
        ));
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.status != NEW.status THEN
            INSERT INTO collection_activity (collection_id, activity_type, details)
            VALUES (NEW.id, 'STATUS_CHANGED', jsonb_build_object(
                'old_status', OLD.status,
                'new_status', NEW.status,
                'contract_address', NEW.contract_address
            ));
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_log_collection_activity
    AFTER INSERT OR UPDATE ON collections
    FOR EACH ROW
    EXECUTE FUNCTION log_collection_activity();

-- Function: Cleanup stale pending collections
CREATE OR REPLACE FUNCTION cleanup_pending_collections()
RETURNS INT AS $$
DECLARE
    deleted_count INT;
BEGIN
    DELETE FROM collections
    WHERE status = 'PENDING'
      AND created_at < NOW() - INTERVAL '24 hours';

    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
```

---

## 📝 Step 3: Implement DOWN Migration

**File**: `E:\zuno-marketplace-api\db\migrations\000003_add_collections_tables.down.sql`

```sql
-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_collection_updated_at ON collections;
DROP TRIGGER IF EXISTS trigger_update_collection_metadata_updated_at ON collection_metadata;
DROP TRIGGER IF EXISTS trigger_increment_user_collections_count ON collections;
DROP TRIGGER IF EXISTS trigger_decrement_user_collections_count ON collections;
DROP TRIGGER IF EXISTS trigger_create_collection_defaults ON collections;
DROP TRIGGER IF EXISTS trigger_log_collection_activity ON collections;

-- Drop functions
DROP FUNCTION IF EXISTS update_collection_updated_at();
DROP FUNCTION IF EXISTS increment_user_collections_count();
DROP FUNCTION IF EXISTS decrement_user_collections_count();
DROP FUNCTION IF EXISTS create_collection_defaults();
DROP FUNCTION IF EXISTS log_collection_activity();
DROP FUNCTION IF EXISTS cleanup_pending_collections();

-- Drop tables (order matters)
DROP TABLE IF EXISTS collection_allowlist;
DROP TABLE IF EXISTS collection_activity;
DROP TABLE IF EXISTS collection_stats;
DROP TABLE IF EXISTS collection_metadata;
DROP TABLE IF EXISTS collections;
```

---

## ✅ Step 4-6: Apply, Test, Seed

See original implementation guide for:
- Migration application commands
- Test queries
- Seed data examples

---

## ✅ Completion Checklist

- [ ] Migration files created
- [ ] Schema matches `database-schema.md`
- [ ] All 5 tables created
- [ ] All indexes created
- [ ] All triggers working
- [ ] User stats integration working
- [ ] Test data loaded
- [ ] Rollback tested

---

## 🔗 Related Documentation

- **Main Schema**: [`database-schema.md`](../../database-schema.md) - Source of truth
- **Next Phase**: [02-COLLECTION-SERVICE.md](./02-COLLECTION-SERVICE.md)

---

**Last Updated**: 2025-11-20
**Status**: Ready for implementation (based on database-schema.md)
