---
title: "Neon + Upstash + CloudAMQP Integration"
description: "Integrate Neon PostgreSQL, Upstash Redis, and CloudAMQP RabbitMQ for hybrid serverless/docker infrastructure"
status: pending
priority: P1
effort: 14h
branch: feature/hybird-serverless-and-servers-infrastructure
tags: [infrastructure, database, redis, rabbitmq, serverless]
created: 2026-01-27
---

# Neon + Upstash + CloudAMQP Integration Plan

## Overview

Full implementation of Neon PostgreSQL, Upstash Redis, and CloudAMQP RabbitMQ integration for hybrid serverless/docker infrastructure mode.

### Current State
- Hybrid config exists (Tiltfile.development, Tiltfile.production)
- Environment templates define DATABASE_URL, REDIS_URL, CLOUDAMQP_URL
- Config package has Mode detection (serverless/docker)
- **Gap**: No actual Redis/RabbitMQ client initialization code

### Goal
- Phase 1: Neon Database integration (verify pgx driver compatibility)
- Phase 2: Upstash Redis integration (caching, rate limiting)
- Phase 3: CloudAMQP RabbitMQ integration (pub/sub events)
- Phase 4: Testing & validation

## Phases

| Phase | Status | Files | Effort |
|-------|--------|-------|--------|
| [01-neon-database.md](./phase-01-neon-database.md) | pending | 3 | 2h |
| [02-upstash-redis.md](./phase-02-upstash-redis.md) | pending | 15 | 6h |
| [03-cloudamqp-rabbitmq.md](./phase-03-cloudamqp-rabbitmq.md) | pending | 12 | 4h |
| [04-testing-validation.md](./phase-04-testing-validation.md) | pending | 8 | 2h |

## Dependencies

```
phase-02-upstash-redis ───┐
                         ├──> phase-04-testing-validation
phase-03-cloudamqp-rabbit ─┘
       ^
       │
phase-01-neon-database
```

## Quick Links

- Research: [./reports/](./reports/)
- Root: `E:\zuno-marketplace-api`
- Branch: `feature/hybird-serverless-and-servers-infrastructure`

## Unresolved Questions

1. Neon-specific connection pooling requirements vs Supabase?
2. Upstash Redis TLS requirements for Go client?
3. CloudAMQP queue topology design (exchanges, routing keys)?
4. Event schema for pub/sub (user.created, wallet.updated)?
