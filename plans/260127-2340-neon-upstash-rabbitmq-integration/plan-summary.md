# Plan Summary: Neon + Upstash + CloudAMQP Integration

**Location:** `plans/260127-2340-neon-upstash-rabbitmq-integration/`

## Overview

Comprehensive plan to integrate Neon PostgreSQL, Upstash Redis, and CloudAMQP RabbitMQ for hybrid serverless/docker infrastructure mode.

## Phase Breakdown

| Phase | Description | Files | Effort |
|-------|-------------|-------|--------|
| 1. Neon Database | Verify pgx compat, update comments | 3 | 2h |
| 2. Upstash Redis | Client, wrapper, caching, rate limit | 15 | 6h |
| 3. CloudAMQP RabbitMQ | Client, wrapper, pub/sub events | 12 | 4h |
| 4. Testing | Unit/integration tests, validation | 8 | 2h |

**Total:** 38 files, 14 hours

## Key Components

### shared/redis Package
- `client.go` - Singleton connection mgmt
- `cache.go` - Get/Set/Delete with TTL
- `rate_limit.go` - Sliding window rate limiting
- `lock.go` - Distributed locking
- `health.go` - Ping check

### shared/rabbitmq Package
- `client.go` - Connection + topology setup
- `events.go` - Event type definitions
- `publisher.go` - Publish with routing keys
- `consumer.go` - Consume with ack/reject
- `health.go` - Connection check

### Integration Points
- **Auth Service:** Session caching, auth.login events
- **User Service:** user.created events
- **Wallet Service:** wallet.created events
- **Gateway:** Rate limiting middleware

## Next Steps

1. Review plan files
2. Start Phase 1 (Neon)
3. Progress through phases sequentially
4. Validate both docker + serverless modes

## Questions for User

1. **Neon vs Supabase:** Switching comments from Supabase to Neon - confirm this is correct?
2. **Event Schema:** OK with proposed event structure (id, type, source, timestamp, data)?
3. **Testcontainers:** Use for isolated integration tests, or require real service connections?
4. **Rate Limits:** Default limits OK (100 req/hour per IP)?

---

**Files Created:**
- `plan.md` - Overview with dependencies
- `phase-01-neon-database.md` - Neon integration
- `phase-02-upstash-redis.md` - Redis integration
- `phase-03-cloudamqp-rabbitmq.md` - RabbitMQ integration
- `phase-04-testing-validation.md` - Testing strategy
