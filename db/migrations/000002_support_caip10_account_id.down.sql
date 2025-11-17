-- Migration: 000002_support_caip10_account_id.down.sql
-- Description: Revert CAIP-10 support and restore legacy Ethereum address format

BEGIN;

-- ======================= REVERT auth_nonces TABLE =======================

-- Drop CAIP-10 constraint
ALTER TABLE auth_nonces DROP CONSTRAINT IF EXISTS chk_account_format;

-- Revert account_id column to original length
ALTER TABLE auth_nonces ALTER COLUMN account_id TYPE varchar(42);

-- Restore original constraint (Ethereum address only)
ALTER TABLE auth_nonces
  ADD CONSTRAINT chk_account_format
  CHECK (account_id = lower(account_id) AND account_id ~ '^0x[0-9a-f]{40}$');

-- ======================= REVERT login_events TABLE =======================

-- Drop CAIP-10 constraint
ALTER TABLE login_events DROP CONSTRAINT IF EXISTS chk_login_account_format;

-- Revert account_id column to original length
ALTER TABLE login_events ALTER COLUMN account_id TYPE varchar(42);

-- Restore original constraint (Ethereum address only)
ALTER TABLE login_events
  ADD CONSTRAINT chk_login_account_format
  CHECK (account_id = lower(account_id) AND account_id ~ '^0x[0-9a-f]{40}$');

COMMIT;
