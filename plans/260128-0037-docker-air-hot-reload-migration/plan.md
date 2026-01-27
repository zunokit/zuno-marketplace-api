---
title: "Docker + Air Hot-Reload Migration"
description: "Migrate development workflow from Docker Compose to Air hot-reload with serverless infrastructure"
status: pending
priority: P1
effort: 7.5h
branch: feature/hybird-serverless-and-servers-infrastructure
tags: [docker, air, hot-reload, serverless, neon, upstash, cloudamqp]
created: 2026-01-28
---

# Docker + Air Hot-Reload Migration Plan

## Context

**Project:** zuno-marketplace-api (Go microservices)

**Current State:**
- Full Docker Compose for all services (app + infra)
- Tiltfile.development (old serverless setup)
- Tiltfile.production (production K8s setup)

**Target State:**
- **Production:** Full Docker (keep `Tiltfile.production`, `docker-compose.yml`)
- **Development:** Local Air hot-reload + serverless infra (Neon/Upstash/CloudAMQP)
- **Cleanup:** Remove `Tiltfile.development` only

**Key Requirement:** Keep production Docker workflow unchanged

---

## Overview

| Phase | Description | Effort | Status |
|-------|-------------|--------|--------|
| [Phase 01: Cleanup Artifacts](phase-01-cleanup-artifacts.md) | Remove old Tilt/serverless files | 30m | pending |
| [Phase 02: Air Configuration](phase-02-air-configuration.md) | Setup Air for 4 services | 1.5h | pending |
| [Phase 03: Environment Setup](phase-03-environment-setup.md) | Update .env for 4xxx ports | 1h | pending |
| [Phase 04: Scripts Creation](phase-04-scripts-creation.md) | Create dev scripts | 1.5h | pending |
| [Phase 05: Makefile Enhancement](phase-05-makefile-enhancement.md) | Add dev-air targets | 1h | pending |
| [Phase 06: Testing & Validation](phase-06-testing-validation.md) | Test everything | 1.5h | pending |
| [Phase 07: Documentation](phase-07-documentation.md) | Update README | 30m | pending |

**Total Estimated Effort:** 7.5 hours

---

## Architecture

### Production (Unchanged)
```
Docker Compose (docker-compose.yml)
├── PostgreSQL (container)
├── Redis (container)
├── RabbitMQ (container)
├── auth-service (container)
├── user-service (container)
├── wallet-service (container)
└── graphql-gateway (container)
```

### Development (New)
```
Local Services (Air Hot-Reload)
├── auth-service:    localhost:4001 (gRPC)
├── user-service:    localhost:4002 (gRPC)
├── wallet-service:  localhost:4003 (gRPC)
└── graphql-gateway: localhost:4080 (HTTP)

Serverless Infrastructure
├── PostgreSQL: Neon (cloud)
├── Redis: Upstash (cloud)
└── RabbitMQ: CloudAMQP (cloud)
```

---

## Dependencies

**Phase Dependencies:**
- Phase 02 → Phase 03 (Air config needed before env setup)
- Phase 02 → Phase 04 (Air config needed for scripts)
- Phase 03 → Phase 05 (Env vars needed for Makefile)
- All phases → Phase 06 (Test after implementation)
- All phases → Phase 07 (Document after testing)

**Blocked By:**
- None (can start immediately)

**Blocks:**
- Team onboarding for new dev workflow
- CI/CD updates (if needed)

---

## Success Criteria

✅ Hot-reload works: Code changes trigger rebuild in <2s
✅ Shared code reloads: Changes in `shared/` restart all services
✅ Single command: `make dev-air` starts everything
✅ Clean workflow: No leftover Tilt/K8s files
✅ Port consistency: All services use 4xxx range
✅ Production unchanged: `docker-compose.yml` still works
✅ Developer onboarding: <5 minutes setup time

---

## Rollback Plan

Each phase includes rollback steps. Full rollback:
```bash
# Restore production workflow
git checkout main -- Makefile .env.development.example
rm -rf services/*/.air.toml scripts/dev-air.sh scripts/watch-shared.sh

# Restore Tiltfile if needed (from git history)
git checkout HEAD~1 -- Tiltfile.development
```

---

## Research References

- [Air Hot-Reload Research](research/researcher-01-air-hot-reload.md)
- [Hybrid Workflow Research](research/researcher-02-hybrid-workflow.md)
- [Brainstorm Report](../reports/brainstorm-260128-0037-docker-air-hot-reload-migration.md)

---

## Validation Summary

**Validated:** 2026-01-28
**Questions asked:** 4

### Confirmed Decisions
- **Shared code watching:** Option 1 - Each service watches specific shared imports
- **Infrastructure mode:** Serverless only (Neon/Upstash/CloudAMQP), no Docker fallback
- **Windows support:** Native Windows (Air supports Windows directly)
- **CI/CD:** Docker only in CI, Air for local development only

### Plan Adjustments Needed
- [ ] Remove Docker fallback logic from environment setup
- [ ] Remove watch-shared.sh script (not needed with Option 1)
- [ ] Document native Windows Air support
- [ ] Clarify CI/CD uses Docker Compose only

---

## Unresolved Questions

1. **gRPC port conflicts:** Verify services don't bind same port during hot-reload restart (test in Phase 06)
2. **Cloud service accounts:** Confirm Neon/Upstash/CloudAMQP credentials ready

---

## Next Steps

1. ✅ Plan validated with user decisions
2. Start Phase 01: Cleanup artifacts
