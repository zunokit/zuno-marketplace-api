# Serverless Development Guide

## Overview

This guide covers setting up a serverless development environment using free-tier cloud services. No Docker required - all infrastructure runs in the cloud.

## Benefits

- ✅ **Zero local resource usage** - No laptop lag
- ✅ **Free for development** - $0/month with generous free tiers
- ✅ **Quick setup** - 10 minutes from account to running code
- ✅ **Team collaboration** - Shared dev environment

## Prerequisites

- Go 1.21+
- Git
- GitHub/GitLab account (for OAuth with providers)

## Quick Start

### 1. Create Provider Accounts

**Supabase (PostgreSQL)**

1. Go to https://supabase.com
2. Sign up with GitHub
3. Create new project:
   - Name: `zuno-api`
   - Region: Select closest to you
   - Plan: Free
4. Save connection string from Project → Settings → Database

**Upstash (Redis)**

1. Go to https://upstash.com
2. Sign up with GitHub
3. Create database:
   - Name: `zuno-marketplace-redis`
   - Region: Select closest to you
   - Plan: Free
4. Save REST API URL and Token

**CloudAMQP (RabbitMQ)**

1. Go to https://www.cloudamqp.com
2. Sign up with GitHub
3. Create instance:
   - Name: `zuno-marketplace-rabbitmq`
   - Plan: Little Lemur (FREE)
   - Region: Select closest to you
4. Save AMQP URL

### 2. Configure Environment

```bash
# Run setup script
./scripts/setup-env.sh

# Select option 1 (serverless)

# Edit .env and add your connection strings
nano .env
```

Add your connection strings:

```bash
DATABASE_URL=postgresql://postgres:[PASSWORD]@db.xxx.supabase.co:5432/postgres
REDIS_URL=redis://default:[PASSWORD]@xxx.upstash.io:6379
CLOUDAMQP_URL=amqp://xxx:xxx@xxx.rmq.cloudamqp.com/xxx
```

### 3. Verify Configuration

```bash
# Run health check
./scripts/health-check.sh

# Expected output:
# ✅ DATABASE_URL set
# ✅ REDIS_URL set
# ✅ CLOUDAMQP_URL set
```

### 4. Run Services

```bash
# Option 1: Direct run
go run ./services/auth-service/cmd/main.go

# Option 2: Build and run
make build
./bin/auth-service

# Option 3: Individual service
cd services/auth-service
go run cmd/main.go
```

## Switching Between Modes

### Docker Mode (Current Production Setup)

```bash
# Set INFRA_MODE in .env
INFRA_MODE=docker

# Start Docker services
docker compose up -d

# Run application
go run ./services/auth-service/cmd/main.go
```

### Serverless Mode (New Development Setup)

```bash
# Set INFRA_MODE in .env
INFRA_MODE=serverless

# No Docker needed!
# Just run application
go run ./services/auth-service/cmd/main.go
```

## Provider Limits (Free Tiers)

| Provider  | Service    | Limit                           |
| --------- | ---------- | ------------------------------- |
| Supabase  | PostgreSQL | 500MB storage, 1GB file storage |
| Upstash   | Redis      | 10K commands/day                |
| CloudAMQP | RabbitMQ   | 100 queues, 10K messages        |

These limits are sufficient for development. If exceeded, consider upgrading or switching back to Docker mode.

## Troubleshooting

See [Troubleshooting Guide](./troubleshooting.md) for common issues.

## Migration from Docker

If you have existing Docker setup with data:

1. **Export existing data** (if needed):

   ```bash
   docker exec postgres pg_dump -U postgres nft_marketplace > backup.sql
   ```

2. **Switch to serverless mode**:

   ```bash
   ./scripts/setup-env.sh  # Select option 1
   ```

3. **Import data** (if needed):
   ```bash
   psql $DATABASE_URL < backup.sql
   ```

## Next Steps

- Read [Development Guide](../DEVELOPMENT.md) for TDD workflow
- Check [System Architecture](./system-architecture.md) for design details
- See [Troubleshooting](./troubleshooting.md) for common issues
