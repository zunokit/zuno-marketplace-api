# PHASE 5: BLOCKCHAIN INDEXER INTEGRATION

**Phase**: 5 of 6
**Dependencies**: Phase 2-3 complete
**Estimated Effort**: ~500 lines Go
**Output**: Webhook endpoints at `/webhooks/*`

---

## 🎯 Overview

Integrate with blockchain indexer (Ponder) to sync on-chain events:
- ✅ Webhook endpoints to receive events
- ✅ HMAC signature verification
- ✅ Idempotency handling
- ✅ Collection status updates
- ✅ Event payload validation

---

## 📋 Implementation Checklist

### Step 1: Webhook Endpoints ✅
**File**: `services/graphql-gateway/internal/handlers/webhook_handler.go`

**Endpoints to create**:
```go
POST /webhooks/collection-created
POST /webhooks/nft-minted
POST /webhooks/nft-transferred
```

**Payload structure**:
```json
{
  "eventType": "CollectionCreated",
  "contractAddress": "0x123...",
  "chainId": "eip155:11155111",
  "blockNumber": 12345678,
  "transactionHash": "0xabc...",
  "logIndex": 0,
  "data": {
    "collectionAddress": "0x456...",
    "creator": "0x789...",
    "tokenStandard": "ERC721"
  },
  "timestamp": "2025-11-20T10:30:00Z"
}
```

---

### Step 2: HMAC Signature Verification ✅
**Implementation**:

```go
func validateWebhookSignature(payload []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedSig := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedSig))
}

func (h *WebhookHandler) HandleCollectionCreated(w http.ResponseWriter, r *http.Request) {
    // 1. Read raw body
    body, _ := io.ReadAll(r.Body)

    // 2. Verify signature
    sig := r.Header.Get("X-Webhook-Signature")
    if !validateWebhookSignature(body, sig, os.Getenv("INDEXER_WEBHOOK_SECRET")) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }

    // 3. Parse payload
    // 4. Process event
}
```

**Environment variable**: `INDEXER_WEBHOOK_SECRET`

---

### Step 3: Event Processing ✅
**Logic**:

```go
func (h *WebhookHandler) processCollectionCreated(event *CollectionCreatedEvent) error {
    // 1. Find collection by creator address + chain
    collection, err := h.collectionRepo.GetByCreatorAndChain(
        event.Data.Creator,
        event.ChainID,
    )

    // 2. Update collection with contract address
    updates := map[string]interface{}{
        "contract_address": event.Data.CollectionAddress,
        "chain_id":         event.ChainID,
        "deployed_block":   event.BlockNumber,
        "status":           "DEPLOYED",
        "deployed_at":      event.Timestamp,
        "index_status":     "SYNCING",
    }

    // 3. Apply updates
    err = h.collectionRepo.Update(ctx, collection.ID, updates)

    // 4. Log activity (trigger will auto-log STATUS_CHANGED)

    return nil
}
```

---

### Step 4: Idempotency Handling ✅
**Table**: `processed_events` (already exists in schema)

```go
func (h *WebhookHandler) isEventProcessed(eventID string) bool {
    var count int64
    h.db.Model(&ProcessedEvent{}).
        Where("event_id = ?", eventID).
        Count(&count)
    return count > 0
}

func (h *WebhookHandler) markEventProcessed(event *WebhookEvent) error {
    return h.db.Create(&ProcessedEvent{
        EventID:     event.TransactionHash + ":" + strconv.Itoa(event.LogIndex),
        EventVersion: 1,
        ChainID:      event.ChainID,
        BlockHash:    event.BlockHash,
        LogIndex:     event.LogIndex,
        ProcessedAt:  time.Now(),
    }).Error
}
```

---

### Step 5: Register Routes ✅
**File**: `services/graphql-gateway/cmd/main.go`

```go
webhookHandler := handlers.NewWebhookHandler(collectionRepo