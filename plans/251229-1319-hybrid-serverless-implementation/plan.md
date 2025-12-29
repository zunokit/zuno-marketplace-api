---
title: "Hybrid Serverless Implementation Plan"
description: "Implement FREE tier serverless services for development, maintain AWS for production"
status: pending
priority: P1
effort: 8h
issue: 30
branch: feature/hybird-serverless-and-servers-infrastructure
tags: [infra, backend, database, devops, cost-optimization]
created: 2025-12-29
---

# Hybrid Serverless Implementation Plan

## Overview

Replace Docker-based infrastructure (PostgreSQL, Redis, RabbitMQ) with FREE tier serverless services for development to eliminate local resource overhead. Production remains on AWS traditional infrastructure.

**Target Stack:**
- PostgreSQL: **Supabase** (free tier)
- Redis: **Upstash** (free tier)
- RabbitMQ: **CloudAMQP Little Lemur** (free tier)

## Phases

| # | Phase | Status | Effort | Link |
|---|-------|--------|--------|------|
| 1 | Account Setup & Configuration | Pending | 1h | [phase-01](./phase-01-account-setup.md) |
| 2 | Environment Configuration | Pending | 1h | [phase-02](./phase-02-environment-config.md) |
| 3 | Application Configuration | Pending | 2h | [phase-03-application-config.md) |
| 4 | Documentation & Scripts | Pending | 2h | [phase-04-documentation-scripts.md) |
| 5 | Testing & Validation | Pending | 2h | [phase-05](./phase-05-testing-validation.md) |

## Dependencies

- GitHub Issue #30 created
- Research completed (see `plans/reports/brainstormer-251229-1302-hybrid-serverless-research.md`)
- Current Docker infrastructure working

## Key Decisions

| Decision | Rationale |
|----------|-----------|
| Supabase over Neon | All-in-one platform, includes auth/storage, familiar to team |
| Upstash for Redis | HTTP-based Redis, generous free tier, easy integration |
| CloudAMQP for RabbitMQ | Full RabbitMQ protocol support, 100 queues free |
| Dev != Prod | Accept infrastructure difference for cost savings |

## Success Criteria

- [ ] Developer can run services locally without Docker infra
- [ ] All services connect to cloud providers via environment variables
- [ ] Zero monthly cost for development environment
- [ ] Documentation complete for team onboarding
- [ ] CI/CD updated for new environment
- [ ] Migration path documented for existing data

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Cloud service downtime | Keep Docker option as fallback in README |
| Network dependency | Document offline development limitations |
| Free tier limits exhausted | Monitor usage, add alerts to dashboard |
| Provider lock-in | Use standard protocols (PostgreSQL, Redis, AMQP) |
| Connection string leaks | Add .env to .gitignore, document security |

## Related Files

- `docker-compose.yml` - Current infra (keep as production fallback)
- `services/*/internal/config/config.go` - Config to update
- `.env.example` - Environment template to extend
- `README.md` - Development setup instructions to update
