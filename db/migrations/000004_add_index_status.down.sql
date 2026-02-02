-- Rollback: Remove index_status tracking from collections

-- Drop indexes
DROP INDEX IF EXISTS idx_collections_needs_indexing;
DROP INDEX IF EXISTS idx_collections_index_status;

-- Drop columns
ALTER TABLE collections
DROP COLUMN IF EXISTS indexed_at,
DROP COLUMN IF EXISTS index_status;
