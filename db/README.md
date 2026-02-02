# Database Migrations

This directory contains database migrations managed by [golang-migrate](https://github.com/golang-migrate/migrate).

## Migration Files

All migrations are in `db/migrations/`:
- `000001_init_schema.up.sql` - Initial database schema (all services)

## Usage

### Run Migrations
```bash
make migrate
```

### Check Current Version
```bash
make migrate-status
```

### Create New Migration
```bash
make migrate-create NAME=add_new_table
```


## Migration Structure

The initial migration (`000001_init_schema.up.sql`) contains schemas for all three services:
- **Auth Service**: `auth_nonces`, `sessions`, `login_events`
- **User Service**: `users`, `profiles`, `user_preferences`, `user_stats`, `user_follows`
- **Wallet Service**: `wallet_links`, `wallet_activity`, `wallet_verifications`

## Environment Variables

Default database connection (can override in Makefile or environment):
```bash
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=nft_marketplace
```

## Notes

- ⚠️ Down migrations are NOT created (no rollback support by design)
- All services share the same PostgreSQL database
- Schema version is tracked in `schema_migrations` table
