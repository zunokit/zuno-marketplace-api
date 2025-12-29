# Phase 4: Documentation & Scripts

**Priority**: P1
**Status**: Pending
**Effort**: 2 hours

## Context Links

- [Phase 3](./phase-03-application-config.md) - Code changes must be complete
- Current: `README.md`
- GitHub Issue: [#30](https://github.com/zunokit/zuno-marketplace-api/issues/30)

## Overview

Create comprehensive documentation and helper scripts for serverless development workflow. Update README with new setup instructions and troubleshooting guide.

**Goal**: Any developer can set up serverless environment in 10 minutes.

## Key Insights

1. Documentation must be clear for new developers
2. Scripts reduce error-prone manual setup
3. Troubleshooting section saves debugging time
4. Health check script validates configuration

## Requirements

### Functional Requirements
- FR1: Update README with serverless setup instructions
- FR2: Create health check script for connections
- FR3: Create troubleshooting guide
- FR4: Document migration from Docker to serverless
- FR5: Add architecture diagram

### Non-Functional Requirements
- NFR1: Instructions work for fresh clone
- NFR2: Scripts handle common errors gracefully
- NFR3: Clear section separation in README
- NFR4: Links to external provider docs

## Related Code Files

### Files to Modify
- `README.md` - Add serverless development section
- `DEVELOPMENT.md` - May exist, add serverless notes

### Files to Create
- `scripts/health-check.sh` - Connection validation script
- `docs/serverless-development.md` - Detailed guide
- `docs/troubleshooting.md` - Common issues and solutions

## Implementation Steps

### Step 1: Create Health Check Script

```bash
#!/bin/bash
# scripts/health-check.sh

set -e

echo "🔍 Zuno Marketplace - Infrastructure Health Check"
echo "================================================"
echo ""

# Load environment
if [ ! -f .env ]; then
    echo "❌ .env file not found"
    echo "   Run: ./scripts/setup-env.sh"
    exit 1
fi

source .env

echo "📋 Configuration:"
echo "   INFRA_MODE: ${INFRA_MODE:-not set}"
echo ""

# Check PostgreSQL
echo "🐘 PostgreSQL:"
if [ "$INFRA_MODE" = "serverless" ]; then
    if [ -n "$DATABASE_URL" ]; then
        echo "   ✅ DATABASE_URL set"
        # Extract host from URL
        HOST=$(echo $DATABASE_URL | awk -F'@' '{print $2}' | cut -d':' -f1)
        echo "   Host: $HOST"
    else
        echo "   ❌ DATABASE_URL not set"
    fi
else
    echo "   Host: ${POSTGRES_HOST:-localhost}:${POSTGRES_PORT:-5432}"
fi

# Check Redis
echo ""
echo "🔴 Redis:"
if [ "$INFRA_MODE" = "serverless" ]; then
    if [ -n "$REDIS_URL" ]; then
        echo "   ✅ REDIS_URL set"
        HOST=$(echo $REDIS_URL | awk -F'@' '{print $2}' | cut -d':' -f1)
        echo "   Host: $HOST"
    else
        echo "   ❌ REDIS_URL not set"
    fi
else
    echo "   Host: ${REDIS_HOST:-localhost}:${REDIS_PORT:-6379}"
fi

# Check RabbitMQ
echo ""
echo "🐰 RabbitMQ:"
if [ "$INFRA_MODE" = "serverless" ]; then
    if [ -n "$CLOUDAMQP_URL" ]; then
        echo "   ✅ CLOUDAMQP_URL set"
        HOST=$(echo $CLOUDAMQP_URL | awk -F'@' '{print $2}' | cut -d':' -f1)
        echo "   Host: $HOST"
    else
        echo "   ❌ CLOUDAMQP_URL not set"
    fi
else
    echo "   Host: ${RABBITMQ_HOST:-localhost}:${RABBITMQ_PORT:-5672}"
fi

echo ""
echo "================================================"
echo "✅ Configuration check complete"
echo ""
echo "Next steps:"
if [ "$INFRA_MODE" = "serverless" ]; then
    echo "   1. Run services directly: go run ./services/auth-service/cmd/main.go"
    echo "   2. Or use make build && ./bin/auth-service"
else
    echo "   1. Start Docker: docker compose up -d"
    echo "   2. Check status: docker compose ps"
fi
```

### Step 2: Create Serverless Development Guide

```markdown
<!-- docs/serverless-development.md -->

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
   - Name: `zuno-marketplace-dev`
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

| Provider | Service | Limit |
|----------|---------|-------|
| Supabase | PostgreSQL | 500MB storage, 1GB file storage |
| Upstash | Redis | 10K commands/day |
| CloudAMQP | RabbitMQ | 100 queues, 10K messages |

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

- Read [Development Guide](./DEVELOPMENT.md) for TDD workflow
- Check [System Architecture](./system-architecture.md) for design details
- See [Troubleshooting](./troubleshooting.md) for common issues
```

### Step 3: Create Troubleshooting Guide

```markdown
<!-- docs/troubleshooting.md -->

# Troubleshooting Guide

## Connection Issues

### "connection refused" to PostgreSQL

**Serverless Mode:**
- Check DATABASE_URL is correct
- Verify Supabase project is active (not paused)
- Check network connectivity
- Try: `psql $DATABASE_URL` to test connection

**Docker Mode:**
- Verify Docker is running: `docker ps`
- Check postgres container is up: `docker compose ps postgres`
- Restart: `docker compose restart postgres`

### Redis connection timeout

**Serverless Mode:**
- Verify REDIS_URL format: `redis://host:port`
- Check Upstash dashboard - database should be active
- Test with: `redis-cli -u $REDIS_URL PING`

**Docker Mode:**
- Check redis container: `docker compose ps redis`
- Verify port not in use: `lsof -i :6379`

### RabbitMQ connection failed

**Serverless Mode:**
- Verify CLOUDAMQP_URL format: `amqp://user:pass@host/vhost`
- Check CloudAMQP dashboard - instance should be running
- Note: Free instances pause after inactivity

**Docker Mode:**
- Check rabbitmq container: `docker compose ps rabbitmq`
- Management UI: http://localhost:15672 (guest/guest)

## Environment Issues

### ".env file not found"

```bash
# Run setup script
./scripts/setup-env.sh
```

### "INFRA_MODE not set"

Add to `.env`:
```bash
INFRA_MODE=serverless  # or docker
```

### Connection string working in dev but not production

- Check if using localhost URL in production
- Verify production environment has correct env vars
- Check firewall/security group rules

## Provider-Specific Issues

### Supabase

**"Project paused"**
- Free projects pause after 1 week of inactivity
- Click "Resume" in Supabase dashboard

**"Connection rate limit"**
- Free tier allows 60 concurrent connections
- Close idle connections in your code

### Upstash

**"Rate limit exceeded"**
- Free tier: 10K commands/day
- Check usage in Upstash dashboard
- Consider upgrading or switching to Docker mode

### CloudAMQP

**"Queue limit reached"**
- Free tier: 100 queues max
- Clean up unused queues in management UI
- Or upgrade to higher tier

**"Message limit reached"**
- Free tier: 10K messages
- Messages auto-delete after 28 days on free tier
- Monitor usage in dashboard

## Performance Issues

### Slow response times

**Serverless Mode:**
- Check provider region selection (closer = faster)
- Upstash HTTP has latency vs native Redis
- Consider connection pooling

**Docker Mode:**
- Check Docker resource limits
- Increase memory in Docker Desktop settings

### High memory usage

**Serverless Mode:**
- Should be minimal (no Docker overhead)
- Check for connection leaks
- Profile with: `pprof`

**Docker Mode:**
- Check container stats: `docker stats`
- Limit container memory in docker-compose.yml

## Getting Help

1. Check [Serverless Development Guide](./serverless-development.md)
2. Search existing [GitHub Issues](https://github.com/zunokit/zuno-marketplace-api/issues)
3. Create new issue with:
   - INFRA_MODE setting
   - Full error message
   - Output of `./scripts/health-check.sh`
```

### Step 4: Update README

Add new section after "Quick Start":

```markdown
## 🌩️ Serverless Development (NEW - Recommended)

**No Docker required!** Use free-tier cloud services for development.

### Benefits
- ✅ Zero local resource usage
- ✅ $0/month (free tiers)
- ✅ Quick 10-minute setup

### Quick Setup

```bash
# 1. Create free accounts
#    - Supabase (PostgreSQL): https://supabase.com
#    - Upstash (Redis): https://upstash.com
#    - CloudAMQP (RabbitMQ): https://www.cloudamqp.com

# 2. Configure environment
./scripts/setup-env.sh  # Select option 1 (serverless)

# 3. Add connection strings to .env
#    DATABASE_URL=postgresql://...
#    REDIS_URL=redis://...
#    CLOUDAMQP_URL=amqp://...

# 4. Verify setup
./scripts/health-check.sh

# 5. Run services (no Docker needed!)
go run ./services/auth-service/cmd/main.go
```

**Full guide:** [Serverless Development Guide](./docs/serverless-development.md)

### Switching Modes

```bash
# Docker mode (current setup)
INFRA_MODE=docker
docker compose up -d

# Serverless mode (new)
INFRA_MODE=serverless
# No Docker needed!
```

---

## 🐳 Docker Development (Current)

*Existing Docker setup remains fully functional*

See "Quick Start" section above for Docker instructions.
```

### Step 5: Update DEVELOPMENT.md

Add section about serverless workflow:

```markdown
## Development Environment Modes

### Serverless Mode (Recommended for Development)

- Zero local infrastructure overhead
- Free-tier cloud services
- Faster iteration (no container rebuilds)
- Best for feature development

### Docker Mode

- Full production parity
- Offline development
- Integration testing
- Best for final testing before deployment

Switch modes via `INFRA_MODE` in `.env`.
```

## Todo List

- [ ] Create `scripts/health-check.sh`
- [ ] Create `docs/serverless-development.md`
- [ ] Create `docs/troubleshooting.md`
- [ ] Update `README.md` with serverless section
- [ ] Update `DEVELOPMENT.md` with mode notes
- [ ] Test all scripts on fresh clone
- [ ] Verify all links work

## Success Criteria

- [ ] New developer can set up in <10 minutes
- [ ] Health check script validates all connections
- [ ] Troubleshooting covers common issues
- [ ] README has clear mode comparison

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Outdated provider docs | Medium | Low | Link to official docs |
| Script permissions issue | Low | Low | Document chmod +x |
| Broken links | Low | Low | Verify all links |

## Security Considerations

- **No secrets in docs**: All examples use placeholders
- **Provider links**: Use official HTTPS URLs
- **Warning about free tiers**: Note limitations clearly

## Next Steps

- Proceed to [Phase 5: Testing & Validation](./phase-05-testing-validation.md)
- Validate complete setup workflow

## Unresolved Questions

- Should we include video walkthrough?
- Add diagrams for architecture comparison?
- Include performance benchmarks?
