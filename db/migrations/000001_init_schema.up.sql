BEGIN;

-- Extensions (used by all services)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ======================= AUTH SERVICE SCHEMA =======================

-- NONCE MANAGEMENT
CREATE TABLE IF NOT EXISTS auth_nonces (
    nonce       varchar(64)  PRIMARY KEY,
    account_id  varchar(42)  NOT NULL,
    domain      varchar(255) NOT NULL,
    chain_id    varchar(32)  NOT NULL,
    issued_at   timestamptz  NOT NULL DEFAULT now(),
    expires_at  timestamptz  NOT NULL,
    used        boolean      NOT NULL DEFAULT FALSE,
    used_at     timestamptz,
    created_at  timestamptz  NOT NULL DEFAULT now()
);

ALTER TABLE auth_nonces
  DROP CONSTRAINT IF EXISTS chk_nonce_expiry,
  ADD  CONSTRAINT chk_nonce_expiry
  CHECK (expires_at > issued_at AND expires_at <= issued_at + INTERVAL '10 minutes');

ALTER TABLE auth_nonces
  DROP CONSTRAINT IF EXISTS chk_account_format,
  ADD  CONSTRAINT chk_account_format
  CHECK (account_id = lower(account_id) AND account_id ~ '^0x[0-9a-f]{40}$');

CREATE INDEX IF NOT EXISTS idx_auth_nonces_expires_at  ON auth_nonces(expires_at);
CREATE INDEX IF NOT EXISTS idx_auth_nonces_account_id  ON auth_nonces(account_id);
CREATE INDEX IF NOT EXISTS idx_auth_nonces_used        ON auth_nonces(used);

-- SESSION MANAGEMENT
CREATE TABLE IF NOT EXISTS sessions (
    session_id   uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid         NOT NULL,
    device_id    uuid,
    refresh_hash varchar(128) NOT NULL,
    previous_refresh_hash varchar(128),
    token_family_id uuid     NOT NULL DEFAULT gen_random_uuid(),
    token_generation integer NOT NULL DEFAULT 1,
    ip_address   inet,
    user_agent   text,
    created_at   timestamptz  NOT NULL DEFAULT now(),
    expires_at   timestamptz  NOT NULL,
    revoked_at   timestamptz,
    last_used_at timestamptz  DEFAULT now(),
    revoked_reason varchar(255),
    collection_intent_context JSONB DEFAULT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_sessions_refresh_hash ON sessions(refresh_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id           ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at        ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_token_family     ON sessions(token_family_id);
CREATE INDEX IF NOT EXISTS idx_sessions_prev_refresh     ON sessions(previous_refresh_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_active_by_user   ON sessions(user_id) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_sessions_collection_context ON sessions(user_id) WHERE collection_intent_context IS NOT NULL;

ALTER TABLE sessions
  DROP CONSTRAINT IF EXISTS chk_session_expiry,
  ADD  CONSTRAINT chk_session_expiry CHECK (expires_at > created_at);

ALTER TABLE sessions
  DROP CONSTRAINT IF EXISTS chk_session_revoked,
  ADD  CONSTRAINT chk_session_revoked CHECK (revoked_at IS NULL OR revoked_at >= created_at);

-- AUDIT LOGGING
CREATE TABLE IF NOT EXISTS login_events (
    id           uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid,
    account_id   varchar(42)  NOT NULL,
    ip_address   inet,
    user_agent   text,
    result       varchar(32)  NOT NULL,
    error_message text,
    chain_id     varchar(32),
    domain       varchar(255),
    timestamp    timestamptz  NOT NULL DEFAULT now()
);

ALTER TABLE login_events
  DROP CONSTRAINT IF EXISTS chk_login_account_format,
  ADD  CONSTRAINT chk_login_account_format
  CHECK (account_id = lower(account_id) AND account_id ~ '^0x[0-9a-f]{40}$');

ALTER TABLE login_events
  DROP CONSTRAINT IF EXISTS chk_login_result,
  ADD  CONSTRAINT chk_login_result
  CHECK (result IN ('success','failed','invalid_signature','invalid_nonce','expired_nonce','invalid_message','rate_limited'));

CREATE INDEX IF NOT EXISTS idx_login_events_user_id     ON login_events(user_id);
CREATE INDEX IF NOT EXISTS idx_login_events_account_id  ON login_events(account_id);
CREATE INDEX IF NOT EXISTS idx_login_events_timestamp   ON login_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_login_events_result      ON login_events(result);
CREATE INDEX IF NOT EXISTS idx_login_events_ip_address  ON login_events(ip_address);

-- ======================= USER SERVICE SCHEMA =======================

-- USERS
CREATE TABLE IF NOT EXISTS users (
    user_id     uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    status      varchar(32)  NOT NULL DEFAULT 'active',
    created_at  timestamptz  NOT NULL DEFAULT now(),
    updated_at  timestamptz  NOT NULL DEFAULT now()
);

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS chk_user_status,
  ADD  CONSTRAINT chk_user_status
  CHECK (status IN ('active', 'banned', 'deleted', 'suspended'));

CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- PROFILES
CREATE TABLE IF NOT EXISTS profiles (
    user_id      uuid         PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    username     varchar(30)  UNIQUE,
    display_name varchar(50),
    avatar_url   text,
    banner_url   text,
    bio          text,
    locale       varchar(10)  DEFAULT 'en',
    timezone     varchar(50)  DEFAULT 'UTC',
    socials_json jsonb,
    updated_at   timestamptz  NOT NULL DEFAULT now()
);

ALTER TABLE profiles
  DROP CONSTRAINT IF EXISTS chk_username_format,
  ADD  CONSTRAINT chk_username_format
  CHECK (username ~ '^[a-zA-Z0-9_]{3,30}$');

ALTER TABLE profiles
  DROP CONSTRAINT IF EXISTS chk_bio_length,
  ADD  CONSTRAINT chk_bio_length
  CHECK (char_length(bio) <= 500);

ALTER TABLE profiles
  DROP CONSTRAINT IF EXISTS chk_display_name_length,
  ADD  CONSTRAINT chk_display_name_length
  CHECK (char_length(display_name) <= 50);

CREATE INDEX IF NOT EXISTS idx_profiles_username ON profiles(username);
CREATE INDEX IF NOT EXISTS idx_profiles_updated_at ON profiles(updated_at);

-- USER PREFERENCES
CREATE TABLE IF NOT EXISTS user_preferences (
    user_id               uuid         PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    email_notifications   boolean      DEFAULT true,
    push_notifications    boolean      DEFAULT true,
    marketing_emails      boolean      DEFAULT false,
    language              varchar(10)  DEFAULT 'en',
    currency              varchar(10)  DEFAULT 'USD',
    theme                 varchar(20)  DEFAULT 'light',
    privacy_level         varchar(20)  DEFAULT 'public',
    show_activity         boolean      DEFAULT true,
    updated_at            timestamptz  NOT NULL DEFAULT now()
);

ALTER TABLE user_preferences
  DROP CONSTRAINT IF EXISTS chk_theme,
  ADD  CONSTRAINT chk_theme
  CHECK (theme IN ('light', 'dark', 'auto'));

ALTER TABLE user_preferences
  DROP CONSTRAINT IF EXISTS chk_privacy_level,
  ADD  CONSTRAINT chk_privacy_level
  CHECK (privacy_level IN ('public', 'private', 'friends'));

-- USER STATS
CREATE TABLE IF NOT EXISTS user_stats (
    user_id           uuid         PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    collections_count integer      DEFAULT 0,
    items_count       integer      DEFAULT 0,
    listings_count    integer      DEFAULT 0,
    sales_count       integer      DEFAULT 0,
    purchases_count   integer      DEFAULT 0,
    volume_sold       numeric(20,8) DEFAULT 0,
    volume_purchased  numeric(20,8) DEFAULT 0,
    followers_count   integer      DEFAULT 0,
    following_count   integer      DEFAULT 0,
    updated_at        timestamptz  NOT NULL DEFAULT now()
);

ALTER TABLE user_stats
  ADD CONSTRAINT chk_stats_non_negative CHECK (
    collections_count >= 0 AND
    items_count >= 0 AND
    listings_count >= 0 AND
    sales_count >= 0 AND
    purchases_count >= 0 AND
    volume_sold >= 0 AND
    volume_purchased >= 0 AND
    followers_count >= 0 AND
    following_count >= 0
  );

-- USER FOLLOWS
CREATE TABLE IF NOT EXISTS user_follows (
    follower_id  uuid         NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    following_id uuid         NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    created_at   timestamptz  NOT NULL DEFAULT now(),
    PRIMARY KEY (follower_id, following_id)
);

ALTER TABLE user_follows
  ADD CONSTRAINT chk_no_self_follow CHECK (follower_id != following_id);

CREATE INDEX IF NOT EXISTS idx_user_follows_follower ON user_follows(follower_id);
CREATE INDEX IF NOT EXISTS idx_user_follows_following ON user_follows(following_id);

-- ======================= WALLET SERVICE SCHEMA =======================

-- WALLET LINKS
CREATE TABLE IF NOT EXISTS wallet_links (
    wallet_id    uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid         NOT NULL,
    account_id   varchar(255) NOT NULL,
    address      varchar(42)  NOT NULL,
    chain_id     varchar(32)  NOT NULL,
    is_primary   boolean      NOT NULL DEFAULT false,
    type         varchar(20)  DEFAULT 'eoa',
    connector    varchar(50),
    label        varchar(100),
    verified_at  timestamptz  NOT NULL DEFAULT now(),
    created_at   timestamptz  NOT NULL DEFAULT now(),
    updated_at   timestamptz  NOT NULL DEFAULT now()
);

ALTER TABLE wallet_links
  DROP CONSTRAINT IF EXISTS chk_address_format,
  ADD  CONSTRAINT chk_address_format
  CHECK (address = lower(address) AND address ~ '^0x[0-9a-f]{40}$');

ALTER TABLE wallet_links
  DROP CONSTRAINT IF EXISTS chk_chain_id_format,
  ADD  CONSTRAINT chk_chain_id_format
  CHECK (chain_id ~ '^[a-z0-9]+:[a-zA-Z0-9]+$');

ALTER TABLE wallet_links
  DROP CONSTRAINT IF EXISTS chk_wallet_type,
  ADD  CONSTRAINT chk_wallet_type
  CHECK (type IN ('eoa', 'contract', 'multisig', 'smart_account'));

CREATE UNIQUE INDEX IF NOT EXISTS uq_wallet_user_address_chain ON wallet_links(user_id, address, chain_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_wallet_primary ON wallet_links(user_id) WHERE is_primary = true;
CREATE INDEX IF NOT EXISTS idx_wallet_links_user_id ON wallet_links(user_id);
CREATE INDEX IF NOT EXISTS idx_wallet_links_address ON wallet_links(address);
CREATE INDEX IF NOT EXISTS idx_wallet_links_chain_id ON wallet_links(chain_id);
CREATE INDEX IF NOT EXISTS idx_wallet_links_created_at ON wallet_links(created_at);

-- WALLET ACTIVITY
CREATE TABLE IF NOT EXISTS wallet_activity (
    id           uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id    uuid         NOT NULL REFERENCES wallet_links(wallet_id) ON DELETE CASCADE,
    user_id      uuid         NOT NULL,
    action       varchar(50)  NOT NULL,
    metadata     jsonb,
    ip_address   inet,
    user_agent   text,
    created_at   timestamptz  NOT NULL DEFAULT now()
);

ALTER TABLE wallet_activity
  DROP CONSTRAINT IF EXISTS chk_activity_action,
  ADD  CONSTRAINT chk_activity_action
  CHECK (action IN ('linked', 'unlinked', 'set_primary', 'verified', 'updated'));

CREATE INDEX IF NOT EXISTS idx_wallet_activity_wallet_id ON wallet_activity(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_activity_user_id ON wallet_activity(user_id);
CREATE INDEX IF NOT EXISTS idx_wallet_activity_created_at ON wallet_activity(created_at);
CREATE INDEX IF NOT EXISTS idx_wallet_activity_action ON wallet_activity(action);

-- WALLET VERIFICATION
CREATE TABLE IF NOT EXISTS wallet_verifications (
    id               uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id        uuid         NOT NULL REFERENCES wallet_links(wallet_id) ON DELETE CASCADE,
    verification_type varchar(50) NOT NULL,
    verification_data jsonb      NOT NULL,
    status           varchar(20)  NOT NULL DEFAULT 'pending',
    verified_at      timestamptz,
    expires_at       timestamptz,
    created_at       timestamptz  NOT NULL DEFAULT now()
);

ALTER TABLE wallet_verifications
  DROP CONSTRAINT IF EXISTS chk_verification_status,
  ADD  CONSTRAINT chk_verification_status
  CHECK (status IN ('pending', 'verified', 'failed', 'expired'));

CREATE INDEX IF NOT EXISTS idx_wallet_verifications_wallet_id ON wallet_verifications(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_verifications_status ON wallet_verifications(status);
CREATE INDEX IF NOT EXISTS idx_wallet_verifications_expires_at ON wallet_verifications(expires_at);

-- ======================= FUNCTIONS =======================

-- Auth service functions
CREATE OR REPLACE FUNCTION cleanup_expired_nonces()
RETURNS integer AS $$
DECLARE
  deleted_count integer;
BEGIN
  DELETE FROM auth_nonces WHERE expires_at < now() - INTERVAL '1 hour';
  GET DIAGNOSTICS deleted_count = ROW_COUNT;
  RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION cleanup_old_login_events(retention_days integer DEFAULT 90)
RETURNS integer AS $$
DECLARE
  deleted_count integer;
BEGIN
  DELETE FROM login_events WHERE timestamp < now() - (retention_days || ' days')::interval;
  GET DIAGNOSTICS deleted_count = ROW_COUNT;
  RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS integer AS $$
DECLARE
  updated_count integer;
BEGIN
  UPDATE sessions SET revoked_at = now() WHERE expires_at < now() AND revoked_at IS NULL;
  GET DIAGNOSTICS updated_count = ROW_COUNT;
  RETURN updated_count;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION try_use_nonce(
  p_nonce   text,
  p_account text,
  p_chain   text,
  p_domain  text
) RETURNS boolean AS $$
DECLARE
  affected integer;
BEGIN
  UPDATE auth_nonces
     SET used = TRUE, used_at = now()
   WHERE nonce = p_nonce
     AND account_id = p_account
     AND chain_id = p_chain
     AND domain = p_domain
     AND used = FALSE
     AND expires_at > now();
  GET DIAGNOSTICS affected = ROW_COUNT;
  RETURN affected = 1;
END;
$$ LANGUAGE plpgsql;

-- User service functions
CREATE OR REPLACE FUNCTION update_user_stats()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_TABLE_NAME = 'user_follows' THEN
        IF TG_OP = 'INSERT' THEN
            UPDATE user_stats SET followers_count = followers_count + 1, updated_at = now() WHERE user_id = NEW.following_id;
            UPDATE user_stats SET following_count = following_count + 1, updated_at = now() WHERE user_id = NEW.follower_id;
        ELSIF TG_OP = 'DELETE' THEN
            UPDATE user_stats SET followers_count = GREATEST(0, followers_count - 1), updated_at = now() WHERE user_id = OLD.following_id;
            UPDATE user_stats SET following_count = GREATEST(0, following_count - 1), updated_at = now() WHERE user_id = OLD.follower_id;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_follow_stats
    AFTER INSERT OR DELETE ON user_follows
    FOR EACH ROW
    EXECUTE FUNCTION update_user_stats();

CREATE OR REPLACE FUNCTION create_user_related_records()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO profiles (user_id, username, display_name) VALUES (NEW.user_id, NULL, NULL) ON CONFLICT (user_id) DO NOTHING;
    INSERT INTO user_preferences (user_id) VALUES (NEW.user_id) ON CONFLICT (user_id) DO NOTHING;
    INSERT INTO user_stats (user_id) VALUES (NEW.user_id) ON CONFLICT (user_id) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER create_user_defaults
    AFTER INSERT ON users
    FOR EACH ROW
    EXECUTE FUNCTION create_user_related_records();

-- Wallet service functions
CREATE OR REPLACE FUNCTION ensure_single_primary_wallet()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_primary = true THEN
        UPDATE wallet_links SET is_primary = false
        WHERE user_id = NEW.user_id AND wallet_id != NEW.wallet_id AND is_primary = true;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ensure_single_primary
    BEFORE INSERT OR UPDATE OF is_primary ON wallet_links
    FOR EACH ROW
    WHEN (NEW.is_primary = true)
    EXECUTE FUNCTION ensure_single_primary_wallet();

CREATE OR REPLACE FUNCTION log_wallet_activity()
RETURNS TRIGGER AS $$
DECLARE
    v_action varchar(50);
    v_metadata jsonb;
BEGIN
    IF TG_OP = 'INSERT' THEN
        v_action := 'linked';
        v_metadata := jsonb_build_object('address', NEW.address, 'chain_id', NEW.chain_id, 'is_primary', NEW.is_primary);
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.is_primary != NEW.is_primary AND NEW.is_primary = true THEN
            v_action := 'set_primary';
        ELSE
            v_action := 'updated';
        END IF;
        v_metadata := jsonb_build_object(
            'old', jsonb_build_object('is_primary', OLD.is_primary, 'label', OLD.label),
            'new', jsonb_build_object('is_primary', NEW.is_primary, 'label', NEW.label)
        );
    ELSIF TG_OP = 'DELETE' THEN
        v_action := 'unlinked';
        v_metadata := jsonb_build_object('address', OLD.address, 'chain_id', OLD.chain_id);
    END IF;

    INSERT INTO wallet_activity (wallet_id, user_id, action, metadata)
    VALUES (COALESCE(NEW.wallet_id, OLD.wallet_id), COALESCE(NEW.user_id, OLD.user_id), v_action, v_metadata);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER log_wallet_changes
    AFTER INSERT OR UPDATE OR DELETE ON wallet_links
    FOR EACH ROW
    EXECUTE FUNCTION log_wallet_activity();

-- ======================= COMMENTS =======================
COMMENT ON TABLE auth_nonces IS 'One-time nonces for SIWE authentication flow';
COMMENT ON TABLE sessions IS 'User sessions after successful SIWE verification';
COMMENT ON TABLE login_events IS 'Audit log of all authentication attempts';
COMMENT ON TABLE users IS 'Core user accounts';
COMMENT ON TABLE profiles IS 'User profile information';
COMMENT ON TABLE user_preferences IS 'User preferences and settings';
COMMENT ON TABLE user_stats IS 'Aggregated user statistics';
COMMENT ON TABLE user_follows IS 'User follow relationships';
COMMENT ON TABLE wallet_links IS 'User wallet connections and metadata';
COMMENT ON TABLE wallet_activity IS 'Audit log of wallet-related actions';
COMMENT ON TABLE wallet_verifications IS 'Wallet ownership verification records';

COMMIT;
