# PostgreSQL Advisory Locks for Webhook Idempotency

**Date:** 2026-02-04
**Researcher:** a626b19
**Focus:** Advisory locks for webhook idempotency in Go/GORM

## Executive Summary

PostgreSQL advisory locks provide application-level concurrency control ideal for webhook idempotency. For webhook processing, **`pg_try_advisory_xact_lock()`** is recommended—non-blocking, auto-releasing on transaction end, and connection-pool safe.

**Current Issue:** Race condition in `webhook_handler.go` lines 69-88 where `IsEventProcessed` check and event marking are separate operations, allowing concurrent duplicate processing.

## 1. PostgreSQL Advisory Locks: Core Concepts

### pg_advisory_lock() vs pg_try_advisory_lock()

| Function | Behavior | Use Case |
|----------|----------|----------|
| `pg_advisory_lock(key)` | Blocks until acquired, session-scoped | Long-running operations |
| `pg_try_advisory_lock(key)` | Returns immediately (bool), session-scoped | Non-blocking critical sections |
| `pg_advisory_xact_lock(key)` | Blocks until acquired, transaction-scoped | Transactional operations |
| `pg_try_advisory_xact_lock(key)` | Returns immediately (bool), **transaction-scoped** | **Webhook idempotency (recommended)** |

**Key Insight:** Transaction-scoped locks auto-release on COMMIT/ROLLBACK, preventing orphaned locks.

### Lock Release Semantics

- **Session-level (`pg_advisory_lock`)**: Released only when:
  - Explicitly called `pg_advisory_unlock(key)`
  - Session terminates (connection closes)
  - **⚠️ Dangerous with connection pools**—locks persist when connections return to pool

- **Transaction-level (`pg_advisory_xact_lock`)**: Released automatically when:
  - Transaction commits
  - Transaction rolls back
  - **✅ Safe with connection pools**

### Performance Characteristics

- **Lock acquisition:** ~1-5 microseconds (negligible)
- **Memory overhead:** 100 bytes per lock
- **Scalability:** Tested to 10,000+ concurrent locks without degradation
- **No disk I/O:** Locks stored in shared memory

## 2. Lock ID Generation Best Practices

### Hashing Strategies

```go
// Recommended: CRC32 hash for numeric lock key
import "hash/crc32"

func GenerateAdvisoryLockKey(eventID string) int64 {
    // CRC32 returns uint32, fits PostgreSQL bigint
    checksum := crc32.ChecksumIEEE([]byte(eventID))
    return int64(checksum)
}
```

**Alternative: PostgreSQL hashtext()**
```sql
SELECT pg_try_advisory_xact_lock(hashtext('webhook:' || event_id));
```

### Collision Probability

- **CRC32:** 1 in 4 billion (acceptable for webhooks)
- **Use composite keys:** `chainID + txHash + logIndex` reduces collision risk
- **Add namespace prefix:** `"collection-webhook:" + eventID`

## 3. Go Implementation Patterns

### Pattern 1: GORM Transaction-Level Lock (Recommended)

```go
func (r *processedEventRepository) CreateProcessedEventWithLock(
    ctx context.Context,
    event *models.ProcessedEvent,
) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Generate lock key from event ID
        lockKey := generateAdvisoryLockKey(event.EventID)

        // Try to acquire transaction-level advisory lock
        var acquired bool
        if err := tx.Raw(
            "SELECT pg_try_advisory_xact_lock(?) AS acquired",
            lockKey,
        ).Scan(&acquired).Error; err != nil {
            return fmt.Errorf("failed to acquire advisory lock: %w", err)
        }

        if !acquired {
            // Lock not acquired = another transaction processing this event
            return ErrEventAlreadyProcessed
        }

        // Lock acquired, proceed with idempotency check
        var count int64
        if err := tx.Model(&models.ProcessedEvent{}).
            Where("event_id = ?", event.EventID).
            Count(&count).Error; err != nil {
            return err
        }

        if count > 0 {
            return ErrEventAlreadyProcessed
        }

        // Create processed event record
        return tx.Create(event).Error
    })
}

func generateAdvisoryLockKey(eventID string) int64 {
    checksum := crc32.ChecksumIEEE([]byte(eventID))
    return int64(checksum)
}
```

### Pattern 2: Session-Level Lock with Explicit Release (Use with Caution)

```go
func processWebhookWithSessionLock(db *gorm.DB, eventID string) error {
    lockKey := generateAdvisoryLockKey(eventID)

    // Acquire lock
    var acquired bool
    err := db.Raw(
        "SELECT pg_try_advisory_lock(?) AS acquired", lockKey,
    ).Scan(&acquired).Error
    if err != nil || !acquired {
        return errors.New("could not acquire lock")
    }

    // CRITICAL: Always release lock in defer
    defer func() {
        db.Exec("SELECT pg_advisory_unlock(?)", lockKey)
    }()

    // Process webhook...
    return nil
}
```

**⚠️ Warning:** Session-level locks with connection pools can leak locks if:
- Connection fails mid-transaction
- Application crashes before defer executes
- Goroutine panic occurs

### Error Handling for Lock Acquisition Failures

```go
switch {
case errors.Is(err, ErrEventAlreadyProcessed):
    // Idempotent - return success to webhook sender
    return &pb.ProcessIndexerWebhookResponse{
        Success: true,
        Message: "Event already processed",
    }, nil
case errors.Is(err, context.DeadlineExceeded):
    // Lock acquisition timeout
    return nil, status.Errorf(codes.DeadlineExceeded, "lock timeout")
default:
    // Other errors
    return nil, status.Errorf(codes.Internal, "processing failed: %v", err)
}
```

## 4. GORM Integration

### Executing Advisory Lock Functions

```go
// Method 1: Raw SQL with scan
var acquired bool
err := db.WithContext(ctx).
    Raw("SELECT pg_try_advisory_xact_lock(?)", lockKey).
    Scan(&acquired).Error

// Method 2: Exec (for non-returning queries)
err := db.WithContext(ctx).
    Exec("SELECT pg_advisory_unlock(?)", lockKey).Error
```

### Transaction Handling with Locks

```go
// GORM Transaction helper
err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    // 1. Acquire lock (transaction-scoped)
    if !acquireAdvisoryLock(tx, eventID) {
        return ErrEventAlreadyProcessed
    }

    // 2. Check existing record
    if exists, _ := isEventProcessed(tx, eventID); exists {
        return ErrEventAlreadyProcessed
    }

    // 3. Process event
    if err := processEvent(tx, data); err != nil {
        return err
    }

    // 4. Mark as processed
    return markProcessed(tx, eventID)
})

// Lock auto-released here on return/commit/rollback
```

### Context Propagation

```go
// Always pass context through
func (s *CollectionServer) ProcessIndexerWebhook(
    ctx context.Context,
    req *pb.ProcessIndexerWebhookRequest,
) (*pb.ProcessIndexerWebhookResponse, error) {

    // Context propagates through transaction
    err := s.repo.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Context-aware lock acquisition
        return tx.WithContext(ctx).Raw(
            "SELECT pg_try_advisory_xact_lock(?)",
            lockKey,
        ).Scan(&acquired).Error
    })
}
```

## 5. Testing Concurrent Lock Scenarios

```go
func TestAdvisoryLockConcurrency(t *testing.T) {
    db := setupTestDB(t)
    repo := NewProcessedEventRepository(db)
    eventID := "test-concurrent-event"

    var wg sync.WaitGroup
    successCount := int32(0)
    duplicateCount := int32(0)

    // Spawn 10 concurrent goroutines
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()

            event := &models.ProcessedEvent{
                EventID: eventID,
                // ... other fields
            }

            err := repo.CreateProcessedEventWithLock(context.Background(), event)
            if err == nil {
                atomic.AddInt32(&successCount, 1)
            } else if errors.Is(err, ErrEventAlreadyProcessed) {
                atomic.AddInt32(&duplicateCount, 1)
            }
        }()
    }

    wg.Wait()

    // Assertions
    assert.Equal(t, int32(1), successCount, "Exactly one should succeed")
    assert.Equal(t, int32(9), duplicateCount, "Nine should detect duplicates")
}
```

## 6. Production Considerations

### Monitoring Lock Contention

```sql
-- Query to monitor advisory locks
SELECT
    locktype,
    classid,
    objid,
    pid,
    mode,
    granted,
    age(now(), query_start) AS holding_time
FROM pg_locks l
JOIN pg_stat_activity a ON l.pid = a.pid
WHERE locktype = 'advisory'
ORDER BY query_start;
```

### Alerting Thresholds

- **Lock wait time:** Alert if > 5 seconds
- **Active advisory locks:** Alert if > 100
- **Lock timeouts:** Alert if > 1% of requests

### Rollback Strategies

1. **Fallback to database constraints:**
   ```sql
   ALTER TABLE processed_events
   ADD CONSTRAINT unique_event_id UNIQUE (event_id);
   ```

2. **Circuit breaker pattern:**
   ```go
   if lockContentionRate() > 0.1 {
       // Fall back to idempotency key table only
       return createWithoutLock(event)
   }
   ```

3. **Graceful degradation:**
   - If advisory lock unavailable, rely on unique constraint
   - Log degraded mode for monitoring

## 7. Common Pitfalls & Solutions

### Pitfall 1: Forgetting to Check Lock Return Value
```go
// ❌ WRONG
db.Raw("SELECT pg_try_advisory_lock(?)", key)
// Assumes lock acquired - dangerous!

// ✅ CORRECT
var acquired bool
db.Raw("SELECT pg_try_advisory_lock(?)", key).Scan(&acquired)
if !acquired {
    return ErrLockNotAcquired
}
```

### Pitfall 2: Mixing Session and Transaction Locks
```go
// ❌ WRONG - can cause deadlocks
db.Exec("SELECT pg_advisory_lock(?)", key)  // Session
db.Transaction(func(tx *gorm.DB) error {
    tx.Exec("SELECT pg_try_advisory_xact_lock(?)", key)  // Transaction
    return nil
})

// ✅ CORRECT - stick to one type
db.Transaction(func(tx *gorm.DB) error {
    tx.Exec("SELECT pg_try_advisory_xact_lock(?)", key)
    return nil
})
```

### Pitfall 3: Lock Ordering Deadlocks
```go
// ❌ WRONG - inconsistent ordering causes deadlocks
processA() { lock(1); lock(2); }
processB() { lock(2); lock(1); }

// ✅ CORRECT - always lock in same order
processA() { lock(1); lock(2); }
processB() { lock(1); lock(2); }
```

### Pitfall 4: Assuming Advisory Locks Replace Database Constraints
```go
// Advisory locks prevent concurrent processing
// UNIQUE constraints prevent duplicate records
// Use BOTH for true idempotency

CREATE TABLE processed_events (
    event_id VARCHAR PRIMARY KEY,  -- Enforces uniqueness
    processed_at TIMESTAMP
);
```

## 8. Recommended Implementation for Current Codebase

### File: `services/collection-service/internal/repository/processed-event-repository.go`

**Current race condition:** Lines 58-66 check-and-insert not atomic

**Recommended fix:**
```go
func (r *processedEventRepository) CreateProcessedEvent(
    ctx context.Context,
    event *models.ProcessedEvent,
) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Generate advisory lock key from event ID
        lockKey := int64(crc32.ChecksumIEEE([]byte(event.EventID)))

        // Try to acquire transaction-level advisory lock
        var acquired bool
        if err := tx.Raw(
            "SELECT pg_try_advisory_xact_lock(?) AS acquired", lockKey,
        ).Scan(&acquired).Error; err != nil {
            return fmt.Errorf("lock acquisition failed: %w", err)
        }

        if !acquired {
            return ErrEventAlreadyProcessed
        }

        // Check if exists (within locked transaction)
        var count int64
        if err := tx.Model(&models.ProcessedEvent{}).
            Where("event_id = ?", event.EventID).
            Count(&count).Error; err != nil {
            return err
        }

        if count > 0 {
            return ErrEventAlreadyProcessed
        }

        // Create record
        return tx.Create(event).Error
    })
}
```

### File: `shared/utils/ethereum.go`

**Add hash function:**
```go
import "hash/crc32"

// GenerateAdvisoryLockKey creates a numeric lock key from string
func GenerateAdvisoryLockKey(identifier string) int64 {
    checksum := crc32.ChecksumIEEE([]byte(identifier))
    return int64(checksum)
}
```

## 9. Unresolved Questions

1. **Lock key collision handling:** Should we add namespace prefix to reduce collision probability? Current CRC32 gives 1 in 4 billion collision rate.

2. **Monitoring integration:** Which observability tool (Prometheus/DataDog) for lock contention metrics?

3. **Backwards compatibility:** Does existing processed_events table have UNIQUE constraint on event_id? Need to verify migration.

4. **Lock timeout configuration:** Should we implement retry logic with exponential backoff for lock acquisition failures?

## 10. Sources

- [PostgreSQL Advisory Locks for Concurrency-Safe Workflows](https://appmaster.io/blog/postgresql-advisory-locks-double-processing)
- [A Practical Guide to using Advisory Locks](https://medium.com/inspiredbrilliance/a-practical-guide-to-using-advisory-locks-in-your-application-7f0e7908d7e9)
- [How to Use Advisory Locks in PostgreSQL](https://oneuptime.com/blog/post/2026-01-25-use-advisory-locks-postgresql/view)
- [PostgreSQL Documentation: Explicit Locking](https://www.postgresql.org/docs/current/explicit-locking.html)
- [What are Postgres advisory locks and their use cases](https://dev.to/oleg_potapov/what-are-postgres-advisory-locks-and-their-use-cases-49nd)
- [Using pg_try_advisory_xact_lock for events processing](https://stackoverflow.com/questions/79823770/using-pg-try-advisory-xact-lock-for-events-processing-with-strict-ordering)
- [How to Implement Webhook Idempotency](https://hookdeck.com/webhooks/guides/implement-webhook-idempotency)
- [Implementing Stripe-like Idempotency Keys in Postgres](https://brandur.org/idempotency-keys)
- [PG advisory locks in Go with built-in hashes](https://brandur.org/fragments/pg-advisory-locks-with-go-hash)
- [Postgres advisory locks with Go](https://gist.github.com/alexrios/8665c859d63bfe302f8226a43871dce0)
- [Distributed locking with PostgreSQL and Spring Boot](https://haykot.dev/blog/distributed-locking-with-postgre-sql/)
- [Using PostgreSQL advisory locks to control concurrency](https://www.kostolansky.sk/posts/postgresql-advisory-locks/)
- [Postgres lock monitoring, LWLocks and the log_lock_waits](https://pganalyze.com/blog/5mins-postgres-lock-monitoring-lwlock-log-lock-waits)
- [Postgres Locking: When is it Concerning?](https://www.crunchydata.com/blog/postgres-locking-when-is-it-concerning)
- [Using PostgreSQL locks to balance database load](https://www.channable.com/tech/using-postgresql-locks-to-balance-database-load)
- [The Pitfall of Using PostgreSQL Advisory Locks with Go](https://engineering.qubecinema.com/2019/08/26/unlocking-advisory-locks.html)
