# Phase 2: Webhook Handler Integration

**Date:** 2026-02-04
**Priority:** P0 (Critical path)
**Status:** Pending
**Estimated:** 2 hours

---

## Context Links

- [Brainstorm Report](../../reports/brainstormer-260204-2205-context-canceled-webhook-fix.md) - Race condition analysis
- [Research: Advisory Locks](./research/researcher-01-advisory-locks.md) - Lock patterns
- [Phase 1](./phase-01-repository-lock-methods.md) - Lock method implementation
- [Main Plan](./plan.md) - Overview

---

## Parallelization Info

**Can run in parallel with:** None (depends on Phase 1)

**Blocks:** Phase 4 (Integration Tests)

**Dependency:** Requires Phase 1 lock methods to be implemented first.

---

## Overview

Integrate advisory lock acquisition into webhook handler to serialize concurrent duplicate events. Lock acquired BEFORE idempotency check to prevent race condition.

**Key Change:** Move lock acquisition to line 69 (before existing idempotency check).

---

## Key Insights from Research

1. **Lock BEFORE idempotency check** - prevents both duplicates from passing check
2. **Lock not acquired = idempotent response** - return success without processing
3. **Transaction-scoped lock** - auto-releases when function returns
4. **Idempotency preserved** - duplicate events return "already being processed"

---

## Requirements

### Functional

- Acquire advisory lock before processing event
- Return success response if lock not acquired (idempotent)
- Process event only if lock acquired
- Mark as processed after successful handling

### Non-Functional

- No performance regression (lock < 5ms)
- Thread-safe for concurrent webhooks
- Backward compatible response format

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              ProcessIndexerWebhook                          │
│  1. Parse event data                                        │
│  2. Validate payload                                        │
│  3. Build eventID                                           │
│  4. ⭐ ACQUIRE LOCK ⭐                                       │
│     ├─ Acquired? → Continue                                 │
│     └─ Not acquired? → Return "Already processing"          │
│  5. Check if processed (existing check)                     │
│  6. Handle event (collection.created/minted)                │
│  7. Mark as processed                                       │
│  8. Transaction commits → Lock auto-releases                │
└─────────────────────────────────────────────────────────────┘
```

**Race Condition Prevention:**
- Both concurrent requests call `AcquireEventLock()`
- First request acquires lock, proceeds
- Second request gets `ErrLockNotAcquired`, returns success
- No concurrent processing = no lock contention

---

## Related Code Files

**References:**
- `services/collection-service/internal/server/webhook_handler.go` - Main webhook handler
- `services/collection-service/internal/repository/processed-event-repository.go` - Lock methods (Phase 1)
- `services/collection-service/internal/models/processed-event-model.go` - Event model

**Creates:**
- `services/collection-service/internal/server/webhook_handler_test.go` - Unit tests

---

## File Ownership (EXCLUSIVE)

This phase EXCLUSIVELY modifies:

1. **`services/collection-service/internal/server/webhook_handler.go`**
   - Add lock acquisition at line 69
   - Handle `ErrLockNotAcquired` error
   - Keep existing idempotency check as defense-in-depth

2. **`services/collection-service/internal/server/webhook_handler_test.go`** (CREATE)
   - Unit tests for webhook with locks
   - Concurrent webhook tests

**Conflict Prevention:** No other phases modify webhook_handler.go.

---

## Implementation Steps

### Step 1: Import Lock Error

**File:** `services/collection-service/internal/server/webhook_handler.go`

Add import for repository errors:
```go
import (
    "github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/repository"
    // ... existing imports
)
```

### Step 2: Add Lock Acquisition

**Location:** After line 67 (after `eventID := BuildEventID(...)`)

**Add:**
```go
// 4. Acquire advisory lock to prevent concurrent processing
if err := s.processedEventRepo.AcquireEventLock(ctx, eventID); err != nil {
    if errors.Is(err, repository.ErrLockNotAcquired) {
        // Another transaction is processing this event - return error to trigger retry
        s.logger.Info("Event already being processed by another transaction",
            zap.String("event_id", eventID),
            zap.String("event_type", req.Event),
        )
        return nil, status.Errorf(codes.Aborted, "event lock not acquired: another transaction processing this event")
    }
    return nil, status.Errorf(codes.Internal, "failed to acquire lock: %v", err)
}
```

**Why return error:** Let indexer retry on lock contention rather than silently dropping the event.

**Note:** Using `codes.Aborted` indicates the operation should be retried.

### Step 3: Keep Existing Idempotency Check

**Do NOT remove** the existing `IsEventProcessed()` check (lines 70-88).

**Reason:** Defense-in-depth. Lock prevents race condition, existing check prevents replay of old events.

### Step 4: Update Error Handling

**No changes needed** to existing error handling. Lock acquisition error already handled in Step 2.

### Step 5: Create Unit Tests

**File:** `services/collection-service/internal/server/webhook_handler_test.go` (NEW)

```go
package server

import (
    "context"
    "encoding/json"
    "sync"
    "sync/atomic"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
    "go.uber.org/zap"
)

// Mock repository for testing
type MockProcessedEventRepo struct {
    mock.Mock
}

func (m *MockProcessedEventRepo) AcquireEventLock(ctx context.Context, eventID string) error {
    args := m.Called(ctx, eventID)
    return args.Error(0)
}

func (m *MockProcessedEventRepo) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
    args := m.Called(ctx, eventID)
    return args.Bool(0), args.Error(1)
}

func (m *MockProcessedEventRepo) CreateProcessedEvent(ctx context.Context, event *ProcessedEvent) error {
    args := m.Called(ctx, event)
    return args.Error(0)
}

// ... other interface methods ...

func TestProcessIndexerWebhook_ConcurrentDuplicates(t *testing.T) {
    // Setup mock repo
    mockRepo := new(MockProcessedEventRepo)
    mockCollectionSvc := new(MockCollectionService)

    server := &CollectionServer{
        processedEventRepo: mockRepo,
        service:            mockCollectionSvc,
        logger:             zap.NewNop(),
    }

    // Create webhook request
    eventData := WebhookEventData{
        CollectionAddress: "0xabc123",
        Creator:           "0xdef456",
        TokenType:         "ERC721",
        BlockNumber:       12345,
        TxHash:            "0xtx789",
        LogIndex:          0,
    }
    dataJson, _ := json.Marshal(eventData)

    req := &pb.ProcessIndexerWebhookRequest{
        Event:      "collection.created",
        ChainId:    1,
        Timestamp:  time.Now().Unix(),
        DataJson:   string(dataJson),
    }

    var wg sync.WaitGroup
    successCount := int32(0)
    lockContentionCount := int32(0)

    // Simulate 10 concurrent identical webhooks
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()

            // Mock: First call acquires lock, rest get ErrLockNotAcquired
            callNumber := atomic.AddInt32(&callCount, 1)
            if callNumber == 1 {
                mockRepo.On("AcquireEventLock", mock.Anything, mock.Anything).Return(nil)
            } else {
                mockRepo.On("AcquireEventLock", mock.Anything, mock.Anything).
                    Return(repository.ErrLockNotAcquired)
            }

            resp, err := server.ProcessIndexerWebhook(context.Background(), req)

            if err == nil && resp.Success {
                atomic.AddInt32(&successCount, 1)
            }
        }()
    }

    wg.Wait()

    // All should return success (idempotent)
    assert.Equal(t, int32(10), successCount, "All concurrent webhooks should return success")
}
```

---

## Todo List

- [ ] Import repository package in webhook_handler.go
- [ ] Add lock acquisition after eventID generation (line 69)
- [ ] Handle `ErrLockNotAcquired` - return success response
- [ ] Verify existing idempotency check remains (defense-in-depth)
- [ ] Create webhook_handler_test.go
- [ ] Add unit test for single webhook processing
- [ ] Add unit test for concurrent duplicate webhooks
- [ ] Add unit test for lock acquisition failure
- [ ] Run `go test ./services/collection-service/internal/server/`
- [ ] Verify code compiles: `go build ./services/collection-service/...`

---

## Success Criteria

- [ ] Lock acquired before event processing
- [ ] `ErrLockNotAcquired` returns `codes.Aborted` error (triggers indexer retry)
- [ ] Existing idempotency check preserved
- [ ] All unit tests pass
- [ ] Code compiles without errors
- [ ] Concurrent webhooks don't cause "context canceled" errors
- [ ] Indexer successfully retries on lock contention

---

## Conflict Prevention

**Files modified exclusively by this phase:**
- `webhook_handler.go` - Only Phase 2 touches this
- `webhook_handler_test.go` - New file, no conflicts

**No overlap with:**
- Phase 1 (repository layer)
- Phase 3 (main.go)
- Phase 4 (integration tests - different directory)

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Lock held too long | Lock only during DB transaction (~ms) |
| Lock starvation | Very unlikely - lock held for < 100ms |
| Response format change | Preserve existing response structure |
| Breaking change | No - lock acquisition is internal implementation |

---

## Security Considerations

- Lock acquisition uses same eventID as idempotency check
- No new attack surface - lock is database-internal
- Idempotency preserved - attackers can't replay events

---

## Next Steps

**After this phase completes:**
1. Phase 4 (Integration Tests) can test complete flow
2. Manual testing with concurrent webhooks
3. Deploy to staging for load testing

**Dependency:** Requires Phase 1 (lock methods) to complete first.
