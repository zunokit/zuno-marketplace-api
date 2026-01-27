---
title: "Phase 07: Documentation"
description: "Update README and create development guides"
status: pending
priority: P1
effort: 30m
branch: feature/hybird-serverless-and-servers-infrastructure
tags: [documentation, readme, guides]
created: 2026-01-28
---

# Phase 07: Documentation

## Context Links

- [Main Plan](plan.md)
- [Existing README](../../README.md) (if exists)
- [Phase 03: Environment Setup](phase-03-environment-setup.md)
- [Phase 06: Testing & Validation](phase-06-testing-validation.md)

## Overview

**Priority:** P1 (developer onboarding)
**Current Status:** Pending
**Description:** Update project documentation with new Air hot-reload workflow, including README.md, quick-start guides, and troubleshooting.

**Documentation to Update:**
1. **README.md** - Main project documentation
2. **DEVELOPMENT.md** - Development workflow guide (create new)
3. **TROUBLESHOOTING.md** - Common issues and solutions (create new)

---

## Key Insights

From [Phase 06](phase-06-testing-validation.md):
- **Dual-mode workflow:** Docker (production) vs Air (development)
- **Quick onboarding:** Developers need < 5 minutes setup
- **Clear separation:** Production workflow unchanged
- **Troubleshooting:** Common issues need documented solutions

**Documentation Principles:**
- Start with quick-start (most important)
- Separate Docker vs Air modes clearly
- Keep production workflow prominent
- Document common pitfalls

---

## Requirements

### Functional Requirements
1. Update README.md with dual-mode workflow
2. Create DEVELOPMENT.md with detailed guide
3. Create TROUBLESHOOTING.md with common issues
4. Document serverless setup (Neon/Upstash/CloudAMQP)
5. Keep production Docker workflow visible

### Non-Functional Requirements
- Clear, concise instructions
- Code blocks with working commands
- Links to external resources
- Platform-specific notes (Windows/macOS/Linux)

---

## Architecture

### Documentation Structure

```
docs/
├── development-setup.md    ← From Phase 03
├── air-hot-reload-guide.md ← NEW (this phase)
└── serverless-setup.md     ← NEW (this phase)

project-root/
├── README.md               ← UPDATE (this phase)
├── DEVELOPMENT.md          ← CREATE (this phase)
└── TROUBLESHOOTING.md      ← CREATE (this phase)
```

---

## Related Code Files

### Files to CREATE
- `DEVELOPMENT.md`
- `TROUBLESHOOTING.md`
- `docs/air-hot-reload-guide.md`
- `docs/serverless-setup.md`

### Files to UPDATE
- `README.md`

### Files to REFERENCE
- `.env.development.example` (from Phase 03)
- `Makefile` (from Phase 05)

---

## Implementation Steps

### Step 1: Update README.md (15 min)

**Current README Structure:**
```markdown
# Zuno NFT Marketplace API

## Quick Start
make dev
```

**Updated README Structure:**
```markdown
# Zuno NFT Marketplace API

[Existing badges and description]

## Quick Start

### Option 1: Air Hot-Reload (Development) ⚡

Fast development with hot-reload and serverless infrastructure:

```bash
# 1. Set up free cloud accounts (5 min)
#    - Neon PostgreSQL: https://console.neon.tech/
#    - Upstash Redis:   https://console.upstash.com/
#    - CloudAMQP:       https://www.cloudamqp.com/

# 2. Configure environment
cp .env.development.example .env.development
# Edit .env.development with your credentials

# 3. Start development
make dev-air

# GraphQL: http://localhost:4080/graphql
```

**Benefits:**
- ⚡ Hot-reload on code changes (< 2s)
- 🚀 Fast rebuild times (native Go compilation)
- 💰 Free serverless infrastructure
- 🔧 Easy debugging with native execution

### Option 2: Docker Compose (Production-like) 🐳

Full containerized environment:

```bash
make dev

# GraphQL: http://localhost:8081/graphql
```

## Development

[Rest of existing content...]
```

**Success:** README shows both workflows

**Rollback:** Restore original README

---

### Step 2: Create DEVELOPMENT.md (20 min)

Create `DEVELOPMENT.md`:

```markdown
# Development Guide

## Table of Contents

- [Development Modes](#development-modes)
- [Air Hot-Reload Setup](#air-hot-reload-setup)
- [Docker Mode](#docker-mode)
- [Testing](#testing)
- [Debugging](#debugging)

---

## Development Modes

This project supports two development modes:

### Air Hot-Reload Mode (Recommended)

**Use for:** Daily development, feature work, debugging

**Benefits:**
- Hot-reload on code changes
- Faster rebuild times (native Go vs Docker)
- Lower resource usage
- Easy debugging

**Setup:** See [Air Hot-Reload Setup](#air-hot-reload-setup)

### Docker Compose Mode

**Use for:** Production-like testing, CI/CD, final validation

**Benefits:**
- Production parity
- Full containerized environment
- Consistent with deployment

**Setup:** See [Docker Mode](#docker-mode)

---

## Air Hot-Reload Setup

### Prerequisites

- Go 1.21+
- Air tool (`go install github.com/air-verse/air@latest`)
- Free cloud accounts (Neon, Upstash, CloudAMQP)

### Step 1: Install Air

```bash
go install github.com/air-verse/air@latest
# Or: make install-tools
```

### Step 2: Set Up Cloud Infrastructure

#### Neon PostgreSQL (Free)

1. Go to https://console.neon.tech/
2. Create free account
3. Create new project
4. Copy connection string
5. Add to `.env.development`:

```bash
DATABASE_URL=postgres://postgres:[PASSWORD]@[PROJECT-ID].aws.neon.tech/neondb?sslmode=require
```

#### Upstash Redis (Free)

1. Go to https://console.upstash.com/
2. Create free account
3. Create Redis database
4. Copy connection string
5. Add to `.env.development`:

```bash
REDIS_URL=redis://default:[PASSWORD]@xxx.upstash.io:6379
```

#### CloudAMQP (Free)

1. Go to https://www.cloudamqp.com/
2. Create free account
3. Create instance (Little Lemur tier)
4. Copy connection string
5. Add to `.env.development`:

```bash
CLOUDAMQP_URL=amqp://[USER]:[PASSWORD]@[HOST]/[VHOST]
```

### Step 3: Configure Environment

```bash
# Copy example
cp .env.development.example .env.development

# Validate configuration
make validate-env

# Edit with your credentials
nano .env.development
```

### Step 4: Start Development

```bash
# Start all services
make dev-air

# Start specific service
make dev-air-auth    # Auth service only
make dev-air-user    # User service only
make dev-air-wallet  # Wallet service only
make dev-air-gateway # GraphQL gateway only

# View logs
make dev-air-logs

# Check status
make dev-air-status

# Stop all services
make dev-air-stop
```

### Step 5: Start Coding

```bash
# Services run with hot-reload
# Make any code change, Air will rebuild automatically

# Watch logs
tail -f logs/auth-service.log
```

### Working with Shared Code

When you modify code in `shared/` or `proto/`:

```bash
# Option 1: Automatic restart (recommended)
# Run in separate terminal:
./scripts/watch-shared.sh

# Option 2: Manual restart
make dev-air-stop && make dev-air
```

### Infrastructure Fallback (Docker)

If you prefer local containers over cloud services:

```bash
# Set INFRA_MODE=docker in .env.development
sed -i 's/INFRA_MODE=.*/INFRA_MODE=docker/' .env.development

# Start infra containers
make infra-up

# Start services
make dev-air

# Stop everything
make dev-air-stop
make infra-down
```

---

## Docker Mode

### Quick Start

```bash
make dev

# GraphQL: http://localhost:8081/graphql
# PostgreSQL: localhost:5433
# Redis: localhost:6379
# RabbitMQ UI: http://localhost:15672
```

### Commands

```bash
make dev        # Start all services
make dev-stop   # Stop all services
make dev-logs   # View logs
make dev-clean  # Remove all data
```

---

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific service tests
cd services/auth-service && go test ./...
```

---

## Debugging

### Checking Logs

**Air Mode:**
```bash
make dev-air-logs
# Or: tail -f logs/*.log
```

**Docker Mode:**
```bash
make dev-logs
```

### Port Conflicts

If ports are already in use:

```bash
# Check what's using the port
lsof -i :4001  # macOS/Linux
netstat -ano | findstr :4001  # Windows

# Change port in .env.development
AUTH_GRPC_PORT=:4002
```

### Build Errors

```bash
# Clear Go cache
go clean -cache -modcache

# Rebuild
make dev-air-stop
make dev-air
```

### Database Connection Issues

**Serverless Mode:**
- Check DATABASE_URL is correct
- Verify Neon database is active
- Check network connectivity

**Docker Mode:**
```bash
# Check infra is running
make infra-up

# Verify postgres container
docker compose ps postgres
docker compose logs postgres
```

---

## Performance Tips

1. **Use specific shared imports** in `.air.toml` for faster rebuilds
2. **Keep Go build cache** warm (avoid `go clean -cache`)
3. **Use SSD** for project directory
4. **Close unused terminals** to free resources

---

## Common Workflows

### Adding a New Service

1. Create service directory
2. Add `.air.toml` configuration
3. Update Makefile with new target
4. Update scripts/dev-air.sh

### Updating Proto Files

```bash
# Generate protobuf
make proto

# Services using proto will auto-rebuild
# (if running with Air)
```

### Database Migrations

```bash
# Create migration
make migrate-create NAME=add_feature

# Run migrations
make migrate

# Rollback
make migrate-down
```

---

## Next Steps

- [Architecture Documentation](./docs/system-architecture.md)
- [API Documentation](./docs/api-documentation.md)
- [Deployment Guide](./docs/deployment-guide.md)
```

**Success:** Development guide created

**Rollback:** Delete DEVELOPMENT.md

---

### Step 3: Create TROUBLESHOOTING.md (15 min)

Create `TROUBLESHOOTING.md`:

```markdown
# Troubleshooting

## Common Issues

### Air Not Installed

**Symptom:** `air: command not found`

**Solution:**
```bash
go install github.com/air-verse/air@latest
# Or
make install-tools
```

---

### Environment File Missing

**Symptom:** `.env.development not found`

**Solution:**
```bash
cp .env.development.example .env.development
# Edit with your credentials
```

---

### Port Already in Use

**Symptom:** `bind: address already in use`

**Solutions:**

1. Find and kill process using port:
```bash
# macOS/Linux
lsof -ti:4001 | xargs kill -9

# Windows
netstat -ano | findstr :4001
taskkill /PID <PID> /F
```

2. Change port in `.env.development`:
```bash
AUTH_GRPC_PORT=:4002  # Use different port
```

---

### Database Connection Failed

**Symptom:** `connection refused` or `timeout`

**Serverless Mode (Neon):**
1. Verify DATABASE_URL is correct
2. Check Neon database is active (not paused)
3. Test connection:
```bash
psql "$DATABASE_URL" -c "SELECT 1"
```

**Docker Mode:**
1. Check infra is running:
```bash
make infra-up
docker compose ps postgres
```

2. Check connection string:
```bash
POSTGRES_HOST=localhost
POSTGRES_PORT=5433
```

---

### Rebuild Too Slow

**Symptom:** Air takes > 10 seconds to rebuild

**Solutions:**

1. **Optimize `.air.toml`** - Watch specific directories:
```toml
include_dir = ["cmd", "internal", "pkg", "../../shared/auth"]
# Don't watch entire ../../shared/
```

2. **Clear Go cache** (once):
```bash
go clean -cache
# Next build will be slower, then faster
```

3. **Check disk performance** - Use SSD if possible

---

### Services Not Communicating

**Symptom:** GraphQL returns "service unavailable"

**Solutions:**

1. **Check all services running:**
```bash
make dev-air-status
```

2. **Verify service URLs in .env:**
```bash
USER_SERVICE_URL=localhost:4002
WALLET_SERVICE_URL=localhost:4003
```

3. **Check logs for errors:**
```bash
tail -f logs/graphql-gateway.log
grep -i error logs/*.log
```

---

### Zombie Processes After Shutdown

**Symptom:** Services still running after `make dev-air-stop`

**Solution:**
```bash
# Kill all Air processes
pkill -f "air"

# Kill specific service binaries
pkill -f "tmp/main"

# Verify cleanup
ps aux | grep -E "air|tmp/main"
```

---

### Watch-Shared Script Not Working

**Symptom:** `entr: command not found`

**Solution:**
```bash
# macOS
brew install entr

# Linux
sudo apt-get install entr
```

---

### Windows-Specific Issues

#### Bash Not Found

**Symptom:** `'bash' is not recognized`

**Solution:** Use WSL2 or Git Bash:
1. Install WSL2: `wsl --install`
2. Run in WSL: `wsl bash`
3. Or use Git Bash (included with Git)

#### File Permission Errors

**Symptom:** `Permission denied` on scripts

**Solution:**
```bash
# In Git Bash
chmod +x scripts/*.sh

# In PowerShell
git update-index --chmod=+x scripts/*.sh
```

#### Line Ending Issues

**Symptom:** Scripts fail with syntax errors

**Solution:**
```bash
# Configure Git autocrlf
git config core.autocrlf input

# Recheckout scripts
rm scripts/*.sh
git checkout scripts/*.sh
```

---

### Cloud Service Downtime

**Symptom:** Can't connect to Neon/Upstash/CloudAMQP

**Solutions:**

1. **Check service status pages**
2. **Use Docker fallback:**
```bash
# In .env.development
INFRA_MODE=docker

# Start infra
make infra-up
make dev-air
```

---

### Build Errors After Go Module Changes

**Symptom:** `module not found` or `version conflict`

**Solution:**
```bash
# Update dependencies
go mod tidy
go mod download

# Clear cache
go clean -cache -modcache

# Restart Air
make dev-air-stop
make dev-air
```

---

## Getting Help

If you encounter issues not covered here:

1. Check [Development Guide](./DEVELOPMENT.md)
2. Search [GitHub Issues](../../issues)
3. Create new issue with:
   - OS and version
   - Error message
   - Steps to reproduce
   - Relevant logs
```

**Success:** Troubleshooting guide created

**Rollback:** Delete TROUBLESHOOTING.md

---

### Step 4: Verify Documentation Links (5 min)

```bash
# Check all links work
grep -rh "](.*.md)" . --include="*.md" | grep -oE "\]\([^)]+\)" | sort -u

# Verify referenced files exist
test -f DEVELOPMENT.md
test -f TROUBLESHOOTING.md
test -f docs/development-setup.md
```

**Success:** All documentation links valid

**Rollback:** Fix broken links

---

### Step 5: Update .env.development.example Comments (5 min)

Add helpful comments to `.env.development.example`:

```bash
# Add helpful comments for each section
# Link to setup guides
```

(Already done in Phase 03)

**Success:** Example file well-documented

**Rollback:** None (already done)

---

## Todo List

- [ ] Update README.md with dual-mode workflow
- [ ] Create DEVELOPMENT.md guide
- [ ] Create TROUBLESHOOTING.md
- [ ] Verify all documentation links
- [ ] Check code examples work
- [ ] Add platform-specific notes
- [ ] Test quick-start instructions

---

## Success Criteria

✅ README.md shows both Docker and Air modes
✅ DEVELOPMENT.md provides detailed guide
✅ TROUBLESHOOTING.md covers common issues
✅ All code examples work
✅ Documentation links valid
✅ Platform-specific notes included
✅ Serverless setup documented
✅ Quick-start < 5 minutes

**Validation:**
```bash
# Test quick-start from README
# New developer should be able to:
# 1. Read README
# 2. Run make validate-env
# 3. Run make dev-air
# 4. See services running
```

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Documentation out of sync | Medium | Medium | Review after each phase |
| Broken links | Low | Low | Test all links |
| Unclear instructions | Medium | Medium | Test with new developer |
| Missing platform notes | Low | Low | Add Windows/macOS/Linux notes |

---

## Security Considerations

- Never document actual credentials
- Use placeholders in examples
- Warn about committing `.env` files
- Document security best practices

---

## Next Steps

**After Phase 07:**
- All phases complete
- Ready for team onboarding
- Monitor for feedback

**Dependencies:**
- All previous phases completed

**Follow-up Tasks:**
- Gather user feedback
- Update documentation based on issues
- Create video tutorials (optional)

---

## Rollback Plan

**Full Rollback:**
```bash
# Restore original README
git checkout README.md

# Remove new documentation
rm DEVELOPMENT.md TROUBLESHOOTING.md
```

**Partial Rollback:**
```bash
# Keep README changes, remove guides
rm DEVELOPMENT.md TROUBLESHOOTING.md
```
