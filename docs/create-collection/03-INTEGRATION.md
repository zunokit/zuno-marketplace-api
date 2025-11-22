# INTEGRATION CHECKLIST & WORKFLOWS

This document outlines the integration points between all services and provides step-by-step workflows for end-to-end testing.

---

## 🔗 Service Integration Matrix

### Backend ↔ Frontend

| Integration Point | Backend Provides | Frontend Consumes | Status |
|-------------------|------------------|-------------------|--------|
| **GraphQL Schema** | Collection queries/mutations | Apollo Client hooks | ⏳ Pending |
| **Authentication** | JWT tokens via auth service | Bearer token in headers | ✅ Ready |
| **WebSocket** | Real-time updates (optional) | Apollo subscriptions | 🔮 Future |
| **Error Handling** | Standard error codes | User-friendly messages | ⏳ Pending |

**Setup Checklist**:
- [ ] Backend GraphQL endpoint running on `http://localhost:8081/graphql`
- [ ] Frontend Apollo Client configured with correct URL
- [ ] CORS enabled for `http://localhost:3000` origin
- [ ] Auth middleware properly validates JWT tokens
- [ ] Error responses follow standard format

---

### Frontend ↔ Backend Upload Proxy ↔ Metadata Service

| Integration Point | Frontend Sends | Backend Proxy Forwards | Metadata Service Returns | Status |
|-------------------|----------------|------------------------|--------------------------|--------|
| **Media Upload** | FormData + JWT | FormData + API key | `{ url, ipfsHash, ipfsUrl }` | ⏳ Phase 4 |
| **Create Metadata** | JSON + JWT | JSON + API key | `{ id, isPinned, ipfsUrl }` | ⏳ Phase 4 |
| **Poll Pinning** | GET + JWT | GET + API key | Updated pinning status | ⏳ Phase 4 |

**Setup Checklist**:
- [ ] Backend upload proxy running at `http://localhost:8081/api/upload/*`
- [ ] Metadata service running on `http://localhost:3001`
- [ ] Backend API key configured in `.env` (NOT frontend!)
- [ ] JWT authentication working
- [ ] CORS enabled for frontend origin
- [ ] Rate limits configured appropriately
- [ ] Pinata JWT configured on metadata service
- [ ] ImageKit credentials configured

**Environment Variables**:
```env
# Frontend .env.local
NEXT_PUBLIC_BACKEND_URL=http://localhost:8081  # Backend GraphQL + Upload Proxy

# Backend .env (API key kept secret!)
METADATA_SERVICE_URL=http://localhost:3001/api
METADATA_SERVICE_API_KEY=your-backend-api-key-secret
```

**Security Note**:
- ✅ API key KHÔNG BAO GIỜ expose ở frontend
- ✅ Frontend gọi `/api/upload/*` với JWT token
- ✅ Backend verify JWT → forward với API key

---

### Frontend ↔ Smart Contracts (via SDK)

| Integration Point | Frontend Action | SDK Function | Blockchain Result |
|-------------------|-----------------|--------------|-------------------|
| **Deploy ERC-721** | Call SDK hook | `createERC721Collection()` | Contract deployed, address returned |
| **Deploy ERC-1155** | Call SDK hook | `createERC1155Collection()` | Contract deployed, address returned |
| **Mint NFT** | Call SDK hook | `mintERC721()` | Token minted, tokenId returned |

**Setup Checklist**:
- [ ] Zuno SDK installed: `pnpm add zuno-marketplace-sdk`
- [ ] `ZunoProvider` wrapping app in `AppWrapper.tsx`
- [ ] SDK configured with correct network (Sepolia/Mainnet)
- [ ] Wagmi provider configured with chains
- [ ] Wallet connection working (MetaMask)
- [ ] Transaction signing tested
- [ ] Gas estimation working
- [ ] Block explorer links functional

**Environment Variables**:
```env
# Frontend .env.local
NEXT_PUBLIC_ZUNO_API_KEY=your-zuno-api-key
NEXT_PUBLIC_ZUNO_ABI_URL=https://abis.zuno.com/api
```

**Code Example**:
```typescript
import { useCollection } from 'zuno-marketplace-sdk/react';
import { parseEther } from 'viem';

const { createERC721 } = useCollection();

const { mutateAsync: deployCollection } = createERC721;

const result = await deployCollection({
  name: 'My Collection',
  symbol: 'MYC',
  baseUri: 'https://gateway.pinata.cloud/ipfs/Qm.../,
  maxSupply: 1000,
  owner: userAddress,
  royaltyFee: 500, // 5%
  mintPrice: parseEther('0.1'),
  allowlistMintPrice: parseEther('0.05'),
  mintStartTime: Math.floor(Date.now() / 1000),
  allowlistStageDuration: 3600 * 24, // 24 hours
});

// result = { address: '0x123...', tx: TransactionReceipt }
```

---

### Backend ↔ Metadata Service

| Integration Point | Backend Action | Metadata Service Endpoint | Status |
|-------------------|----------------|---------------------------|--------|
| **Validate Metadata** | GET /metadata/:id | Returns metadata record | ⏳ Pending |
| **Verify IPFS** | Check ipfsHash exists | Metadata record has hash | ⏳ Pending |

**Setup Checklist**:
- [ ] Backend HTTP client configured
- [ ] API key stored in backend `.env`
- [ ] Retry logic for transient errors
- [ ] Timeout handling (30s max)
- [ ] Response validation

**Environment Variables**:
```env
# Backend .env
METADATA_SERVICE_URL=http://localhost:3001/api
METADATA_SERVICE_API_KEY=your-api-key
```

---

### Backend ↔ Blockchain Indexer

| Integration Point | Indexer Action | Backend Endpoint | Status |
|-------------------|----------------|------------------|--------|
| **Collection Created** | Webhook POST | `/webhooks/collection-created` | ⏳ Pending |
| **NFT Minted** | Webhook POST | `/webhooks/nft-minted` | ⏳ Pending |
| **NFT Transferred** | Webhook POST | `/webhooks/nft-transferred` | ⏳ Pending |

**Setup Checklist**:
- [ ] Indexer (Ponder) configured with factory contract ABIs
- [ ] Webhook endpoints implemented in backend
- [ ] Webhook secret configured (`INDEXER_WEBHOOK_SECRET`)
- [ ] HMAC signature validation implemented
- [ ] Idempotency handling (prevent duplicate processing)
- [ ] Event payload schema validated

**Webhook Payload Example**:
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

**Security**:
```go
// Validate HMAC signature
func validateWebhookSignature(payload []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

---

## 🔄 End-to-End Workflows

### Workflow 1: Create ERC-721 Collection (Happy Path)

**Prerequisites**:
- User authenticated (wallet connected + SIWE signed in)
- All services running (Backend, Metadata, Indexer)
- Testnet (Sepolia) has test ETH for gas

**Steps**:

1. **User Opens Form**
   ```
   Navigate to: /create-collection
   Form loads with empty fields
   Wallet status: Connected (0x123...)
   ```

2. **Fill Basic Info**
   ```
   Name: "Genesis Apes"
   Symbol: "GAPE"
   Description: "A collection of 1000 unique apes"
   Token Standard: ERC-721 (selected)
   ```

3. **Upload Images via Backend Proxy**
   ```
   Logo: genesis-logo.png (500x500)
   Banner: genesis-banner.png (1400x400)

   → Frontend calls: POST /api/upload/media (Backend proxy)
      Authorization: Bearer <JWT_TOKEN>
      Content-Type: multipart/form-data

   → Backend verifies JWT ✅
   → Backend forwards: POST http://localhost:3001/api/media
      x-api-key: <BACKEND_API_KEY> 🔒

   ← Returns: { url: "https://ik.imagekit.io/zuno/genesis-logo.png" }

   [Same for banner image]
   ← Returns: { url: "https://ik.imagekit.io/zuno/genesis-banner.png" }
   ```

4. **Configure Minting**
   ```
   Max Supply: 1000
   Mint Start: 2025-12-01 00:00:00
   Allowlist Price: 0.05 ETH
   Allowlist Duration: 24 hours
   Public Price: 0.1 ETH
   ```

5. **Set Royalty**
   ```
   Royalty: 5%
   Recipient: 0x123... (creator address)
   ```

6. **Review & Create Metadata via Backend Proxy**
   ```
   → Frontend calls: POST /api/upload/metadata (Backend proxy)
      Authorization: Bearer <JWT_TOKEN>
      Content-Type: application/json
      {
        name: "Genesis Apes",
        image: "https://ik.imagekit.io/zuno/genesis-logo.png",
        bannerImage: "https://ik.imagekit.io/zuno/genesis-banner.png",
        symbol: "GAPE",
        description: "A collection of 1000 unique apes"
      }

   → Backend verifies JWT ✅
   → Backend forwards: POST http://localhost:3001/api/metadata
      x-api-key: <BACKEND_API_KEY> 🔒
      x-api-version: v1
      { ...same payload }

   ← Returns: {
       success: true,
       data: {
         id: "metadata_abc123",
         isPinned: false,  // Not pinned yet!
         ipfsHash: null,
         ipfsUrl: null
       },
       meta: {
         requestId: "uuid",
         timestamp: "2025-11-20T12:34:56Z"
       }
     }
   ```

7. **Poll IPFS Pinning via Backend Proxy** ⏱️ **ASYNC** (~10-30 seconds)
   ```
   ⚠️ IMPORTANT: IPFS pinning happens in background via cron job!

   → Frontend polls: GET /api/upload/metadata/metadata_abc123 (Backend proxy)
      Authorization: Bearer <JWT_TOKEN>
   [Check every 2s...]

   → Backend forwards: GET http://localhost:3001/api/metadata/metadata_abc123
      x-api-key: <BACKEND_API_KEY> 🔒

   ← Returns (initially): {
       isPinned: false,
       ipfsHash: null
     }

   [Wait 2s, retry...]
   [Wait 2s, retry...]
   [Background cron processes pinning...]

   ← Returns (after ~10-30s): {
       id: "metadata_abc123",
       isPinned: true,  // ✅ Now pinned!
       ipfsHash: "QmXyz...",
       ipfsUrl: "https://gateway.pinata.cloud/ipfs/QmXyz...",
       pinnedAt: "2025-11-20T12:35:00Z"
     }

   BaseURI: "https://gateway.pinata.cloud/ipfs/QmXyz.../"
   ```

   **Note**: Show loading indicator to user during polling!

8. **Create Collection in Backend**
   ```
   → Frontend calls: createCollectionMutation({
       input: {
         name: "Genesis Apes",
         symbol: "GAPE",
         tokenStandard: ERC721,
         metadataUri: "https://gateway.pinata.cloud/ipfs/QmXyz.../",
         deployerAddress: "0x123...",
         maxSupply: 1000,
         mintPriceAllowlist: "50000000000000000", // 0.05 ETH in Wei
         mintPricePublic: "100000000000000000", // 0.1 ETH in Wei
         royaltyFeeBps: 500,
         imageUrl: "https://ik.imagekit.io/zuno/genesis-logo.png",
         bannerImageUrl: "https://ik.imagekit.io/zuno/genesis-banner.png"
       }
     })

   ← Backend creates record with status=PENDING
   ← Returns: {
       collection: {
         id: "col_xyz789",
         status: PENDING
       }
     }
   ```

9. **Deploy Smart Contract**
   ```
   → Frontend calls: sdk.collection.createERC721Collection({
       name: "Genesis Apes",
       symbol: "GAPE",
       baseUri: "https://gateway.pinata.cloud/ipfs/QmXyz.../",
       maxSupply: 1000,
       owner: "0x123...",
       royaltyFee: 500,
       mintPrice: parseEther('0.1'),
       allowlistMintPrice: parseEther('0.05'),
       mintStartTime: 1733011200,
       allowlistStageDuration: 86400
     })

   → MetaMask popup appears
   → User reviews transaction:
       - Gas: ~0.005 ETH
       - Network: Sepolia
   → User clicks "Confirm"

   → Transaction sent to blockchain
   → Wait for confirmation (1-2 blocks)

   ← Returns: {
       address: "0x456...",
       tx: { hash: "0xabc...", chainId: 11155111 }
     }
   ```

10. **Update Backend with Contract Address**
    ```
    → Frontend calls: updateCollectionMutation({
        id: "col_xyz789",
        input: {
          contractAddress: "0x456...",
          chainId: "eip155:11155111",
          deployedAt: "2025-11-20T10:35:00Z"
        }
      })

    ← Backend updates collection status=DEPLOYED
    ← Returns: { collection: { status: DEPLOYED, contractAddress: "0x456..." } }
    ```

11. **Indexer Catches Event**
    ```
    → Blockchain emits: CollectionCreated(0x456..., 0x123...)
    → Ponder indexer catches event
    → Webhook POST to backend: /webhooks/collection-created
    ← Backend verifies HMAC signature
    ← Backend confirms deployment (already updated in step 10)
    ```

12. **Success**
    ```
    → Frontend shows success message
    → User redirected to: /collections/col_xyz789
    → Collection page displays:
       - Name, symbol, images
       - Contract address (link to Etherscan)
       - Status: DEPLOYED
       - Stats: 0 items, 0 owners
    ```

**Total Time**: ~2-3 minutes (including IPFS pinning + blockchain confirmation)

---

### Workflow 2: Handle Transaction Rejection

**Scenario**: User rejects MetaMask signature

**Steps 1-9**: Same as Workflow 1

**Step 10**: User Rejects
```
→ MetaMask popup appears
→ User clicks "Reject"
← SDK throws error: "User rejected transaction"

→ Frontend catches error
→ Shows error message: "Transaction cancelled. Please try again."
→ Collection remains in PENDING status
→ "Retry" button available
```

**Recovery**:
```
User clicks "Retry"
→ Frontend re-attempts contract deployment (step 9)
→ MetaMask popup appears again
→ User can try again
```

---

### Workflow 3: Handle Insufficient Gas

**Scenario**: User has insufficient ETH for gas

**Step 10**: Insufficient Funds
```
→ MetaMask shows error: "Insufficient funds for gas"
→ SDK throws error: "Insufficient funds"

→ Frontend catches error
→ Shows error message: "Insufficient ETH for gas. Please add funds to your wallet."
→ Provides link to faucet (Sepolia testnet)
```

**Recovery**:
```
User adds ETH to wallet
→ Clicks "Retry" button
→ Frontend re-attempts deployment
```

---

### Workflow 4: Add Allowlist After Deployment

**Prerequisites**: Collection already deployed

**Steps**:

1. **Navigate to Collection**
   ```
   Go to: /collections/col_xyz789
   Click: "Manage Allowlist" button
   ```

2. **Open Allowlist Modal**
   ```
   Modal shows:
   - Current allowlist (empty or existing addresses)
   - "Add Addresses" section
   - CSV upload option
   ```

3. **Add Addresses**
   ```
   Input addresses (one per line):
   0xaaa111...
   0xbbb222...
   0xccc333...

   Or upload CSV:
   address
   0xaaa111...
   0xbbb222...
   ```

4. **Submit to Backend**
   ```
   → Frontend calls: addToAllowlistMutation({
       input: {
         collectionId: "col_xyz789",
         walletAddresses: ["0xaaa111...", "0xbbb222...", "0xccc333..."],
         maxMintAmount: 2
       }
     })

   ← Backend validates addresses (checksum)
   ← Backend inserts into collection_allowlist table
   ← Returns: { success: true }
   ```

5. **Sync to Smart Contract** (Optional)
   ```
   NOTE: Depending on contract implementation, allowlist may be:
   - Off-chain only (backend verifies before allowing mint)
   - On-chain (requires contract call to addToAllowlist)

   If on-chain:
   → Frontend calls: sdk.collection.addToAllowlist(collectionAddress, addresses)
   → User signs transaction
   → Contract updated
   ```

6. **Success**
   ```
   → Modal shows success message
   → Allowlist table refreshes with new addresses
   ```

---

## 🧪 Testing Checklist

### Unit Testing

**Backend**:
- [ ] Repository layer with mocked DB
- [ ] Service layer with mocked repos
- [ ] gRPC server with mocked services
- [ ] GraphQL resolvers with mocked gRPC clients
- [ ] Webhook signature validation

**Frontend**:
- [ ] Form validation (Zod schemas)
- [ ] Metadata client with mocked fetch
- [ ] GraphQL hooks with MockedProvider
- [ ] Utility functions (formatters, validators)

### Integration Testing

**Backend**:
- [ ] Database migrations (up + down)
- [ ] gRPC service with real database (testcontainers)
- [ ] GraphQL gateway with real gRPC clients
- [ ] External service clients with mock servers

**Frontend**:
- [ ] Form submission flow with mocked backend
- [ ] Apollo Client with mocked GraphQL
- [ ] SDK integration with mocked blockchain

### E2E Testing (Playwright)

- [ ] Happy path: Create ERC-721 collection (mocked MetaMask)
- [ ] Create ERC-1155 collection
- [ ] Transaction rejection handling
- [ ] Add allowlist after deployment
- [ ] Browse collections page
- [ ] Search collections
- [ ] View collection detail page
- [ ] Edit collection metadata

### Load Testing

- [ ] 100 concurrent collection creations (backend)
- [ ] 1000 GraphQL queries/sec (gateway)
- [ ] Media upload stress test (metadata service)

---

## 🚨 Error Scenarios & Handling

| Scenario | Detection | User Message | Recovery |
|----------|-----------|--------------|----------|
| **IPFS pinning timeout** | Poll exceeds 30 retries | "IPFS pinning is taking longer than expected. Please check back later." | Retry button, email notification when ready |
| **Metadata service down** | HTTP 503 error | "Metadata service is temporarily unavailable. Please try again later." | Retry with exponential backoff |
| **Backend GraphQL error** | 500 response | "An error occurred. Please contact support with ID: {requestId}" | Log error, show request ID for support |
| **Transaction failed** | SDK throws error | "Transaction failed: {reason}. Please try again." | Retry button, check gas + balance |
| **Duplicate collection name** | Backend validation error | "A collection with this name already exists. Please choose a different name." | Allow user to edit name |
| **Invalid wallet address** | Frontend validation | "Invalid wallet address. Please check and try again." | Highlight invalid field |
| **Rate limit exceeded** | 429 response | "Too many requests. Please wait {retryAfter} seconds." | Show countdown timer |

---

## 📊 Monitoring & Observability

### Key Metrics to Track

**Backend**:
- Collection creation success rate
- Average collection creation time
- GraphQL query latency (p50, p95, p99)
- gRPC call latency
- Database query performance
- Webhook processing time

**Frontend**:
- Form abandonment rate
- Form completion time
- Transaction rejection rate
- IPFS pinning success rate
- Media upload success rate
- Page load time (Lighthouse scores)

**Blockchain**:
- Gas costs (average, median)
- Transaction confirmation time
- Failed transaction rate
- Contract deployment success rate

### Alerts to Configure

- Collection creation failure rate > 5%
- IPFS pinning timeout > 10%
- GraphQL p95 latency > 500ms
- Webhook delivery failure rate > 1%
- Database connection pool exhausted
- Metadata service downtime > 1 minute

---

**Last Updated**: 2025-11-20
**Status**: Integration planning
