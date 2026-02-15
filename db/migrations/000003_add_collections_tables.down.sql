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
