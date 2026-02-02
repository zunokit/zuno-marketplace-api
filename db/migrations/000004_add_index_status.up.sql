-- Migration: Add index_status tracking for collections
-- Description: Adds index_status and indexed_at fields to track indexer synchronization

-- Add index_status column to collections table
ALTER TABLE collections
ADD COLUMN IF NOT EXISTS index_status VARCHAR(20) DEFAULT 'NOT_INDEXED' CHECK (index_status IN ('NOT_INDEXED', 'INDEXING', 'INDEXED', 'FAILED')),
ADD COLUMN IF NOT EXISTS indexed_at TIMESTAMPTZ;

-- Add index for faster lookups by index_status
CREATE INDEX IF NOT EXISTS idx_collections_index_status
ON collections(index_status);

-- Add index for collections that need indexing (status DEPLOYED but not indexed)
CREATE INDEX IF NOT EXISTS idx_collections_needs_indexing
ON collections(status, index_status)
WHERE status = 'DEPLOYED' AND index_status != 'INDEXED';

-- Add comment
COMMENT ON COLUMN collections.index_status IS 'Indexer synchronization status';
COMMENT ON COLUMN collections.indexed_at IS 'Timestamp when collection was indexed';
