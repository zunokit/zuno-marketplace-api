---
title: "Phase 03: Environment Setup"
description: "Update environment files for Air hot-reload with 4xxx ports"
status: pending
priority: P1
effort: 1h
branch: feature/hybird-serverless-and-servers-infrastructure
tags: [environment, ports, serverless, neon, upstash, cloudamqp]
created: 2026-01-28
---

# Phase 03: Environment Setup

## Context Links

- [Main Plan](plan.md)
- [Hybrid Workflow Research](research/researcher-02-hybrid-workflow.md)
- [Phase 02: Air Configuration](phase-02-air-configuration.md)

## Overview

**Priority:** P1 (required for Air workflow)
**Current Status:** Pending
**Description:** Update `.env.development.example` with 4xxx ports and serverless infrastructure URLs (Neon/Upstash/CloudAMQP).

**Key Changes:**
- Change ports from 5xxx to 4xxx range
- Add serverless infrastructure URLs
- Keep Docker infra fallback for flexibility

---

## Key Insights

From [Hybrid Workflow Research](research/researcher-02-hybrid-workflow.md):
- **Dual-mode support:** Serverless (primary) + Docker (fallback)
- **Port conflicts:** 4xxx range avoids conflicts with Docker infra (5433, 6379, 5672)
- **Environment precedence:** CLI args > `.env.local` > `.env` > defaults
- **Validation scripts:** Verify required vars before startup

**Serverless Services:**
- **Neon:** PostgreSQL (free tier, branching support)
- **Upstash:** Redis (free tier, edge caching)
- **CloudAMQP:** RabbitMQ (free tier available)

---

## Requirements

### Functional Requirements
1. Update `.env.development.example` with 4xxx ports
2. Add `INFRA_MODE` flag (serverless/docker)
3. Document Neon/Upstash/CloudAMQP setup
4. Keep Docker fallback config

### Non-Functional Requirements
- Backward compatible with Docker mode
- Clear documentation for setup
- No production changes (`.env.production.example` unchanged)

---

## Architecture

### Environment Modes

**Serverless Mode (New - Development):**
```
Local Services (Air)
├── auth-service:    localhost:4001
├── user-service:    localhost:4002
├── wallet-service:  localhost:4003
└── graphql-gateway: localhost:4080

Cloud Infrastructure
├── DATABASE_URL (Neon)
├── REDIS_URL (Upstash)
└── CLOUDAMQP_URL (CloudAMQP)
```

**Docker Mode (Existing - Production-like):**
```
Docker Services
├── All services in containers
├── Infra in containers (postgres, redis, rabbitmq)
└── Original port mapping
```

---

## Related Code Files

### Files to MODIFY
- `.env.development.example` - Update for Air + serverless
- `scripts/setup-env.sh` - Update prompts (if exists)

### Files to KEEP (unchanged)
- `.env.production.example` - No changes
- `docker-compose.yml` - No changes

---

## Implementation Steps

### Step 1: Read Current .env.development.example (5 min)

```bash
# Read current file
cat .env.development.example

# Note current port configuration
# Note existing infrastructure URLs
```

**Success:** Understand current configuration

**Rollback:** None (read-only)

---

### Step 2: Create New .env.development.example (20 min)

Create new `.env.development.example`:

```bash
# ============================================================
# Development Environment (Air Hot-Reload + Serverless Infra)
# ============================================================
# Setup: cp .env.development.example .env.development
# Then edit .env.development with your actual credentials

# ============================================================
# Infrastructure Mode
# ============================================================
INFRA_MODE=serverless  # Options: serverless | docker
ENVIRONMENT=development

# ============================================================
# Local Service Ports (4xxx range for Air)
# ============================================================
# gRPC Services (internal communication)
AUTH_GRPC_PORT=:4001
USER_GRPC_PORT=:4002
WALLET_GRPC_PORT=:4003

# HTTP Gateway (external API)
GATEWAY_HTTP_ADDR=:4080

# ============================================================
# Service URLs (localhost for local Air services)
# ============================================================
USER_SERVICE_URL=localhost:4002
WALLET_SERVICE_URL=localhost:4003

# ============================================================
# Authentication
# ============================================================
JWT_SECRET=dev-secret-key-change-in-production-32chars
REFRESH_SECRET=dev-refresh-secret-change-in-production-32chars
JWT_EXPIRATION=15m
REFRESH_EXPIRATION=7d

# ============================================================
# Database (Neon - Serverless PostgreSQL)
# ============================================================
# Get free database at: https://console.neon.tech/
DATABASE_URL=postgres://postgres:[YOUR_PASSWORD]@[PROJECT_ID].aws.neon.tech/neondb?sslmode=require

# Fallback: Docker PostgreSQL (if INFRA_MODE=docker)
POSTGRES_HOST=localhost
POSTGRES_PORT=5433
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DATABASE=nft_marketplace

# ============================================================
# Cache (Upstash - Serverless Redis)
# ============================================================
# Get free Redis at: https://console.upstash.com/
REDIS_URL=redis://default:[YOUR_PASSWORD]@xxx.upstash.io:6379

# Fallback: Docker Redis (if INFRA_MODE=docker)
REDIS_HOST=localhost
REDIS_PORT=6379

# ============================================================
# Message Queue (CloudAMQP - Serverless RabbitMQ)
# ============================================================
# Get free instance at: https://www.cloudamqp.com/
CLOUDAMQP_URL=amqp://[USER]:[PASSWORD]@[HOST]/[VHOST]

# Fallback: Docker RabbitMQ (if INFRA_MODE=docker)
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_VHOST=/

# ============================================================
# External Services
# ============================================================

# Wallet/NFT (Ethereum/Polygon)
ETHEREUM_RPC_URL=https://eth-sepolia.g.alchemy.com/v2/[YOUR_API_KEY]
POLYGON_RPC_URL=https://polygon-amoy.g.alchemy.com/v2/[YOUR_API_KEY]
WALLET_PRIVATE_KEY=0x[YOUR_PRIVATE_KEY]  # Only for dev/testing

# IPFS (NFT metadata)
IPFS_API_URL=https://api.pinata.cloud/api/pinning/pinFileToIPFS
IPFS_API_KEY=[YOUR_PINATA_API_KEY]
IPFS_SECRET_KEY=[YOUR_PINATA_SECRET]

# ============================================================
# Feature Flags
# ============================================================
ENABLE_SWAGGER=true
ENABLE_METRICS=false
ENABLE_TRACING=false

# ============================================================
# Logging
# ============================================================
LOG_LEVEL=debug
LOG_FORMAT=console  # Options: console | json
```

**Success:** New file created with all configuration

**Rollback:** `git checkout .env.development.example`

---

### Step 3: Validate .env.development.example (10 min)

```bash
# Check for syntax errors (no unexpected characters)
grep -n '[^a-zA-Z0-9=:._/@% -]' .env.development.example

# Verify all placeholders documented
grep -n '\[YOUR_' .env.development.example | wc -l

# Check port range (all should be 4xxx except infra)
grep -n 'PORT=' .env.development.example
```

**Success:** No syntax errors, all placeholders clear

**Rollback:** Fix syntax issues in file

---

### Step 4: Update setup-env.sh Script (15 min)

If `scripts/setup-env.sh` exists, update it:

```bash
#!/bin/bash
# Setup .env file from example with interactive prompts

set -e

ENV_FILE=".env.development"
ENV_EXAMPLE=".env.development.example"

echo "============================================================"
echo "  Zuno NFT Marketplace - Environment Setup"
echo "============================================================"
echo

# Check if example exists
if [ ! -f "$ENV_EXAMPLE" ]; then
  echo "Error: $ENV_EXAMPLE not found"
  exit 1
fi

# Ask for infrastructure mode
echo "Select infrastructure mode:"
echo "  1) Serverless (Neon/Upstash/CloudAMQP) - Recommended for dev"
echo "  2) Docker (local containers)"
read -p "Choose [1/2]: " infra_choice

if [ "$infra_choice" = "2" ]; then
  infra_mode="docker"
else
  infra_mode="serverless"
fi

echo
echo "Infrastructure mode: $infra_mode"
echo

# Copy example
cp "$ENV_EXAMPLE" "$ENV_FILE"

# Update INFRA_MODE
if [[ "$OSTYPE" == "darwin"* ]]; then
  sed -i '' "s/INFRA_MODE=.*/INFRA_MODE=$infra_mode/" "$ENV_FILE"
else
  sed -i "s/INFRA_MODE=.*/INFRA_MODE=$infra_mode/" "$ENV_FILE"
fi

if [ "$infra_mode" = "serverless" ]; then
  echo
  echo "------------------------------------------------------------"
  echo "  Serverless Setup Required"
  echo "------------------------------------------------------------"
  echo
  echo "You need to set up free accounts:"
  echo "  • Neon PostgreSQL: https://console.neon.tech/"
  echo "  • Upstash Redis:   https://console.upstash.com/"
  echo "  • CloudAMQP:       https://www.cloudamqp.com/"
  echo
  echo "After setup, edit $ENV_FILE and replace:"
  echo "  • [YOUR_PASSWORD] in DATABASE_URL"
  echo "  • [PROJECT_ID] in DATABASE_URL"
  echo "  • [YOUR_PASSWORD] in REDIS_URL"
  echo "  • [USER], [PASSWORD], [HOST] in CLOUDAMQP_URL"
  echo
else
  echo
  echo "Docker mode: Make sure to run 'make infra-up' before 'make dev-air'"
fi

echo
echo "✓ Environment file created: $ENV_FILE"
echo
echo "Next steps:"
echo "  1. Edit $ENV_FILE with your credentials"
echo "  2. Run: make dev-air"
echo
```

If script doesn't exist, create it.

**Success:** Script guides users through setup

**Rollback:** Revert script changes

---

### Step 5: Document Serverless Setup (10 min)

Create `docs/development-setup.md`:

```markdown
# Development Environment Setup

## Quick Start (Serverless Mode)

### 1. Set Up Free Cloud Accounts

**Neon PostgreSQL**
1. Go to https://console.neon.tech/
2. Create free account
3. Create new project
4. Copy connection string: `postgres://postgres:[PASSWORD]@[PROJECT-ID].aws.neon.tech/neondb?sslmode=require`

**Upstash Redis**
1. Go to https://console.upstash.com/
2. Create free account
3. Create new Redis database
4. Copy connection string: `redis://default:[PASSWORD]@xxx.upstash.io:6379`

**CloudAMQP**
1. Go to https://www.cloudamqp.com/
2. Create free account
3. Create new instance (Little Lemur tier)
4. Copy connection string: `amqp://xxx:xxx@xxx.rmq.cloudamqp.com/xxx`

### 2. Configure Environment

```bash
# Copy example
cp .env.development.example .env.development

# Edit with your credentials
nano .env.development
```

Replace:
- `[YOUR_PASSWORD]`, `[PROJECT_ID]` in `DATABASE_URL`
- `[YOUR_PASSWORD]` in `REDIS_URL`
- `[USER]`, `[PASSWORD]`, `[HOST]`, `[VHOST]` in `CLOUDAMQP_URL`

### 3. Start Development

```bash
make dev-air
```

This starts all 4 services with Air hot-reload.

## Docker Mode (Alternative)

If you prefer local containers:

```bash
# Set INFRA_MODE=docker in .env.development
# Then run:
make infra-up  # Start postgres/redis/rabbitmq
make dev-air   # Start services with Air
```

## Ports

| Service | Port | Protocol |
|---------|------|----------|
| auth-service | 4001 | gRPC |
| user-service | 4002 | gRPC |
| wallet-service | 4003 | gRPC |
| graphql-gateway | 4080 | HTTP |
```

**Success:** Setup guide created

**Rollback:** Delete documentation file

---

## Todo List

- [ ] Read current `.env.development.example`
- [ ] Create new `.env.development.example` with 4xxx ports
- [ ] Add `INFRA_MODE` flag
- [ ] Add serverless URLs (Neon/Upstash/CloudAMQP)
- [ ] Keep Docker fallback config
- [ ] Validate file for syntax errors
- [ ] Update/create `scripts/setup-env.sh`
- [ ] Create `docs/development-setup.md`

---

## Success Criteria

✅ `.env.development.example` uses 4xxx ports
✅ `INFRA_MODE` flag added (serverless/docker)
✅ Serverless URLs documented with placeholders
✅ Docker fallback config preserved
✅ Setup script guides users
✅ No production changes (`.env.production.example` unchanged)

**Validation Commands:**
```bash
# Check ports
grep "PORT=" .env.development.example | grep -E "400[0-9]|4080"

# Check serverless vars
grep -E "DATABASE_URL|REDIS_URL|CLOUDAMQP_URL" .env.development.example

# Check INFRA_MODE
grep "INFRA_MODE" .env.development.example
```

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Port conflicts with Docker | Low | Medium | 4xxx range avoids Docker infra ports |
| Serverless service downtime | Medium | Medium | Docker fallback available |
| Users don't have accounts | High | Low | Clear signup links in docs |
| .env file not in .gitignore | Low | High | Verify .gitignore |

---

## Security Considerations

**CRITICAL:** Never commit actual `.env` files with credentials
- `.env.development` must be in `.gitignore`
- Use `.env.development.example` with placeholders only
- Document security best practices in setup guide

---

## Next Steps

**After Phase 03:**
- Phase 04: Scripts Creation (use env vars)
- Phase 05: Makefile Enhancement (respect INFRA_MODE)

**Dependencies:**
- Phase 02 should be completed (Air configs reference ports)

**Follow-up Tasks:**
- Sign up for Neon/Upstash/CloudAMQP accounts
- Test serverless connections

---

## Rollback Plan

**Full Rollback:**
```bash
# Restore original .env
git checkout .env.development.example
git checkout scripts/setup-env.sh
rm docs/development-setup.md
```

**Partial Rollback:**
```bash
# Keep ports, revert serverless URLs
git checkout .env.development.example
# Then manually update ports only
```
