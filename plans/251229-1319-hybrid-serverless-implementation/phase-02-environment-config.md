# Phase 2: Environment Configuration

**Priority**: P1
**Status**: Done (Completed 2025-12-29 / 251229)
**Effort**: 1 hour
**Code Review**: `plans/reports/code-reviewer-251229-2149-phase2-env-config.md`

## Context Links

- [Phase 1](./phase-01-account-setup.md) - Complete this first to get connection strings
- Current: `services/*/internal/config/config.go`
- Template: `.env.example` (may need privacy approval to read)

## Overview

Update environment configuration files to support both serverless (dev) and Docker (production/fallback) environments. Create `.env.development` template for serverless setup.

**Goal**: Developer can switch between Docker and serverless via single ENV variable.

## Key Insights

1. Current code uses individual env vars (POSTGRES_HOST, POSTGRES_PORT, etc.)
2. Better to use full connection URLs (DATABASE_URL, REDIS_URL, CLOUDAMQP_URL)
3. Support both modes: `INFRA_MODE=docker` or `INFRA_MODE=serverless`
4. Need minimal code changes - mostly environment variable naming

## Requirements

### Functional Requirements
- FR1: Create `.env.development` template for serverless
- FR2: Create `.env.production` template for AWS/Docker
- FR3: Add INFRA_MODE switch to config
- FR4: Support URL-based connections (DATABASE_URL, etc.)
- FR5: Update `.gitignore` for new env files

### Non-Functional Requirements
- NFR1: Existing Docker workflow must continue working
- NFR2: Zero code changes in this phase (config only)
- NFR3: Backward compatible with existing .env files
- NFR4: Clear documentation for switching modes

## Architecture

```
INFRA_MODE environment variable
    │
    ├── docker (current default)
    │   ├── POSTGRES_HOST=localhost
    │   ├── POSTGRES_PORT=5432
    │   └── ... individual vars
    │
    └── serverless (new)
        ├── DATABASE_URL=postgresql://...
        ├── REDIS_URL=redis://...
        └── CLOUDAMQP_URL=amqp://...
```

## Related Code Files

### Files to Modify
- `services/auth-service/internal/config/config.go` - Add URL parsing logic
- `services/user-service/internal/config/config.go` - Add URL parsing logic
- `services/wallet-service/internal/config/config.go` - Add URL parsing logic
- `services/graphql-gateway/internal/config/config.go` - Add URL parsing logic
- `.gitignore` - Add `.env.development`, `.env.production`

### Files to Create
- `.env.development.example` - Serverless template
- `.env.production.example` - AWS template
- `scripts/setup-env.sh` - Environment setup helper

## Implementation Steps

### Step 1: Create .env.development.example

```bash
# .env.development.example
# Infrastructure Mode: docker | serverless
INFRA_MODE=serverless

# Service Ports (unchanged)
AUTH_GRPC_PORT=:50051
USER_GRPC_PORT=:50052
WALLET_GRPC_PORT=:50053
GATEWAY_HTTP_ADDR=:8080

# Database (Supabase)
# Get from: Supabase Project → Settings → Database → Connection string
DATABASE_URL=postgresql://postgres:[YOUR-PASSWORD]@db.xxx.supabase.co:5432/postgres

# Redis (Upstash)
# Get from: Upstash Dashboard → Database → Details → REST API
REDIS_URL=redis://default:[YOUR-PASSWORD]@xxx.upstash.io:6379

# RabbitMQ (CloudAMQP)
# Get from: CloudAMQP Dashboard → Instance → Details
CLOUDAMQP_URL=amqp://xxx:xxx@xxx.rmq.cloudamqp.com/xxx

# JWT Secrets (change in production!)
JWT_SECRET=your-256-bit-secret-key-here
REFRESH_SECRET=your-256-bit-refresh-secret-key-here

# Service URLs (for inter-service communication)
USER_SERVICE_URL=localhost:50052
WALLET_SERVICE_URL=localhost:50053
```

### Step 2: Create .env.production.example

```bash
# .env.production.example
# Infrastructure Mode: docker | serverless
INFRA_MODE=docker

# Service Ports
AUTH_GRPC_PORT=:50051
USER_GRPC_PORT=:50052
WALLET_GRPC_PORT=:50053
GATEWAY_HTTP_ADDR=:8080

# Database (AWS RDS or Docker)
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DATABASE=nft_marketplace

# Redis (AWS ElastiCache or Docker)
REDIS_HOST=redis
REDIS_PORT=6379

# RabbitMQ (AWS MQ or Docker)
RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_EXCHANGE=nft_events

# JWT Secrets
JWT_SECRET=your-production-jwt-secret
REFRESH_SECRET=your-production-refresh-secret
```

### Step 3: Update .gitignore

```gitignore
# Environment files
.env
.env.development
.env.production
.env.local

# Connection details (local only)
*-connections.txt
supabase-upstash-cloudamqp-connections.txt
```

### Step 4: Create Setup Script

```bash
#!/bin/bash
# scripts/setup-env.sh

set -e

echo "🔧 Zuno Marketplace - Environment Setup"
echo ""

# Check if .env exists
if [ -f .env ]; then
    echo "⚠️  .env file already exists"
    read -p "Overwrite? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted"
        exit 1
    fi
    rm .env
fi

# Ask for infrastructure mode
echo "Select infrastructure mode:"
echo "  1) serverless - Supabase, Upstash, CloudAMQP (free, no Docker)"
echo "  2) docker - Local containers (resource intensive)"
read -p "Choice [1-2]: " mode_choice

case $mode_choice in
    1)
        echo "📋 Copying .env.development.example → .env"
        cp .env.development.example .env
        echo "✅ Serverless environment configured"
        echo ""
        echo "⚠️  IMPORTANT: Edit .env and add your connection strings:"
        echo "   - DATABASE_URL (from Supabase)"
        echo "   - REDIS_URL (from Upstash)"
        echo "   - CLOUDAMQP_URL (from CloudAMQP)"
        ;;
    2)
        echo "📋 Copying .env.production.example → .env"
        cp .env.production.example .env
        echo "✅ Docker environment configured"
        echo ""
        echo "🐳 Start services with: docker compose up -d"
        ;;
    *)
        echo "❌ Invalid choice"
        exit 1
        ;;
esac

echo ""
echo "✨ Setup complete! Edit .env as needed, then run services."
```

### Step 5: Update Config Code (Example for auth-service)

**Note**: This involves code changes - detailed in Phase 3.

```go
// services/auth-service/internal/config/config.go

type DatabaseConfig struct {
    Mode       string // "docker" or "serverless"
    Host       string
    Port       string
    User       string
    Password   string
    Database   string
    SSLMode    string
    URL        string // Full connection URL for serverless
}

func Load() *Config {
    mode := env.GetString("INFRA_MODE", "docker")

    dbConfig := DatabaseConfig{Mode: mode}

    if mode == "serverless" {
        // Use full URL
        dbConfig.URL = env.GetString("DATABASE_URL", "")
    } else {
        // Use individual vars (current behavior)
        dbConfig.Host = env.GetString("POSTGRES_HOST", "localhost")
        dbConfig.Port = env.GetString("POSTGRES_PORT", "5432")
        // ... etc
    }

    return &Config{
        Database: dbConfig,
        // ... rest of config
    }
}

func (c *DatabaseConfig) GetDSN() string {
    if c.Mode == "serverless" && c.URL != "" {
        return c.URL
    }
    // Fall back to existing behavior
    return "host=" + c.Host + " port=" + c.Port + " user=" + c.User +
        " password=" + c.Password + " dbname=" + c.Database + " sslmode=" + c.SSLMode
}
```

## Todo List

- [x] Create `.env.development.example` with serverless vars
- [x] Create `.env.production.example` with Docker/AWS vars
- [x] Update `.gitignore` for new env files
- [x] Create `scripts/setup-env.sh` setup script
- [x] Make setup script executable (chmod +x)
- [x] Test script with both modes
- [x] Document setup process in README
- [x] **Commit changes to git**

## Success Criteria

- [x] Developer can run `./scripts/setup-env.sh` to create .env
- [x] .env.development.example has all required vars documented
- [x] .env.production.example has all required vars documented
- [x] .gitignore prevents committing actual .env files
- [x] Setup script works for both modes
- [x] Changes committed to git

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Existing .env gets overwritten | Medium | Medium | Script confirms before overwrite |
| Wrong mode selected | Low | Low | Easy to re-run script |
| Connection string format incompatibility | Low | Medium | Test both modes before commit |

## Code Review Summary (2025-12-29)

**Grade**: B+ (Ready with Minor Improvements)
**Report**: `plans/reports/code-reviewer-251229-2149-phase2-env-config.md`

### Findings

| Severity | Count | Items |
|----------|-------|-------|
| Critical | 0 | - |
| High | 2 | Files not committed to git, Missing connection validation |
| Medium | 3 | .gitignore duplicates, Emoji compatibility, .env.example fate |
| Low | 2 | Dry-run mode, Value preservation on switch |

### Action Items

1. **[BLOCKING]** Commit changes:
   ```bash
   chmod +x scripts/setup-env.sh
   git add .env.development.example .env.production.example scripts/ .gitignore README.md
   git commit -m "feat(infra): complete Phase 2 - Environment Configuration"
   ```

2. **[Recommended]** Fix .gitignore duplicates (remove lines 34-46)

3. **[Defer to Phase 5]** Connection string validation

## Security Considerations

- **Example files only**: Never commit actual .env files
- **Documentation**: Clearly mark which vars need real values
- **Validation**: Add script to validate required vars set

## Next Steps

- Proceed to [Phase 3: Application Configuration](./phase-03-application-config.md)
- Update config code to support URL-based connections

## Unresolved Questions

- Should we validate connection strings during setup?
- How to handle migration from existing Docker .env files?
- Should we include environment variable documentation in example files?

---

**Files to modify** (Privacy Block Warning):
- `.env.example` - May require approval to read current file
- `.gitignore` - Safe to modify
