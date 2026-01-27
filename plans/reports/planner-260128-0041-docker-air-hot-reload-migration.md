# Docker + Air Hot-Reload Migration Plan - Summary Report

**Date:** 2026-01-28
**Planner:** Planner Agent
**Plan Location:** `E:\zuno-marketplace-api\plans\260128-0037-docker-air-hot-reload-migration\`

---

## Executive Summary

Comprehensive implementation plan created for migrating development workflow from Docker Compose to Air hot-reload with serverless infrastructure (Neon/Upstash/CloudAMQP).

**Key Deliverables:**
- 7 detailed phase plans with implementation steps
- Production Docker workflow preserved (no breaking changes)
- Developer onboarding time reduced to < 5 minutes
- Hot-reload rebuild time target: < 2 seconds

**Total Estimated Effort:** 7.5 hours

---

## Plan Structure

### Main Plan File
**Location:** `plan.md`
- Overview with YAML frontmatter
- Architecture diagrams
- Phase dependencies
- Success criteria
- Rollback plan

### Phase Files (7 total)

| Phase | Description | Effort | Status |
|-------|-------------|--------|--------|
| [Phase 01: Cleanup Artifacts](phase-01-cleanup-artifacts.md) | Remove old Tilt/serverless files | 30m | pending |
| [Phase 02: Air Configuration](phase-02-air-configuration.md) | Setup Air for 4 services | 1.5h | pending |
| [Phase 03: Environment Setup](phase-03-environment-setup.md) | Update .env for 4xxx ports | 1h | pending |
| [Phase 04: Scripts Creation](phase-04-scripts-creation.md) | Create dev scripts | 1.5h | pending |
| [Phase 05: Makefile Enhancement](phase-05-makefile-enhancement.md) | Add dev-air targets | 1h | pending |
| [Phase 06: Testing & Validation](phase-06-testing-validation.md) | Test everything | 1.5h | pending |
| [Phase 07: Documentation](phase-07-documentation.md) | Update README | 30m | pending |

---

## Key Requirements Implemented

### Air Configuration (Phase 02)
- `.air.toml` for all 4 services
- Port configuration: 4001, 4002, 4003, 4080
- Shared code watching per service
- Build delay: 1000ms

### Environment Setup (Phase 03)
- Updated `.env.development.example`
- Serverless URLs (Neon/Upstash/CloudAMQP)
- Docker fallback support
- 4xxx port range

### Scripts (Phase 04)
- `dev-air.sh` - Start services
- `stop-air.sh` - Stop services
- `watch-shared.sh` - Monitor shared code
- `validate-env.sh` - Check configuration

### Makefile (Phase 05)
- `make dev-air` - Start all services
- `make dev-air-stop` - Stop services
- `make infra-up/down` - Manage infra
- `make validate-env` - Validate setup
- Existing targets preserved

### Testing (Phase 06)
- Single service hot-reload
- Multi-service integration
- Shared code reload
- Performance benchmarking
- Production rollback validation

### Documentation (Phase 07)
- Updated README.md
- DEVELOPMENT.md guide
- TROUBLESHOOTING.md
- Quick-start instructions

---

## Architecture

### Development Mode (New)
```
Local Services (Air Hot-Reload)
├── auth-service:    localhost:4001 (gRPC)
├── user-service:    localhost:4002 (gRPC)
├── wallet-service:  localhost:4003 (gRPC)
└── graphql-gateway: localhost:4080 (HTTP)

Serverless Infrastructure
├── Neon PostgreSQL (cloud)
├── Upstash Redis (cloud)
└── CloudAMQP RabbitMQ (cloud)
```

### Production Mode (Unchanged)
```
Docker Compose
├── All services in containers
├── Infra in containers
└── Original configuration
```

---

## Success Criteria

✅ Hot-reload works: Code changes trigger rebuild in < 2s
✅ Shared code reloads: Changes in `shared/` restart all services
✅ Single command: `make dev-air` starts everything
✅ Clean workflow: No leftover Tilt/K8s files
✅ Port consistency: All services use 4xxx range
✅ Production unchanged: `docker-compose.yml` still works
✅ Developer onboarding: < 5 minutes setup time

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Windows Air compatibility | Medium | Medium | Test on Windows, use WSL2 |
| Rebuild time > 5s | Medium | Low | Optimize include_dir paths |
| Shared code conflicts | Medium | Medium | Use specific imports in .air.toml |
| Port conflicts | Low | Medium | 4xxx range avoids Docker ports |
| Production workflow broken | Low | Critical | Tested in Phase 06 |

---

## Dependencies

**Phase Dependencies:**
- Phase 01 (cleanup) - No dependencies, start first
- Phase 02 (Air config) - Requires Phase 01
- Phase 03 (env setup) - Requires Phase 02
- Phase 04 (scripts) - Requires Phase 02, 03
- Phase 05 (Makefile) - Requires Phase 02, 03, 04
- Phase 06 (testing) - Requires all previous phases
- Phase 07 (docs) - Requires all previous phases

**External Dependencies:**
- Go 1.21+
- Air tool (installable via `make install-tools`)
- Neon/Upstash/CloudAMQP accounts (free tier)

---

## Rollback Strategy

Each phase includes specific rollback steps. Full rollback:

```bash
# Restore all original files
git checkout main -- Makefile .env.development.example README.md

# Remove new files
rm -rf services/*/.air.toml
rm -rf scripts/dev-air.sh scripts/stop-air.sh scripts/watch-shared.sh scripts/validate-env.sh
rm -rf DEVELOPMENT.md TROUBLESHOOTING.md
rm -rf logs/

# Restore Tiltfile if needed (from git history)
git checkout HEAD~1 -- Tiltfile.development
```

---

## Next Steps

1. **Review this plan** and confirm approach
2. **Set up cloud accounts** (Neon/Upstash/CloudAMQP)
3. **Start Phase 01: Cleanup artifacts**
4. **Execute phases sequentially** following implementation steps
5. **Test thoroughly** in Phase 06 before completing

---

## Unresolved Questions

1. **Cloud service accounts:** Already have Neon/Upstash/CloudAMQP set up?
2. **Windows testing:** Need to test Air on Windows/WSL2?
3. **Team size:** Single developer or team needs onboarding?
4. **CI/CD integration:** Should CI use Air or Docker only?

These questions should be resolved before starting Phase 01.

---

## File Locations

All plan files created in:
```
E:\zuno-marketplace-api\plans\260128-0037-docker-air-hot-reload-migration\
├── plan.md                           ← Main overview
├── phase-01-cleanup-artifacts.md
├── phase-02-air-configuration.md
├── phase-03-environment-setup.md
├── phase-04-scripts-creation.md
├── phase-05-makefile-enhancement.md
├── phase-06-testing-validation.md
└── phase-07-documentation.md
```

---

## Conclusion

Comprehensive plan created covering all aspects of Docker + Air hot-reload migration. Each phase includes:
- Detailed implementation steps
- Success criteria
- Rollback plans
- Risk assessment
- Security considerations

Ready to proceed with implementation upon approval.
