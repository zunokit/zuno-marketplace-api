---
title: "Phase 06: Testing & Validation"
description: "End-to-end testing of Air hot-reload workflow"
status: pending
priority: P1
effort: 1.5h
branch: feature/hybird-serverless-and-servers-infrastructure
tags: [testing, validation, air, hot-reload]
created: 2026-01-28
---

# Phase 06: Testing & Validation

## Context Links

- [Main Plan](plan.md)
- [Phase 02: Air Configuration](phase-02-air-configuration.md)
- [Phase 03: Environment Setup](phase-03-environment-setup.md)
- [Phase 04: Scripts Creation](phase-04-scripts-creation.md)
- [Phase 05: Makefile Enhancement](phase-05-makefile-enhancement.md)

## Overview

**Priority:** P1 (validation before completion)
**Current Status:** Pending
**Description:** Comprehensive testing of Air hot-reload workflow, including single service, multi-service, shared code reloads, and rollback validation.

**Test Categories:**
1. **Single Service Testing** - Individual Air hot-reload
2. **Multi-Service Testing** - All services together
3. **Shared Code Reload** - Changes in `shared/` directory
4. **Infrastructure Testing** - Serverless vs Docker modes
5. **Rollback Testing** - Verify production workflow unchanged
6. **Performance Testing** - Rebuild times

---

## Key Insights

From [Air Research](research/researcher-01-air-hot-reload.md):
- **Rebuild time target:** < 5 seconds per service
- **Shared code watching:** Can cause conflicts if not configured properly
- **Port binding:** Must verify services don't conflict during restart
- **Build cache:** Go cache speeds up incremental builds

**Testing Strategy:**
- Start with single service (simplest)
- Progress to multi-service (integration)
- Test shared code changes (critical path)
- Verify rollback (safety net)
- Measure performance (optimization)

---

## Requirements

### Functional Requirements
1. All 4 services start successfully with Air
2. Code changes trigger hot-reload in < 5 seconds
3. Shared code changes restart affected services
4. Services communicate correctly (gRPC/HTTP)
5. Infra connections work (Neon/Upstash/CloudAMQP or Docker)
6. Production Docker workflow unchanged

### Non-Functional Requirements
- No port conflicts
- No zombie processes
- Clean shutdown possible
- Logs written to `logs/` directory
- Windows/WSL compatibility verified

---

## Architecture

### Test Sequence

```
1. Environment Validation
   └─→ validate-env.sh

2. Single Service Test
   └─→ dev-air.sh auth
   └─→ Make code change
   └─→ Verify rebuild

3. Multi-Service Test
   └─→ dev-air.sh all
   └─→ Check all services running

4. Communication Test
   └─→ GraphQL → Auth (gRPC)
   └─→ GraphQL → User (gRPC)
   └─→ GraphQL → Wallet (gRPC)

5. Shared Code Reload Test
   └─→ Change shared/auth code
   └─→ Verify auth-service restarts

6. Rollback Test
   └─→ make dev (Docker)
   └─→ Verify production workflow

7. Performance Test
   └─→ Measure rebuild times
   └─→ Verify < 5 second target
```

---

## Related Code Files

### Files to TEST
- `services/*/.air.toml` - Air configs
- `.env.development` - Environment configuration
- `scripts/dev-air.sh` - Startup script
- `scripts/stop-air.sh` - Shutdown script
- `scripts/validate-env.sh` - Validation script
- `Makefile` - All new targets

### Test Data
- Modify service code to trigger rebuilds
- Modify shared code to test cross-service reloads
- Query GraphQL endpoints for integration tests

---

## Implementation Steps

### Step 1: Prerequisites Check (5 min)

```bash
# Verify Air is installed
air version

# Verify environment file exists
test -f .env.development || echo "Create .env.development first"

# Validate environment
make validate-env

# Verify Docker is available (for infra testing)
docker --version
docker compose version
```

**Success:** All prerequisites met

**Rollback:** Install missing tools

---

### Step 2: Single Service Hot-Reload Test (15 min)

```bash
# Start auth-service only
make dev-air-auth

# In another terminal, watch logs
tail -f logs/auth-service.log

# Make a code change
echo "// Test hot-reload" >> services/auth-service/cmd/main.go

# Verify in logs:
# Expected: "building..." followed by "running..."
# Should happen within 5 seconds

# Verify service still responds
# (If you have a health endpoint)
curl localhost:4001 || echo "No health endpoint, check logs"

# Stop service
make dev-air-stop
```

**Success:** Code change triggers rebuild < 5s

**Rollback:** Debug Air config if rebuild doesn't trigger

---

### Step 3: Multi-Service Startup Test (15 min)

```bash
# Start all services
make dev-air

# Wait for startup
sleep 10

# Check status
make dev-air-status

# Expected output:
#   ✓ auth-service (PID: xxxxx)
#   ✓ user-service (PID: xxxxx)
#   ✓ wallet-service (PID: xxxxx)
#   ✓ graphql-gateway (PID: xxxxx)

# Check all logs
tail -n 20 logs/*.log

# Verify ports are bound
netstat -an | grep -E "400[0-9]|4080" || echo "Ports bound successfully"

# Stop all services
make dev-air-stop
```

**Success:** All 4 services start and stop cleanly

**Rollback:** Check port conflicts, fix configs

---

### Step 4: Inter-Service Communication Test (15 min)

```bash
# Start all services
make dev-air

# Wait for startup
sleep 10

# Test GraphQL gateway (port 4080)
curl -X POST http://localhost:4080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "{ __typename }"}'

# Expected: {"data":{"__typename":"Query"}}

# If you have authentication queries:
# curl -X POST http://localhost:4080/graphql \
#   -H "Content-Type: application/json" \
#   -d '{"query": "mutation { login(email: \"test@example.com\", password: \"test\") { token } }"}'

# Check logs for gRPC communication
grep -r "grpc" logs/*.log

# Stop services
make dev-air-stop
```

**Success:** GraphQL gateway communicates with gRPC services

**Rollback:** Debug service URLs in .env

---

### Step 5: Shared Code Reload Test (15 min)

```bash
# Start all services
make dev-air

# Wait for startup
sleep 10

# Make a change in shared code
echo "// Shared code test" >> shared/auth/logger.go

# Watch auth-service log for rebuild
tail -f logs/auth-service.log

# Expected: auth-service should rebuild within 5-10 seconds
# Other services should NOT rebuild (they don't import shared/auth)

# Alternative: Change shared/proto
echo "// Proto test" >> shared/proto/common.go

# All services importing proto should rebuild
tail -f logs/*.log

# Stop services
make dev-air-stop
```

**Success:** Shared code changes trigger correct rebuilds

**Rollback:** Adjust include_dir in .air.toml configs

---

### Step 6: Infrastructure Mode Test (20 min)

**Option A: Serverless Mode**

```bash
# Ensure .env.development has INFRA_MODE=serverless
grep "INFRA_MODE=serverless" .env.development

# Start services (no infra containers needed)
make dev-air

# Wait for startup
sleep 10

# Check logs for cloud connections
grep -E "DATABASE_URL|REDIS_URL|CLOUDAMQP_URL" logs/*.log

# Test database connection
# (If you have a health check that queries DB)
curl localhost:4080/health || echo "No health endpoint"

# Stop services
make dev-air-stop
```

**Option B: Docker Infra Mode**

```bash
# Set INFRA_MODE=docker in .env.development
sed -i 's/INFRA_MODE=.*/INFRA_MODE=docker/' .env.development

# Start infra containers
make infra-up

# Verify infra running
docker compose ps postgres redis rabbitmq

# Start services
make dev-air

# Wait for startup
sleep 10

# Check logs for localhost connections
grep -E "localhost:5433|localhost:6379" logs/*.log

# Stop everything
make dev-air-stop
make infra-down

# Restore serverless mode
sed -i 's/INFRA_MODE=.*/INFRA_MODE=serverless/' .env.development
```

**Success:** Both infra modes work

**Rollback:** Debug connection strings

---

### Step 7: Production Workflow Rollback Test (10 min)

```bash
# Verify Docker mode still works
make dev

# Wait for startup
sleep 10

# Check services running
docker compose ps

# Test GraphQL
curl -X POST http://localhost:8081/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "{ __typename }"}'

# Stop Docker
make dev-stop

# Verify clean shutdown
docker compose ps
```

**Success:** Production workflow unchanged

**Rollback:** Investigate Docker config changes

---

### Step 8: Performance Benchmark (10 min)

```bash
# Start a single service
make dev-air-auth

# Measure rebuild time (3 iterations)
for i in 1 2 3; do
  echo "Iteration $i:"
  time {
    # Make a change
    echo "// Test $i" >> services/auth-service/cmd/main.go

    # Wait for rebuild (check in logs)
    sleep 6
  }
done

# Stop service
make dev-air-stop

# Expected: Each rebuild < 5 seconds
```

**Success:** Average rebuild time < 5s

**Rollback:** Optimize build if slow (check include_dir)

---

### Step 9: Cleanup Validation (5 min)

```bash
# Start all services
make dev-air

# Stop with Ctrl+C on logs, then:
make dev-air-stop

# Check for zombie processes
ps aux | grep -E "air|tmp/main" | grep -v grep

# Should return empty

# Check PID files cleaned
ls logs/*.pid 2>/dev/null || echo "All PID files cleaned"

# Check tmp directories
ls services/*/tmp 2>/dev/null | wc -l

# Should show 4 tmp directories (one per service)
```

**Success:** Clean shutdown, no zombie processes

**Rollback:** Fix stop-air.sh script

---

## Todo List

- [ ] Verify Air installed
- [ ] Run make validate-env
- [ ] Test single service hot-reload
- [ ] Test multi-service startup
- [ ] Test inter-service communication
- [ ] Test shared code reload
- [ ] Test serverless infra mode
- [ ] Test Docker infra mode
- [ ] Test production rollback (make dev)
- [ ] Measure rebuild performance
- [ ] Verify cleanup (no zombies)

---

## Success Criteria

✅ All 4 services start with Air successfully
✅ Code changes trigger rebuild in < 5 seconds
✅ Shared code changes restart affected services
✅ GraphQL gateway communicates with gRPC services
✅ Serverless infra connections work
✅ Docker infra fallback works
✅ Production workflow (make dev) unchanged
✅ No zombie processes after shutdown
✅ Logs written to logs/ directory
✅ Make help shows both modes clearly

**Validation Checklist:**
```bash
# All services start
make dev-air && make dev-air-status

# Hot-reload works
# (Make code change, verify rebuild)

# Communication works
curl http://localhost:4080/graphql

# Production unchanged
make dev && make dev-stop

# Performance OK
# (Rebuild times < 5s)
```

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Rebuild time > 5s | Medium | Low | Optimize include_dir paths |
| Shared code conflicts | Medium | Medium | Use specific include_dir |
| Port conflicts | Low | High | Check 4xxx range availability |
| Zombie processes | Low | Medium | Improve stop-air.sh script |
| Serverless downtime | Medium | Medium | Docker fallback available |
| Production workflow broken | Low | Critical | Test make dev thoroughly |

---

## Security Considerations

- Test with `.env.development` (never commit)
- Verify no credentials in logs
- Ensure `.env.development` in `.gitignore`
- Check for sensitive data in error messages

---

## Next Steps

**After Phase 06:**
- Phase 07: Documentation (update README, create guides)
- Address any test failures
- Optimize rebuild times if needed

**Dependencies:**
- All previous phases completed

**Follow-up Tasks:**
- Document test results
- Create troubleshooting guide
- Set up CI/CD integration (if needed)

---

## Rollback Plan

**Full Rollback:**
```bash
# If Air workflow has critical issues:
git checkout main -- Makefile .env.development.example
rm -rf services/*/.air.toml scripts/dev-air.sh scripts/stop-air.sh scripts/watch-shared.sh scripts/validate-env.sh

# Use production workflow
make dev
```

**Partial Rollback:**
```bash
# If specific service has issues:
rm services/auth-service/.air.toml
# Rebuild config from Phase 02
```

**Issue Tracking:**
Document any failures for resolution in Phase 07
