---
title: "Advisory Lock Webhook Idempotency Fix"
description: "Fix context canceled errors using PostgreSQL advisory locks for webhook idempotency with gRPC observability"
status: pending
priority: P0
effort: 6h
branch: fix/chore-compare-address-collection
tags: [webhook, idempotency, postgresql, advisory-locks, grpc, observability]
created: 2026-02-04
---

## Overview

Fix "context canceled" errors in webhook processing by implementing PostgreSQL advisory locks for event serialization, adding gRPC observability interceptor, and creating concurrent webhook processing tests.

**Problem:** Race condition in idempotency check allows duplicate events to process concurrently, causing database lock contention.

**Solution:** Advisory locks + gRPC interceptor + integration tests

---

## Dependency Graph & Parallelization

### Round 1: Parallel Execution (Independent Phases)
```
Phase 1: Repository Lock Methods  ─┐
                                   ├─→ Can run in parallel
Phase 3: gRPC Interceptor        ─┘

Phase 2: Webhook Handler (depends on Phase 1)
Phase 4: Integration Tests (depends on 1, 2, 3)
```

### Execution Strategy

| Round | Phases | Description | Files Modified |
|-------|--------|-------------|----------------|
| **Round 1** | Phase 1 + Phase 3 | Repository lock methods & gRPC interceptor (no overlap) | `processed-event-repository.go` + `main.go` |
| **Round 2** | Phase 2 | Webhook handler integration (uses lock methods) | `webhook_handler.go` |
| **Round 3** | Phase 4 | Integration tests (tests complete flow) | `*_test.go` files |

**Estimated Timeline:**
- Round 1: 2 hours (parallel)
- Round 2: 2 hours
- Round 3: 2 hours
- **Total: 6 hours**

---

## File Ownership Matrix

| Phase | File | Exclusive? | Purpose |
|-------|------|------------|---------|
| **Phase 1** | `processed-event-repository.go` | ✅ Yes | Add lock/unlock methods |
| **Phase 1** | `processed-event-repository_test.go` | ✅ Yes | Unit tests for lock methods |
| **Phase 1** | `ethereum.go` (shared/utils) | ✅ Yes | Add CRC32 hash function |
| **Phase 2** | `webhook_handler.go` | ✅ Yes | Integrate lock in webhook flow |
| **Phase 2** | `webhook_handler_test.go` | ✅ Yes | Unit tests for webhook handler |
| **Phase 3** | `main.go` | ✅ Yes | Add gRPC interceptor |
| **Phase 4** | `concurrent-webhook_test.go` | ✅ Yes | Integration tests |

**Conflict Prevention:** Each file modified by exactly ONE phase.

---

## Phase Files

- [Phase 1: Repository Lock Methods](./phase-01-repository-lock-methods.md) - Add advisory lock acquisition/release
- [Phase 2: Webhook Handler Integration](./phase-02-webhook-handler-integration.md) - Integrate locks into webhook flow
- [Phase 3: gRPC Interceptor Observability](./phase-03-grpc-interceptor-observability.md) - Add Sentry interceptor
- [Phase 4: Concurrent Integration Tests](./phase-04-concurrent-integration-tests.md) - Test concurrent webhook processing

---

## Context Links

- [Brainstorm Report](../../reports/brainstormer-260204-2205-context-canceled-webhook-fix.md) - Full problem analysis
- [Research: Advisory Locks](./research/researcher-01-advisory-locks.md) - PostgreSQL lock patterns
- [Research: gRPC Interceptor](./research/researcher-02-grpc-interceptor.md) - Interceptor patterns

---

## Key Implementation Details

### Advisory Lock Pattern

Use `pg_try_advisory_xact_lock()` with transaction-scoped locks:

```go
lockKey := int64(crc32.ChecksumIEEE([]byte(eventID)))
var acquired bool
err := tx.Raw("SELECT pg_try_advisory_xact_lock(?) AS acquired", lockKey).
    Scan(&acquired).Error
if !acquired {
    return ErrLockNotAcquired
}
```

**Why transaction-scoped:** Auto-releases on commit/rollback, connection-pool safe.

### gRPC Interceptor Pattern

```go
grpcServer := grpc.NewServer(
    grpc.MaxRecvMsgSize(10*1024*1024),
    grpc.MaxSendMsgSize(10*1024*1024),
    grpc.ChainUnaryInterceptor(
        grpcMiddleware.UnaryServerInterceptor(),
    ),
)
```

**Follows:** Same pattern as wallet-service, auth-service.

### Webhook Handler Integration

```go
// Acquire lock BEFORE processing
if err := s.processedEventRepo.AcquireEventLock(ctx, eventID); err != nil {
    if errors.Is(err, ErrLockNotAcquired) {
        return &pb.ProcessIndexerWebhookResponse{
            Success: true,
            Message: "Event already being processed",
        }, nil
    }
    return nil, err
}
```

**Idempotency:** Lock not acquired = another transaction processing = return success.

---

## Success Criteria

- [ ] 0% "context canceled" errors under concurrent webhook load (100 identical webhooks)
- [ ] 100% idempotency - exactly 1 event processes, 99 return "already processed"
- [ ] gRPC interceptor adds distributed tracing to all collection-service RPCs
- [ ] Advisory lock acquisition < 5ms (p95)
- [ ] All tests pass (unit + integration)
- [ ] Code compiles without errors

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Lock collision (CRC32) | LOW | MEDIUM | 1 in 4B collision rate, acceptable for webhooks |
| Deadlock | VERY LOW | HIGH | Single lock per event, no nested locks |
| Lock leak | VERY LOW | MEDIUM | Transaction-scoped locks auto-release |
| Performance regression | LOW | MEDIUM | Lock acquisition ~1-5μs, negligible overhead |

---

## Validation Log

### Session 1 — 2026-02-04
**Trigger:** Initial plan validation before implementation
**Questions asked:** 6

#### Questions & Answers

1. **[Architecture]** The plan uses transaction-scoped advisory locks (auto-release). Should we proceed with this or use session-scoped locks requiring explicit cleanup?
   - Options: Transaction-scoped | Session-scoped | Unsure
   - **Answer:** Transaction-scoped (after clarification)
   - **Rationale:** Transaction-scoped locks are safer for connection pools and auto-release on commit/rollback, preventing lock leaks.

2. **[Architecture]** Phase 3 adds Sentry interceptor. Does collection-service currently have Sentry config fields (DSN, Environment) defined, or should we add them?
   - Options: Add config now | Defer interceptor | Use empty config
   - **Answer:** Add config now
   - **Rationale:** Config needed for Sentry integration in Phase 3.

3. **[Testing]** Phase 4 integration tests require PostgreSQL (advisory locks not supported in SQLite). How should we handle test database setup?
   - Options: PostgreSQL required | Testcontainers | Use mocks only
   - **Answer:** PostgreSQL required
   - **Rationale:** Advisory locks are PostgreSQL-specific, need real database for accurate testing.

4. **[Architecture]** Phase 2 keeps the existing IsEventProcessed check alongside the new lock. Should we maintain both (defense-in-depth) or rely solely on the lock?
   - Options: Keep both checks | Lock only | Other approach
   - **Answer:** Keep both checks
   - **Rationale:** Defense-in-depth - lock prevents race condition, existing check prevents replay of old events.

5. **[Architecture]** When advisory lock is not acquired (another request processing the event), should we return success (idempotent) or error (trigger retry)?
   - Options: Return success | Return error | Retry with backoff
   - **Answer:** Return error
   - **Rationale:** Let indexer retry on lock contention rather than silently dropping duplicate events.

6. **[Architecture]** Important: The plan currently returns 'success' when lock not acquired. Returning 'error' would cause the indexer to retry. Which behavior do you want?
   - Options: Keep plan (return success) | Update plan (return error) | Make configurable
   - **Answer:** Update plan (return error)
   - **Rationale:** **PLAN CHANGE REQUIRED** - Return error to enable indexer retry on lock contention.

#### Confirmed Decisions
- Lock scope: Transaction-scoped (auto-release, connection-pool safe)
- Sentry config: Add to collection-service config structure
- Test database: Require PostgreSQL for integration tests
- Idempotency: Keep both lock + IsEventProcessed check (defense-in-depth)
- Lock contention behavior: Return error (not success) to trigger indexer retry

#### Action Items
- [ ] **UPDATE Phase 2:** Change lock not acquired response from success to error
- [ ] **UPDATE Phase 4:** Update tests to expect error on lock contention
- [ ] Add Sentry config to collection-service/internal/config/config.go
- [ ] Document indexer retry behavior

#### Impact on Phases
- **Phase 2 (Webhook Handler):** MUST UPDATE - Change ErrLockNotAcquired handling from returning success to returning error
- **Phase 4 (Integration Tests):** MUST UPDATE - Tests should expect errorCount > 0, not all successes
- **Phase 1 (Repository):** No change - lock method remains same
- **Phase 3 (gRPC Interceptor):** No change - Sentry config needs to be added

---

## Next Steps

1. **Review validation log** - Approve plan changes (error on lock contention)
2. **Update Phase 2** - Change lock not acquired response behavior
3. **Update Phase 4** - Update test expectations
4. **Execute Round 1** - Phase 1 (repository) + Phase 3 (gRPC) in parallel
5. **Execute Round 2** - Phase 2 (webhook handler) after Phase 1 completes
6. **Execute Round 3** - Phase 4 (integration tests) after all phases complete
7. **Deploy & Monitor** - Track metrics: context cancellation rate, lock acquisition time, retry rate

---

## Unresolved Questions

1. ~~Should we add Sentry DSN to collection-service config?~~ **RESOLVED: Add now**
2. What's acceptable lock contention threshold? (Suggest < 1% of requests)
3. Should we add monitoring metrics for lock acquisition time?
4. **NEW:** What's the indexer's retry policy for lock contention errors?
