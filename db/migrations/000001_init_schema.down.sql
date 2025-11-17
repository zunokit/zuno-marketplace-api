BEGIN;

-- ======================= DROP TRIGGERS =======================

DROP TRIGGER IF EXISTS log_wallet_changes ON wallet_links;
DROP TRIGGER IF EXISTS ensure_single_primary ON wallet_links;
DROP TRIGGER IF EXISTS create_user_defaults ON users;
DROP TRIGGER IF EXISTS update_follow_stats ON user_follows;

-- ======================= DROP FUNCTIONS =======================

DROP FUNCTION IF EXISTS log_wallet_activity();
DROP FUNCTION IF EXISTS ensure_single_primary_wallet();
DROP FUNCTION IF EXISTS create_user_related_records();
DROP FUNCTION IF EXISTS update_user_stats();
DROP FUNCTION IF EXISTS try_use_nonce(text, text, text, text);
DROP FUNCTION IF EXISTS cleanup_expired_sessions();
DROP FUNCTION IF EXISTS cleanup_old_login_events(integer);
DROP FUNCTION IF EXISTS cleanup_expired_nonces();

-- ======================= DROP WALLET SERVICE TABLES =======================

DROP TABLE IF EXISTS wallet_verifications;
DROP TABLE IF EXISTS wallet_activity;
DROP TABLE IF EXISTS wallet_links;

-- ======================= DROP USER SERVICE TABLES =======================

DROP TABLE IF EXISTS user_follows;
DROP TABLE IF EXISTS user_stats;
DROP TABLE IF EXISTS user_preferences;
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS users;

-- ======================= DROP AUTH SERVICE TABLES =======================

DROP TABLE IF EXISTS login_events;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS auth_nonces;

-- ======================= DROP EXTENSIONS =======================

DROP EXTENSION IF EXISTS pgcrypto;
DROP EXTENSION IF EXISTS "uuid-ossp";

COMMIT;
