# Project Roadmap

Zuno NFT Marketplace API - Development roadmap and milestone tracking.

**Version**: 0.1.0
**Last Updated**: 2025-12-29
**Maintainer**: Zuno Development Team

---

## Overview

This roadmap tracks implementation progress across all project phases, including infrastructure, core features, advanced capabilities, and production readiness.

## Current Status

| Component | Status | Progress | Notes |
|-----------|--------|----------|-------|
| Foundation (v0.1.0) | Complete | 100% | Skeleton, gRPC schemas, DB infrastructure |
| Observability (Sentry) | In Progress | 25% | Phase 01 complete (Core Package) |
| Core Features (v0.2.0) | Pending | 0% | SIWE auth, JWT, sessions, profiles |
| Advanced Features (v0.3.0) | Pending | 0% | Redis, RabbitMQ, social features |
| Production Ready (v0.4.0) | Pending | 0% | K8s manifests, monitoring, hardening |

---

## Roadmap Phases

### Phase 1: Foundation (v0.1.0) - Complete

**Status**: ✅ Complete
**Timeline**: 2025-11-15 → Complete

| Milestone | Status | Date |
|-----------|--------|------|
| Clean skeleton structure | Complete | 2025-11-15 |
| gRPC service definitions | Complete | 2025-11-15 |
| Database schema (11 tables) | Complete | 2025-11-15 |
| CI/CD pipeline | Complete | 2025-11-15 |
| Dev tooling (Docker, Tilt) | Complete | 2025-11-15 |

---

### Phase 2: Observability (Sentry Integration) - In Progress

**Status**: 🔄 In Progress (25%)
**Timeline**: 2025-12-29 → 2025-01-05 (est)
**Effort**: ~6 hours total

| Phase | Status | Completion | Tests | Coverage |
|-------|--------|------------|-------|----------|
| 01: Core Package | ✅ Complete | 2025-12-29 | 10/10 | 71.8% |
| 02: Middleware Layer | Pending | - | 0/0 | - |
| 03: gRPC Interceptors | Pending | - | 0/0 | - |
| 04: Performance Monitoring | Pending | - | 0/0 | - |

#### Phase 01: Core Sentry Package (Complete)
- Created `shared/observability/sentry/` package
- Implemented Sentry initialization with production defaults
- Privacy scrubbing (ETH addresses, JWT, emails, private keys, CAIP-10)
- Security fixes applied per code review
- Report: `plans/reports/code-reviewer-251229-0041-sentry-phase01-security-fixes.md`

#### Remaining Phases
- **Phase 02**: HTTP middleware for request tracking, panic recovery
- **Phase 03**: gRPC interceptors for server/stream error capture
- **Phase 04**: Performance monitoring, custom transactions, APM integration

---

### Phase 3: Core Features (v0.2.0) - Pending

**Status**: ⏳ Not Started
**Timeline**: TBD
**Effort**: ~40 hours

| Feature | Status | Dependencies |
|---------|--------|--------------|
| SIWE authentication | Pending | Foundation |
| JWT token service | Pending | SIWE |
| Session management | Pending | JWT |
| User profile CRUD | Pending | Foundation |
| Wallet linking/verification | Pending | User service |
| GraphQL schema integration | Pending | All services |

---

### Phase 4: Advanced Features (v0.3.0) - Pending

**Status**: ⏳ Not Started
**Timeline**: TBD

| Feature | Status | Dependencies |
|---------|--------|--------------|
| Redis session caching | Pending | Sessions |
| RabbitMQ event integration | Pending | Core services |
| Social follow system | Pending | User profiles |
| User statistics | Pending | Activity data |
| WebSocket subscriptions | Pending | GraphQL |

---

### Phase 5: Production Readiness (v0.4.0) - Pending

**Status**: ⏳ Not Started
**Timeline**: TBD

| Component | Status | Dependencies |
|-----------|--------|--------------|
| Production K8s manifests | Pending | All features |
| Monitoring (Prometheus, Jaeger) | Pending | Sentry complete |
| Rate limiting | Pending | API complete |
| API versioning | Pending | GraphQL |
| Security hardening | Pending | All services |

---

## Changelog

### 2025-12-29 - Sentry Phase 01 Complete

**Completed**: `plans/251229-0004-sentry-comprehensive-integration/phase-01-core-package.md`

- ✅ Created `shared/observability/sentry/` package
- ✅ Implemented Sentry initialization with production defaults
- ✅ Privacy scrubbing for sensitive patterns:
  - Ethereum addresses (0x + 40 hex)
  - JWT tokens (header.payload.signature)
  - Email addresses
  - Private keys (64 hex chars)
  - CAIP-10 account IDs
- ✅ Security fixes applied:
  - Removed unused `AddBreadcrumb()` function (dead code)
  - Clarified scrubber regex patterns
- ✅ Tests: 10/10 passing, 71.8% coverage
- ✅ Code review approved

**Files Modified**:
- `shared/observability/sentry/sentry.go`
- `shared/observability/sentry/scrubber.go`
- `shared/observability/sentry/sentry_test.go`
- `go.mod` (added sentry-go dependency)

**Reports**:
- `plans/reports/code-reviewer-251229-0041-sentry-phase01-security-fixes.md`

**Next Steps**:
- Phase 02: Middleware Layer (HTTP request tracking, panic recovery)
- Phase 03: gRPC Interceptors
- Phase 04: Performance Monitoring

---

### 2025-12-04 - Project Init (v0.1.0)

**Completed**: Foundation skeleton

- ✅ gRPC service definitions (Auth, User, Wallet)
- ✅ PostgreSQL schema (11 tables across 3 databases)
- ✅ Docker Compose + Tilt dev environment
- ✅ CI/CD pipeline (lint, test, build, security)
- ✅ Project documentation structure

---

## Success Metrics

### Code Quality
- [ ] 80%+ test coverage on new code
- [ ] Zero critical security issues
- [ ] Zero golangci-lint warnings
- [x] All tests passing in CI/CD

### Performance Targets
- [ ] API response time < 100ms (p95)
- [ ] gRPC latency < 50ms (p95)
- [ ] Database query time < 20ms (p95)

### Reliability
- [ ] 99.9% uptime target
- [ ] All endpoints health-checked
- [ ] Graceful degradation

### Security
- [x] SIWE signature verification planned
- [x] JWT expiration enforcement planned
- [x] Prepared statements (SQL injection prevention)
- [x] No plaintext secrets in code

---

## Dependencies

```
Sentry Integration (Phase 2)
    ├── Core Package (Phase 01) ✅ COMPLETE
    ├── Middleware Layer (Phase 02) → Depends: Core
    ├── gRPC Interceptors (Phase 03) → Depends: Core
    └── Performance Monitoring (Phase 04) → Depends: Middleware, Interceptors

Core Features (Phase 3)
    ├── SIWE Auth → Depends: Foundation ✅
    ├── JWT Service → Depends: SIWE
    ├── Sessions → Depends: JWT
    ├── User Profiles → Depends: Foundation ✅
    ├── Wallets → Depends: Users
    └── GraphQL → Depends: All services
```

---

## Risks & Blockers

| Risk | Severity | Status | Mitigation |
|------|----------|--------|------------|
| Sentry integration complexity | Medium | Active | Phased approach, code review each phase |
| SIWE test coverage gaps | High | Known | Placeholder tests exist, need implementation |
| RabbitMQ integration unknown | Medium | Not started | Research phase planned |
| Production K8s experience | Low | Not started | Reference patterns from similar projects |

---

## Unresolved Questions

1. **Panic Recovery**: Should middleware recover from panics or let them crash?
2. **gRPC Error Context**: How much request detail in Sentry events?
3. **Performance Budget**: Target transaction duration thresholds?
4. **Session Token TTL**: Current 15min access / 7d refresh - confirm?
5. **Wallet Verification**: Mandatory or optional for multi-wallet?

## Resolved Decisions (2025-12-29)

**Sentry Configuration**:
1. **Sentry Project Strategy**: Share one project for all services (current), future scaling to per-service projects
2. **Trace Sampling Rate**: 0.2 (20%) for production - approved
3. **Custom Tags**: Add `user_id` and `wallet_hash` (SHA256) tags for user context

---

**Navigation**:
- Project Overview: `docs/project-overview-pdr.md`
- Code Standards: `docs/code-standards.md`
- Architecture: `docs/system-architecture.md`
- Deployment: `docs/deployment-guide.md`
