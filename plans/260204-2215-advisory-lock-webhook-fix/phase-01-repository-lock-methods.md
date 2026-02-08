# Phase 1: Repository Lock Methods

**Date:** 2026-02-04
**Priority:** P0 (Critical path)
**Status:** Pending
**Estimated:** 2 hours

---

## Context Links

- [Brainstorm Report](../../reports/brainstormer-260204-2205-context-canceled-webhook-fix.md) - Root cause analysis
- [Research: Advisory Locks](./research/researcher-01-advisory-locks.md) - PostgreSQL lock patterns
- [Main Plan](./plan.md) - Overview and dependencies

---

## Parallelization Info

**Can run in parallel with:** Phase 3 (gRPC Interceptor)

**Reason:** No file overlap. This phase modifies repository layer only.

**Blocks:** Phase 2 (Webhook Handler) depends on lock methods

---

## Overview

Add PostgreSQL advisory lock acquisition/release methods to `ProcessedEventRepository`. Uses transaction-scoped locks (`pg_try_advisory_xact_lock`) for connection-pool safety.

**Key Change:** Lock auto-releases on transaction commit/rollback - no explicit unlock needed.

---

## Key Insights from Research

1. **Transaction-scoped locks are safer** than session-scoped for connection pools
2. **CRC32 hash** provides 1 in 4B collision rate (acceptable for webhooks)
3. **Must check lock return value** - `pg_try_advisory_xact_lock` returns bool
4. **Lock acquisition cost:** ~1-5 microseconds (negligible)

---

## Requirements

### Functional

- Add `AcquireEventLock(ctx, eventID)` method to repository interface
- Use CRC32 hash to generate lock key from eventID
- Return `ErrLockNotAcquired` when lock already held
- Transaction-scoped lock (auto-releases)

### Non-Functional

- Lock acquisition < 5ms (p95)
- Thread-safe for concurrent goroutines
- Connection-pool safe

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  ProcessIndexerWebhook                       │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│         AcquireEventLock(eventID)                            │
│  1. Generate lock key: CRC32(eventID)                        │
│  2. Execute: SELECT pg_try_advisory_xact_lock(?)             │
│  3. Check return value (bool)                                │
│  4. Return nil or ErrLockNotAcquired                         │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
              ┌──────────┴──────────┐
              │                     │
        Lock Acquired          Lock Not Acquired
              │                     │
              ▼                     ▼
     Process Event         Return "Already Processing"
              │
              ▼
    Transaction Commits → Lock Auto-Released
```

---

## Related Code Files

**References:**
- `services/collection-service/internal/models/processed-event-model.go` - Event model
- `shared/utils/ethereum.go` - Address utils (add hash function here)

**Creates:**
- `services/collection-service/internal/repository/processed-event-repository_test.go` - Unit tests

---

## File Ownership (EXCLUSIVE)

This phase EXCLUSIVELY modifies:

1. **`services/collection-service/internal/repository/processed-event-repository.go`**
   - Add `AcquireEventLock()` method
   - Add `ErrLockNotAcquired` error
   - Update interface

2. **`services/collection-service/internal/repository/processed-event-repository_test.go`** (CREATE)
   - Unit tests for lock methods
   - Concurrent lock acquisition tests

3. **`shared/utils/ethereum.go`**
   - Add `GenerateAdvisoryLockKey()` function
   - CRC32 hash implementation

**Conflict Prevention:** No other phases modify these files.

---

## Implementation Steps

### Step 1: Add Lock Key Generation to shared/utils/ethereum.go

**File:** `shared/utils/ethereum.go`

```go
import (
    "hash/crc32"
    // ... existing imports
)

// GenerateAdvisoryLockKey creates a numeric lock key from string identifier
// Uses CRC32 checksum for consistent distribution across lock space
func GenerateAdvisoryLockKey(identifier string) int64 {
    checksum := crc32.ChecksumIEEE([]byte(identifier))
    return int64(checksum)
}
```

**Why shared utils:** Other services may need advisory locks later.

### Step 2: Add Lock Methods to Repository

**File:** `services/collection-service/internal/repository/processed-event-repository.go`

```go
var (
    ErrEventAlreadyProcessed = errors.New("event already processed")
    ErrLockNotAcquired       = errors.New("advisory lock not acquired")
)

type ProcessedEventRepository interface {
    // ... existing methods ...

    // AcquireEventLock acquires PostgreSQL advisory lock for event
    // Returns ErrLockNotAcquired if lock already held by another transaction
    AcquireEventLock(ctx context.Context, eventID string) error
}
```

**Implementation:**

```go
// AcquireEventLock acquires PostgreSQL advisory lock for event serialization
// Uses transaction-scoped lock (pg_try_advisory_xact_lock) which auto-releases
// on transaction commit/rollback. Safe for connection pools.
func (r *processedEventRepository) AcquireEventLock(ctx context.Context, eventID string) error {
    // Generate lock key from event ID using CRC32 hash
    lockKey := utils.GenerateAdvisoryLockKey(eventID)

    // Try to acquire transaction-level advisory lock
    var acquired bool
    err := r.db.WithContext(ctx).Raw(
        "SELECT pg_try_advisory_xact_lock(?) AS acquired", lockKey,
    ).Scan(&acquired).Error

    if err != nil {
        return fmt.Errorf("lock acquisition failed: %w", err)
    }

    // Lock not acquired = another transaction is processing this event
    if !acquired {
        return ErrLockNotAcquired
    }

    return nil
}
```

**Key Points:**
- `WithContext(ctx)` ensures context cancellation propagates
- Transaction-scoped lock = auto-release on commit/rollback
- Check `acquired` bool to handle contention gracefully

### Step 3: Create Unit Tests

**File:** `services/collection-service/internal/repository/processed-event-repository_test.go` (NEW)

```go
package repository

import (
    "context"
    "sync"
    "sync/atomic"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    assert.NoError(t, err)
    db.Exec("CREATE TABLE processed_events (id TEXT PRIMARY KEY, event_id TEXT UNIQUE)")
    return db
}

func TestAcquireEventLock_SingleAcquisition(t *testing.T) {
    db := setupTestDB(t)
    repo := NewProcessedEventRepository(db)

    err := repo.AcquireEventLock(context.Background(), "test-event-1")
    assert.NoError(t, err, "First acquisition should succeed")
}

func TestAcquireEventLock_ConcurrentSameEvent(t *testing.T) {
    db := setupTestDB(t)
    repo := NewProcessedEventRepository(db)
    eventID := "concurrent-event-123"

    var wg sync.WaitGroup
    successCount := int32(0)
    lockNotAcquiredCount := int32(0)

    // Spawn 10 goroutines trying to acquire same lock
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            err := repo.AcquireEventLock(context.Background(), eventID)
            if err == nil {
                atomic.AddInt32(&successCount, 1)
            } else if err == ErrLockNotAcquired {
                atomic.AddInt32(&lockNotAcquiredCount, 1)
            }
        }()
    }

    wg.Wait()

    // With SQLite (no advisory locks), all succeed
    // With PostgreSQL, exactly 1 succeeds, 9 get ErrLockNotAcquired
    assert.True(t, successCount >= 1, "At least one should acquire lock")
}

func TestAcquireEventLock_DifferentEvents(t *testing.T) {
    db := setupTestDB(t)
    repo := NewProcessedEventRepository(db)

    // Different events should not block each other
    err1 := repo.AcquireEventLock(context.Background(), "event-1")
    err2 := repo.AcquireEventLock(context.Background(), "event-2")

    assert.NoError(t, err1, "Event 1 should acquire lock")
    assert.NoError(t, err2, "Event 2 should acquire lock (different key)")
}
```

---

## Todo List

- [ ] Add `GenerateAdvisoryLockKey()` to `shared/utils/ethereum.go`
- [ ] Add `ErrLockNotAcquired` to repository
- [ ] Add `AcquireEventLock()` to interface
- [ ] Implement `AcquireEventLock()` method with transaction-scoped lock
- [ ] Create `processed-event-repository_test.go`
- [ ] Add unit test for single lock acquisition
- [ ] Add unit test for concurrent lock (same event)
- [ ] Add unit test for concurrent lock (different events)
- [ ] Run `go test ./services/collection-service/internal/repository/`
- [ ] Verify code compiles: `go build ./services/collection-service/...`

---

## Success Criteria

- [ ] `GenerateAdvisoryLockKey()` added to shared utils
- [ ] `AcquireEventLock()` method implemented
- [ ] All unit tests pass
- [ ] Code compiles without errors
- [ ] Lock acquisition returns `ErrLockNotAcquired` when contention occurs

---

## Conflict Prevention

**Files modified exclusively by this phase:**
- `processed-event-repository.go` - Only Phase 1 touches this
- `ethereum.go` - Only Phase 1 touches this
- `processed-event-repository_test.go` - New file, no conflicts

**No overlap with:**
- Phase 2 (webhook_handler.go)
- Phase 3 (main.go)
- Phase 4 (integration tests)

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| SQLite tests don't validate advisory locks | Document that PostgreSQL required for real lock behavior |
| CRC32 collision (1 in 4B) | Acceptable rate for webhooks, same event always hashes to same key |
| Lock not releasing | Transaction-scoped locks auto-release on commit/rollback |

---

## Security Considerations

- Lock key uses CRC32 of eventID - no sensitive data exposed
- Advisory locks are database-scoped - no external access
- No SQL injection - uses parameterized query

---

## Next Steps

**After this phase completes:**
1. Phase 2 can integrate lock into webhook handler
2. Phase 3 can run in parallel (no dependency)
3. Phase 4 waits for Phase 2 completion

**Dependency:** Phase 2 (Webhook Handler) requires this phase's lock methods.
