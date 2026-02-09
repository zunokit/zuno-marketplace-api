# Planner Report: Advisory Lock Webhook Idempotency Fix

**Date:** 2026-02-04
**Planner:** planner agent
**Branch:** fix/chore-compare-address-collection
**Plan:** `plans/260204-2215-advisory-lock-webhook-fix/`

---

## Executive Summary

Created comprehensive parallel-optimized implementation plan for fixing "context canceled" errors in webhook processing using PostgreSQL advisory locks, gRPC observability interceptor, and integration tests.

**Problem:** Race condition in idempotency check allows concurrent duplicate events to cause database lock contention.

**Solution:** Advisory locks serialize event processing + gRPC interceptor adds observability.

**Effort:** 6 hours (4 phases)

---

## Plan Structure

```
E:/zuno-marketplace-api/plans/260204-2215-advisory-lock-webhook-fix/
├── plan.md                                    # Overview with dependency graph
├── phase-01-repository-lock-methods.md         # Lock/unlock implementation
├── phase-02-webhook-handler-integration.md     # Webhook handler with locks
├── phase-03-grpc-interceptor-observability.md  # gRPC interceptor
└── phase-04-concurrent-integration-tests.md    # Integration tests
```

---

## Parallelization Strategy

### Round 1: Parallel Execution (2 hours)

| Phase | Files | Description |
|-------|-------|-------------|
| **Phase 1** | `processed-event-repository.go`<br/>`ethereum.go` | Add advisory lock methods |
| **Phase 3** | `main.go` | Add gRPC interceptor |

**Why parallel:** No file overlap.

### Round 2: Sequential (2 hours)

| Phase | Files | Description |
|-------|-------|-------------|
| **Phase 2** | `webhook_handler.go` | Integrate locks into webhook flow |

**Blocks:** Depends on Phase 1 lock methods.

### Round 3: Sequential (2 hours)

| Phase | Files | Description |
|-------|-------|-------------|
| **Phase 4** | `concurrent-webhook_test.go` | Integration tests |

**Blocks:** Depends on complete implementation.

---

## File Ownership Matrix

| File | Phase | Purpose |
|------|-------|---------|
| `processed-event-repository.go` | Phase 1 | Lock acquisition methods |
| `processed-event-repository_test.go` | Phase 1 | Unit tests |
| `ethereum.go` (shared/utils) | Phase 1 | CRC32 hash function |
| `webhook_handler.go` | Phase 2 | Lock integration |
| `webhook_handler_test.go` | Phase 2 | Webhook unit tests |
| `main.go` | Phase 3 | gRPC interceptor |
| `concurrent-webhook_test.go` | Phase 4 | Integration tests |

**Conflict Prevention:** Each file modified by exactly ONE phase.

---

## Key Implementation Details

### Advisory Lock Pattern

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
    grpc.ChainUnaryInterceptor(
        grpcMiddleware.UnaryServerInterceptor(),
    ),
)
```

**Follows:** Same pattern as wallet-service, auth-service.

### Weblock Handler Integration

```go
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

---

## Success Criteria

- [ ] 0% "context canceled" errors under concurrent load
- [ ] 100% idempotency - exactly 1 event processes
- [ ] gRPC interceptor adds distributed tracing
- [ ] Advisory lock acquisition < 5ms (p95)
- [ ] All tests pass (unit + integration)

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| CRC32 collision | LOW | MEDIUM | 1 in 4B rate, acceptable |
| Deadlock | VERY LOW | HIGH | Single lock per event |
| Lock leak | VERY LOW | MEDIUM | Transaction-scoped auto-release |
| Performance regression | LOW | MEDIUM | Lock ~1-5μs overhead |

---

## Next Steps

1. **Review plan** - Approve phases and parallelization
2. **Execute Round 1** - Phase 1 + Phase 3 (parallel)
3. **Execute Round 2** - Phase 2 (after Phase 1)
4. **Execute Round 3** - Phase 4 (after all phases)
5. **Deploy & Monitor** - Track metrics

---

## References

- [Brainstorm Report](../reports/brainstormer-260204-2205-context-canceled-webhook-fix.md)
- [Research: Advisory Locks](../260204-2215-advisory-lock-webhook-fix/research/researcher-01-advisory-locks.md)
- [Research: gRPC Interceptor](../260204-2215-advisory-lock-webhook-fix/research/researcher-02-grpc-interceptor.md)

---

## Unresolved Questions

1. Should we add Sentry DSN to collection-service config? (Currently missing)
2. What's acceptable lock contention threshold? (Suggest < 1%)
3. Should we add monitoring metrics for lock acquisition time?
