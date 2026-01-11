-- Migration: 000003_support_caip10_account_id.up.sql
-- Description: Update account_id to support CAIP-10 format (eip155:chainId:address)

BEGIN;

-- ======================= UPDATE auth_nonces TABLE =======================

-- Drop old constraint
ALTER TABLE auth_nonces DROP CONSTRAINT IF EXISTS chk_account_format;

-- Increase account_id length to support CAIP-10 format
-- Format: eip155:31337:0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 (~60 chars)
ALTER TABLE auth_nonces ALTER COLUMN account_id TYPE varchar(100);

-- Add new constraint to support both formats:
-- 1. CAIP-10: eip155:{chainId}:{address}
-- 2. Legacy Ethereum address: 0x{40 hex chars}
ALTER TABLE auth_nonces
  ADD CONSTRAINT chk_account_format
  CHECK (
    -- CAIP-10 format: eip155:chainId:0xaddress
    account_id ~ '^eip155:[0-9]+:0x[0-9a-f]{40}$'
    OR
    -- Legacy Ethereum address format
    (account_id = lower(account_id) AND account_id ~ '^0x[0-9a-f]{40}$')
  );

-- ======================= UPDATE login_events TABLE =======================

-- Drop old constraint
ALTER TABLE login_events DROP CONSTRAINT IF EXISTS chk_login_account_format;

-- Increase account_id length
ALTER TABLE login_events ALTER COLUMN account_id TYPE varchar(100);

-- Add new constraint to support both formats
ALTER TABLE login_events
  ADD CONSTRAINT chk_login_account_format
  CHECK (
    -- CAIP-10 format: eip155:chainId:0xaddress
    account_id ~ '^eip155:[0-9]+:0x[0-9a-f]{40}$'
    OR
    -- Legacy Ethereum address format
    (account_id = lower(account_id) AND account_id ~ '^0x[0-9a-f]{40}$')
  );

COMMIT;
