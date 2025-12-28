---
title: "Comprehensive Sentry Integration"
description: "Integrate error tracking, performance monitoring, and distributed tracing using Sentry Go SDK across all microservices"
status: pending
priority: P1
effort: 5h
issue: 0004
branch: feature/add-sentry
tags: [observability, backend, security, infra]
created: 2025-12-29
---

# Comprehensive Sentry Integration Plan

## Overview

Integrate **Sentry Go SDK** for production-ready observability across all 4 microservices. Includes error tracking, performance monitoring, distributed tracing, auto-scrubbing of sensitive data (wallet addresses, JWTs, PII), and CI/CD integration.

**Approach**: Hybrid KISS - shared observability package with single SDK dependency, following YAGNI/KISS/DRY principles.

**Related**: `plans/reports/brainstormer-251229-0004-sentry-comprehensive-integration.md`

---

## Phases

| # | Phase | Status | Effort | Link |
|---|-------|--------|--------|------|
| 1 | Core Package | Pending | 1.5h | [phase-01-core-package.md](./phase-01-core-package.md) |
| 2 | Middleware Layer | Pending | 1.5h | [phase-02-middleware.md](./phase-02-middleware.md) |
| 3 | Service Integration | Pending | 1h | [phase-03-service-integration.md](./phase-03-service-integration.md) |
| 4 | Distributed Tracing | Pending | 0.5h | [phase-04-distributed-tracing.md](./phase-04-distributed-tracing.md) |
| 5 | CI/CD Integration | Pending | 0.5h | [phase-05-cicd-integration.md](./phase-05-cicd-integration.md) |

---

## Architecture

```
Sentry Cloud
    │
    ▼ sentry-trace header
┌─────────────────────────────────────────────┐
│  shared/observability/ (NEW)                │
│  ├── sentry/        Init, Flush, Scrubbing  │
│  ├── middleware/    HTTP, gRPC, GraphQL     │
│  └── tracing/       Propagation, Sampling   │
└─────────────────────────────────────────────┘
         │           │           │           │
         ▼           ▼           ▼           ▼
    GraphQL GW   Auth Svc   User Svc   Wallet Svc
    (HTTP+GraphQL) (gRPC)    (gRPC)      (gRPC)
```

---

## Dependencies

| Type | Dependency | Required For |
|------|------------|--------------|
| External | `github.com/getsentry/sentry-go` v0.27.0+ | Sentry SDK |
| Internal | `shared/env/` | Config loading |
| Internal | `shared/proto/pb/` | gRPC definitions |
| Service | All 4 services (`services/*/cmd/main.go`) | Integration points |

---

## Success Criteria

- [ ] 100% of panics/errors captured in Sentry
- [ ] Distributed traces visible across all services (waterfalls)
- [ ] 0 sensitive data leaked (wallet addresses, JWTs, emails)
- [ ] Performance impact <5ms p99 latency
- [ ] CI/CD creates releases and tracks deploys
- [ ] All services flush events on graceful shutdown

---

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| DSN leaked | High | Use secrets manager, .gitignore |
| Quota exceeded | Medium | Smart sampling (5-20% traces) |
| Trace propagation breaks | Medium | E2E test coverage |
| PII leakage | High | Multi-layer scrubbing + tests |

---

## File Changes Summary

### Create (11 files)
```
shared/observability/
├── sentry/
│   ├── sentry.go
│   └── scrubber.go
├── middleware/
│   ├── http.go
│   ├── grpc.go
│   └── graphql.go
├── tracing/
│   ├── propagator.go
│   └── sampler.go
└── README.md
```

### Modify (8 files)
```
services/auth-service/
├── cmd/main.go
└── internal/config/config.go
services/graphql-gateway/
├── cmd/main.go
└── internal/config/config.go
services/user-service/
├── cmd/main.go
└── internal/config/config.go
services/wallet-service/
├── cmd/main.go
└── internal/config/config.go
```

### Modify (3 files)
```
.env.example
.github/workflows/deploy.yml
go.mod
```

---

## References

- Brainstorming Report: `plans/reports/brainstormer-251229-0004-sentry-comprehensive-integration.md`
- Sentry Go Docs: https://docs.sentry.io/platforms/go/
- Sentry Tracing: https://docs.sentry.io/platforms/go/tracing/
