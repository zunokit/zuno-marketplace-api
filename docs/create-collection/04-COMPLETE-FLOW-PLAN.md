# Complete Create Collection Flow - Architecture & Implementation Plan

## Overview

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│ Frontend │───►│ Backend  │    │   SDK    │───►│ Contract │───►│ Indexer  │
│  (Next)  │    │ (Go gRPC)│    │ (Viem)   │    │ (Solidity)│    │ (Ponder) │
└──────────┘    └────┬─────┘    └──────────┘    └──────────┘    └─────┬────┘
                     │                                                  │
                     │                                                  │
                     └──────────────────────────────────────────────────┘
                                   Webhook (HTTP POST)
```

---

## Current State Analysis

### ✅ Already Implemented

#### Backend (zuno-marketplace-api)
- ✅ Collection Service (gRPC + GraphQL)
  - Create collection in database (status: PENDING)
  - Update collection with contract address
  - Add allowlist to database
  - Query collections, my collections
- ✅ Media Service
  - Upload images to Metadata Service (ImageKit + IPFS)
- ✅ GraphQL Gateway
  - Mutations: createCollection, updateCollection, addToAllowlist
  - Queries: collection, myCollections, collections
- ✅ Database Schema
  - collections, collection_metadata, collection_stats, collection_allowlist tables

#### Frontend (zuno-marketplace-ui)
- ✅ CollectionForm component (Magic Eden style)
- ✅ useCreateCollection hook
  - Step 1: Upload media
  - Step 2: Create in database
  - Step 3: Add allowlist to database
- ✅ GraphQL hooks (createCollection, updateCollection, addToAllowlist)
- ✅ Upload client for media
- ✅ My Collections page

#### Smart Contracts (zuno-marketplace-contracts)
- ✅ ERC721CollectionFactory
- ✅ ERC1155CollectionFactory
- ✅ Collection contracts with:
  - Mint functionality
  - Allowlist support
  - Mint stages (INACTIVE → ALLOWLIST → PUBLIC)
  - Royalty support

#### SDK (zuno-marketplace-sdk)
- ✅ React hooks: useCollection
- ✅ Deploy methods:
  - createERC721Collection(params)
  - createERC1155Collection(params)
- ✅ Returns: contract address, transaction hash

### ❌ Not Yet Implemented

#### Indexer (zuno-marketplace-indexer)
- ❌ Webhook system (no HTTP POST to backend)
- ❌ Dynamic contract registration (for mint events)
- ❌ Webhook delivery tracking
- ✅ Event indexing (collection creation, minting)
- ✅ Event storage in database
- ✅ REST/GraphQL APIs

#### Frontend
- ❌ SDK integration in useCreateCollection
- ❌ Step 4: Deploy contract via SDK
- ❌ Step 5: Update database with contract address

#### Backend
- ❌ Webhook endpoint to receive indexer events
- ❌ Event processing logic

---

## Complete Flow Architecture

### Phase 1: Create Collection (Database Only)

```
User fills form
     │
     ▼
┌──────────────────────────────────────────────────────────────┐
│ FRONTEND: useCreateCollection hook                           │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  Step 1: Upload Images                                       │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ uploadFile(collectionImage)                         │    │
│  │      │                                               │    │
│  │      ▼                                               │    │
│  │ POST /api/upload/media                              │────┼──► Media Service (gRPC)
│  │                                                       │    │         │
│  │                                                       │    │         ▼
│  │ Result: imageUrl (ImageKit)                          │    │    Metadata Service
│  └─────────────────────────────────────────────────────┘    │    (ImageKit + IPFS)
│                                                               │
│  Step 2: Create Collection in Database                       │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ GraphQL Mutation: createCollection                  │────┼──► GraphQL Gateway
│  │   input: {                                           │    │         │
│  │     name, symbol, tokenStandard,                    │    │         ▼
│  │     chainId, deployerAddress,                       │    │    Collection Service (gRPC)
│  │     imageUrl, baseUri, maxSupply,                   │    │         │
│  │     mintPrice, royaltyFeeBps, ...                   │    │         ▼
│  │   }                                                  │    │    PostgreSQL
│  │                                                       │    │    collections table
│  │ Result: { id, status: 'PENDING' }                   │    │    status = PENDING
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  Step 3: Add Allowlist to Database                           │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ GraphQL Mutation: addToAllowlist                    │────┼──► GraphQL Gateway
│  │   input: {                                           │    │         │
│  │     collectionId,                                    │    │         ▼
│  │     walletAddresses: [...],                         │    │    Collection Service
│  │     maxMintAmount                                    │    │         │
│  │   }                                                  │    │         ▼
│  │                                                       │    │    collection_allowlist table
│  │ Result: true                                         │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
└──────────────────────────────────────────────────────────────┘

Database State:
  collections table:
    - id: uuid
    - status: 'PENDING'
    - contract_address: null
    - deployed_at: null

  collection_allowlist table:
    - collection_id: uuid
    - wallet_address: '0x...'
    - max_mint_amount: 5
```

### Phase 2: Deploy Contract (Blockchain)

```
┌──────────────────────────────────────────────────────────────┐
│ FRONTEND: useCreateCollection hook (continued)               │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  Step 4: Deploy Smart Contract                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ import { useCollection } from 'zuno-marketplace-sdk' │    │
│  │                                                       │    │
│  │ const { createERC721 } = useCollection();           │    │
│  │                                                       │    │
│  │ const result = await createERC721.mutateAsync({     │    │
│  │   name: 'My Collection',                            │    │
│  │   symbol: 'MYC',                                     │    │
│  │   baseUri: 'https://metadata.com/...',              │    │
│  │   maxSupply: 10000,                                  │    │
│  │   options: {                                         │    │
│  │     onSent: (txHash) => {                           │    │
│  │       // Show tx hash in UI                         │    │
│  │     }                                                │    │
│  │   }                                                  │    │
│  │ });                                                  │    │
│  │                                                       │    │
│  │ // User signs transaction in wallet popup           │◄───┼─── MetaMask/RainbowKit
│  │                                                       │    │     (User approval)
│  │                                                       │    │
│  │ Result:                                              │    │
│  │   - address: '0xABC...' (contract address)          │    │
│  │   - tx.hash: '0xDEF...'                             │    │
│  └─────────────────────────────────────────────────────┘    │
│                          │                                    │
│                          ▼                                    │
│                  Blockchain Transaction                       │
│                          │                                    │
│                          ▼                                    │
│      ┌────────────────────────────────────────┐             │
│      │ ERC721CollectionFactory.createClone()  │             │
│      │                                         │             │
│      │ 1. Deploy ERC721Collection proxy       │             │
│      │ 2. Initialize with parameters          │             │
│      │ 3. Set owner = creator                 │             │
│      │ 4. Emit ERC721CollectionCreated event  │             │
│      └────────────────────────────────────────┘             │
│                          │                                    │
│                          ▼                                    │
│                   Contract Deployed                           │
│                   address: 0xABC...                          │
│                                                               │
│  Step 5: Update Database with Contract Address               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ GraphQL Mutation: updateCollection                  │────┼──► GraphQL Gateway
│  │   id: collectionId                                   │    │         │
│  │   input: {                                           │    │         ▼
│  │     contractAddress: '0xABC...',                    │    │    Collection Service
│  │     status: 'DEPLOYED',                             │    │         │
│  │     deployedAt: '2025-01-22T...'                    │    │         ▼
│  │   }                                                  │    │    UPDATE collections
│  │                                                       │    │    SET contract_address = '0xABC...'
│  │ Result: { id, status: 'DEPLOYED' }                  │    │        status = 'DEPLOYED'
│  └─────────────────────────────────────────────────────┘    │        deployed_at = NOW()
│                                                               │
│  SUCCESS: Redirect to /collections/{id}                      │
│                                                               │
└──────────────────────────────────────────────────────────────┘

Database State:
  collections table:
    - id: uuid
    - status: 'DEPLOYED'
    - contract_address: '0xABC...'
    - deployed_at: '2025-01-22T10:30:00Z'
    - chain_id: 'eip155:11155111'
```

### Phase 3: Indexer Detects and Indexes Events

```
┌──────────────────────────────────────────────────────────────┐
│ BLOCKCHAIN: Events Emitted                                   │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  Event 1: ERC721CollectionCreated                            │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ Event: ERC721CollectionCreated                      │    │
│  │ Address: 0xFACTORY_ADDRESS                          │    │
│  │ Topics:                                              │    │
│  │   [0]: event signature hash                         │    │
│  │   [1]: collectionAddress (indexed)                  │    │
│  │   [2]: creator (indexed)                            │    │
│  │                                                       │    │
│  │ Data: (none for this event)                         │    │
│  │                                                       │    │
│  │ Block: 12345678                                      │    │
│  │ Transaction: 0xDEF...                                │    │
│  │ Timestamp: 1737540600                                │    │
│  └─────────────────────────────────────────────────────┘    │
│                          │                                    │
│                          ▼                                    │
└──────────────────────────┼───────────────────────────────────┘
                           │
                           │ RPC Polling / WebSocket
                           │
                           ▼
┌──────────────────────────────────────────────────────────────┐
│ INDEXER: Ponder Event Processing                             │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  Event Handler: erc721-created.handler.ts                    │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ 1. Validate event data (Zod schema)                 │    │
│  │                                                       │    │
│  │ 2. Store in event table (event-first architecture)  │────┼──► PostgreSQL
│  │    INSERT INTO event {                               │    │     event table
│  │      event_type: 'collection_created',              │    │     (source of truth)
│  │      category: 'collection',                        │    │
│  │      actor: creator_address,                        │    │
│  │      collection: collection_address,                │    │
│  │      chain_id: 11155111,                            │    │
│  │      data: { name, symbol, tokenStandard },         │    │
│  │      tx_hash: '0xDEF...',                           │    │
│  │      block_number: 12345678,                         │    │
│  │      timestamp: 1737540600                           │    │
│  │    }                                                 │    │
│  │                                                       │    │
│  │ 3. Update account cache                              │    │
│  │    UPDATE account                                    │────┼──► account table
│  │    SET last_active_at = NOW(),                      │    │
│  │        event_count = event_count + 1                │    │
│  │    WHERE address = creator                          │    │
│  │                                                       │    │
│  │ 4. [NEW] Trigger webhook (TO BE IMPLEMENTED)        │    │
│  │    webhookClient.sendWebhook({                      │    │
│  │      event: 'collection.created',                   │    │
│  │      chainId: 11155111,                             │    │
│  │      data: {                                         │    │
│  │        collectionAddress: '0xABC...',               │    │
│  │        creator: '0x123...',                         │    │
│  │        tokenType: 'ERC721',                         │    │
│  │        blockNumber: 12345678,                        │    │
│  │        txHash: '0xDEF...'                           │    │
│  │      }                                               │    │
│  │    })                                                │    │
│  └─────────────────────────────────────────────────────┘    │
│                          │                                    │
│                          ▼                                    │
│                    Webhook Triggered                          │
│                          │                                    │
└──────────────────────────┼───────────────────────────────────┘
                           │
                           │ HTTP POST
                           │
                           ▼
┌──────────────────────────────────────────────────────────────┐
│ BACKEND: Webhook Endpoint (TO BE IMPLEMENTED)                │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  gRPC Service: Collection Service                            │
│  Method: ProcessIndexerWebhook                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ POST /api/webhooks/indexer                          │    │
│  │                                                       │    │
│  │ Headers:                                             │    │
│  │   X-Webhook-Signature: HMAC-SHA256(secret, payload) │    │
│  │                                                       │    │
│  │ Payload: {                                           │    │
│  │   event: 'collection.created',                      │    │
│  │   chainId: 11155111,                                │    │
│  │   timestamp: 1737540600,                             │    │
│  │   data: {                                            │    │
│  │     collectionAddress: '0xABC...',                  │    │
│  │     creator: '0x123...',                            │    │
│  │     tokenType: 'ERC721',                            │    │
│  │     blockNumber: 12345678,                           │    │
│  │     txHash: '0xDEF...'                              │    │
│  │   }                                                  │    │
│  │ }                                                    │    │
│  │                                                       │    │
│  │ 1. Verify HMAC signature                            │    │
│  │ 2. Find collection by contract_address              │    │
│  │ 3. Update index_status = 'INDEXED'                  │────┼──► PostgreSQL
│  │ 4. Store event in collection_activity table         │    │     collections table
│  │ 5. Publish to event stream (optional)               │    │     index_status = 'INDEXED'
│  │                                                       │    │
│  │ Response: 200 OK                                     │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
└──────────────────────────────────────────────────────────────┘

Final Database State:
  collections table:
    - status: 'DEPLOYED'
    - index_status: 'INDEXED'
    - contract_address: '0xABC...'
    - indexed_at: '2025-01-22T10:31:00Z'
```

---

## Implementation Plan

### Part 1: Frontend SDK Integration (zuno-marketplace-ui)

**Priority**: HIGH
**Effort**: Medium (2-3 days)

#### Files to Create/Modify:

1. **src/shared/providers/ZunoSDKProvider.tsx** (NEW)
   ```typescript
   "use client";
   import { ZunoProvider } from 'zuno-marketplace-sdk/react';

   export function ZunoSDKProvider({ children }) {
     return (
       <ZunoProvider
         apiKey={process.env.NEXT_PUBLIC_ZUNO_API_KEY!}
         network="sepolia"
       >
         {children}
       </ZunoProvider>
     );
   }
   ```

2. **src/modules/mint/create-form/hooks/useCreateCollection.ts** (MODIFY)
   - Add `useCollection` hook from SDK
   - Add Step 4: Deploy contract
   - Add Step 5: Update database
   - Handle wallet signature

3. **src/modules/mint/create-form/components/CollectionProcess.tsx** (MODIFY)
   - Add steps 4 & 5 to UI
   - Show transaction hash
   - Show contract address
   - Link to block explorer

4. **.env.local** (MODIFY)
   ```env
   NEXT_PUBLIC_ZUNO_API_KEY=your-api-key
   NEXT_PUBLIC_ZUNO_NETWORK=sepolia
   ```

**Testing**:
- [ ] SDK provider wraps app
- [ ] Deploy ERC721 works
- [ ] Deploy ERC1155 works
- [ ] Wallet popup appears
- [ ] Contract address saved
- [ ] Status updates to DEPLOYED

---

### Part 2: Indexer Webhook System (zuno-marketplace-indexer)

**Priority**: HIGH
**Effort**: High (3-5 days)

#### Files to Create:

1. **src/infrastructure/webhooks/config.ts**
   ```typescript
   export interface WebhookConfig {
     enabled: boolean;
     url: string;
     secret: string;
     events: string[];
     retryAttempts: number;
     timeout: number;
   }

   export const webhookConfig: WebhookConfig = {
     enabled: process.env.WEBHOOK_ENABLED === 'true',
     url: process.env.WEBHOOK_URL || '',
     secret: process.env.WEBHOOK_SECRET || '',
     events: process.env.WEBHOOK_EVENTS?.split(',') || [],
     retryAttempts: 3,
     timeout: 5000,
   };
   ```

2. **src/infrastructure/webhooks/client.ts**
   ```typescript
   export class WebhookClient {
     async sendWebhook(payload: WebhookPayload): Promise<Result<void>> {
       const signature = this.generateSignature(payload);

       const response = await fetch(webhookConfig.url, {
         method: 'POST',
         headers: {
           'Content-Type': 'application/json',
           'X-Webhook-Signature': signature,
           'X-Webhook-Event': payload.event,
         },
         body: JSON.stringify(payload),
         signal: AbortSignal.timeout(webhookConfig.timeout),
       });

       if (!response.ok) {
         throw new Error(`Webhook failed: ${response.status}`);
       }

       return ok(undefined);
     }

     private generateSignature(payload: WebhookPayload): string {
       const hmac = createHmac('sha256', webhookConfig.secret);
       hmac.update(JSON.stringify(payload));
       return hmac.digest('hex');
     }
   }
   ```

3. **src/infrastructure/webhooks/middleware.ts**
   ```typescript
   export function withWebhookTrigger(handler: EventHandler): EventHandler {
     return async ({ event, context }) => {
       // Execute original handler first
       const result = await handler({ event, context });

       // After successful event storage, trigger webhook
       if (webhookConfig.enabled && shouldTriggerWebhook(event.name)) {
         await webhookClient.sendWebhook({
           event: mapEventName(event.name),
           chainId: context.network.chainId,
           timestamp: event.block.timestamp,
           data: event.args,
         });
       }

       return result;
     };
   }
   ```

#### Files to Modify:

4. **src/infrastructure/monitoring/handler-wrapper.ts**
   - Wrap handlers with webhook middleware
   - Add webhook error handling

5. **ponder.config.ts**
   - Add dynamic contract registration for mints

**.env.local**:
```env
WEBHOOK_ENABLED=true
WEBHOOK_URL=http://localhost:50054/webhooks/indexer
WEBHOOK_SECRET=your-webhook-secret
WEBHOOK_EVENTS=collection_created,nft_minted,batch_minted
```

**Testing**:
- [ ] Webhook fires on collection creation
- [ ] HMAC signature validates
- [ ] Retries on failure
- [ ] Logs webhook delivery status

---

### Part 3: Backend Webhook Endpoint (zuno-marketplace-api)

**Priority**: HIGH
**Effort**: Medium (2-3 days)

#### Files to Create:

1. **services/collection-service/internal/server/webhook_handler.go**
   ```go
   func (s *CollectionServer) ProcessIndexerWebhook(
       ctx context.Context,
       req *pb.IndexerWebhookRequest,
   ) (*pb.IndexerWebhookResponse, error) {
       // 1. Verify HMAC signature
       if !s.verifyWebhookSignature(req) {
           return nil, status.Error(codes.Unauthenticated, "invalid signature")
       }

       // 2. Handle event based on type
       switch req.Event {
       case "collection.created":
           return s.handleCollectionCreated(ctx, req.Data)
       case "collection.minted":
           return s.handleCollectionMinted(ctx, req.Data)
       default:
           return nil, status.Error(codes.InvalidArgument, "unknown event")
       }
   }
   ```

2. **services/graphql-gateway/internal/handlers/webhook_handler.go**
   ```go
   func (h *WebhookHandler) HandleIndexerWebhook(c *gin.Context) {
       var payload IndexerWebhookPayload
       if err := c.BindJSON(&payload); err != nil {
           c.JSON(400, gin.H{"error": "invalid payload"})
           return
       }

       // Verify signature
       signature := c.GetHeader("X-Webhook-Signature")
       if !h.verifySignature(payload, signature) {
           c.JSON(401, gin.H{"error": "invalid signature"})
           return
       }

       // Forward to Collection Service
       _, err := h.collectionClient.ProcessIndexerWebhook(c, &pb.IndexerWebhookRequest{
           Event:     payload.Event,
           ChainId:   payload.ChainId,
           Timestamp: payload.Timestamp,
           Data:      payload.Data,
       })

       if err != nil {
           c.JSON(500, gin.H{"error": err.Error()})
           return
       }

       c.JSON(200, gin.H{"success": true})
   }
   ```

3. **Add route in gateway main.go**
   ```go
   router.POST("/api/webhooks/indexer", webhookHandler.HandleIndexerWebhook)
   ```

#### Database Migration:

**File**: `db/migrations/000004_add_webhook_tracking.up.sql`
```sql
-- Add index_status to collections
ALTER TABLE collections
ADD COLUMN IF NOT EXISTS index_status VARCHAR(20) DEFAULT 'NOT_INDEXED',
ADD COLUMN IF NOT EXISTS indexed_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_collections_index_status
ON collections(index_status);

-- Track webhook deliveries (optional)
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    collection_id UUID REFERENCES collections(id),
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL, -- 'delivered', 'failed', 'pending'
    attempts INT DEFAULT 1,
    last_attempt_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_webhook_deliveries_collection
ON webhook_deliveries(collection_id);
```

**.env**:
```env
WEBHOOK_SECRET=your-webhook-secret
```

**Testing**:
- [ ] Endpoint accepts POST requests
- [ ] Signature verification works
- [ ] Updates index_status correctly
- [ ] Logs webhook events

---

## Event Flow Summary

### Create Collection Event Sequence

```
Time | Component | Action | Database State
-----|-----------|--------|----------------
T0   | Frontend  | User submits form | -
T1   | Frontend  | Upload images | -
T2   | Backend   | Create collection | status=PENDING
T3   | Backend   | Add allowlist | allowlist entries
T4   | Frontend  | Deploy contract (user signs) | -
T5   | Blockchain| Transaction confirmed | -
T6   | Frontend  | Update collection | status=DEPLOYED, contract_address
T7   | Indexer   | Detect event | event table
T8   | Indexer   | Send webhook | -
T9   | Backend   | Process webhook | index_status=INDEXED
```

### Mint Event Sequence (Future)

```
Time | Component | Action | Database State
-----|-----------|--------|----------------
T0   | Frontend  | User mints NFT | -
T1   | Contract  | Mint() called | -
T2   | Blockchain| Minted event emitted | -
T3   | Indexer   | Detect mint event | event table
T4   | Indexer   | Send webhook | -
T5   | Backend   | Update stats | total_items++
```

---

## Environment Variables Summary

### Frontend (.env.local)
```env
NEXT_PUBLIC_BACKEND_URL=http://localhost:8081
NEXT_PUBLIC_ZUNO_API_KEY=your-zuno-api-key
NEXT_PUBLIC_ZUNO_NETWORK=sepolia
```

### Backend (.env)
```env
WEBHOOK_SECRET=your-webhook-secret
```

### Indexer (.env.local)
```env
WEBHOOK_ENABLED=true
WEBHOOK_URL=http://localhost:8081/api/webhooks/indexer
WEBHOOK_SECRET=your-webhook-secret
WEBHOOK_EVENTS=collection_created,nft_minted,batch_minted
PONDER_RPC_URL_11155111=https://sepolia.infura.io/v3/YOUR_KEY
```

---

## Testing Strategy

### Integration Test Flow

1. **Start all services**
   ```bash
   # Backend
   cd zuno-marketplace-api && make dev-up

   # Frontend
   cd zuno-marketplace-ui && pnpm dev

   # Indexer
   cd zuno-marketplace-indexer && pnpm dev
   ```

2. **Create collection end-to-end**
   - Fill form → Upload → Create → Deploy → Verify indexing

3. **Verify each step**
   - Check database after Step 2 (status=PENDING)
   - Check blockchain after Step 4 (contract exists)
   - Check database after Step 5 (status=DEPLOYED)
   - Check indexer event table (event stored)
   - Check database after webhook (index_status=INDEXED)

---

## Success Criteria

- [ ] User can create collection from frontend
- [ ] Images upload successfully
- [ ] Collection saved to database with PENDING status
- [ ] Allowlist saved to database
- [ ] Smart contract deploys to blockchain
- [ ] Contract address saved to database
- [ ] Status updates to DEPLOYED
- [ ] Indexer detects creation event
- [ ] Webhook fires to backend
- [ ] Backend updates index_status to INDEXED
- [ ] All steps show in progress dialog
- [ ] Error handling works at each step

---

**Total Implementation Effort**: 7-11 days
**Priority**: HIGH - Core marketplace feature
**Dependencies**:
- zuno-marketplace-sdk (ready)
- Smart contracts (deployed)
- Backend services (ready)
