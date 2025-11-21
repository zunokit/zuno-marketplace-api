# CREATE COLLECTION FEATURE - OVERVIEW

## 🎯 Feature Description

Enable users to create and deploy their own NFT collections (ERC-721 and ERC-1155) on the Zuno Marketplace platform.

---

## 🏗️ Architecture Confirmation

### Backend (zuno-marketplace-api)
- ✅ Chuẩn bị và validate dữ liệu collection
- ✅ Gọi Metadata Service để store metadata/images lên IPFS
- ✅ Lưu collection records vào PostgreSQL database
- ✅ Index blockchain events sau khi collection được deploy
- ✅ Provide GraphQL API cho FE query collections

### Frontend (zuno-marketplace-ui)
- ✅ User điền form (name, symbol, description, images)
- ✅ Upload images → Metadata Service → nhận URLs
- ✅ Tạo metadata → Metadata Service → nhận baseURI (IPFS)
- ✅ **Connect wallet + call smart contract** via Zuno SDK
- ✅ User **manual verify & sign transaction** trong ví
- ✅ Transaction confirmed → notify Backend để index

### Smart Contracts (zuno-marketplace-contracts)
- ✅ Factory contracts đã deploy sẵn
- ✅ SDK wrapper functions ready: `createERC721Collection()`, `createERC1155Collection()`
- ✅ 3 mint stages: INACTIVE → ALLOWLIST → PUBLIC
- ✅ EIP-2981 royalty standard support

### Metadata Service (zuno-marketplace-metadata)
- ✅ RESTful API với API key authentication
- ✅ Upload images → ImageKit CDN
- ✅ Store metadata → IPFS via Pinata
- ✅ Background worker for pinning
- ✅ Return baseURI for collections

### Indexer (zuno-marketplace-indexer)
- ✅ Ponder framework listening to blockchain events
- ✅ Index `CollectionCreated` events
- ✅ Webhook to notify backend API
- ✅ Real-time sync on-chain data

---

## 📂 Documentation Structure

```
docs/create-collection/
├── 00-OVERVIEW.md           # This file - Architecture overview
├── 01-BACKEND-PLAN.md       # Backend implementation phases
├── 02-FRONTEND-PLAN.md      # Frontend implementation phases
├── 03-INTEGRATION.md        # Integration checklist & workflows
└── 04-COMPLETION-CRITERIA.md # Feature completion criteria
```

---

## 🔄 End-to-End Flow

### 1. User Opens Create Collection Form (FE)
- User connects wallet via RainbowKit
- User fills out multi-step form:
  - Basic info: name, symbol, description, token standard
  - Media: logo, banner, featured images
  - Minting: max supply, prices, mint start time, allowlist duration
  - Royalty: percentage, recipient address
  - Allowlist: wallet addresses (optional)

### 2. Upload Media to Metadata Service (FE → Metadata Service)
```
POST /api/media
Authorization: Bearer API_KEY
Content-Type: multipart/form-data

→ Returns: { url: "https://ik.imagekit.io/zuno/logo.png" }
```

### 3. Create Metadata Record (FE → Metadata Service)
```
POST /api/metadata
Authorization: Bearer API_KEY
Content-Type: application/json
{
  "name": "My Collection",
  "image": "https://ik.imagekit.io/zuno/logo.png",
  "symbol": "MYC",
  ...
}

→ Returns: { id: "metadata_123", isPinned: false }
```

### 4. Poll for IPFS Pinning (FE → Metadata Service)
```
GET /api/metadata/metadata_123
Authorization: Bearer API_KEY

→ Poll until: { isPinned: true, ipfsUrl: "https://gateway.pinata.cloud/ipfs/Qm..." }
```

### 5. Create Collection Record in Backend (FE → Backend API)
```graphql
mutation CreateCollection($input: CreateCollectionInput!) {
  createCollection(input: $input) {
    id
    name
    status  # PENDING
  }
}
```

Backend creates record with `status=PENDING`.

### 6. Deploy Smart Contract (FE → Blockchain via SDK)
```typescript
const { address, tx } = await sdk.collection.createERC721Collection({
  name: "My Collection",
  symbol: "MYC",
  baseUri: "https://gateway.pinata.cloud/ipfs/Qm.../",
  maxSupply: 1000,
  // ... other params
});

// User signs transaction in wallet
// Transaction confirmed on blockchain
```

### 7. Update Collection with Contract Address (FE → Backend API)
```graphql
mutation UpdateCollection($id: ID!, $input: UpdateCollectionInput!) {
  updateCollection(id: $id, input: $input) {
    id
    contractAddress  # 0x123...
    chainId          # eip155:11155111
    status           # DEPLOYED
  }
}
```

### 8. Add Allowlist (Optional, FE → Backend API)
```graphql
mutation AddToAllowlist($input: AddToAllowlistInput!) {
  addToAllowlist(input: {
    collectionId: "col_123",
    walletAddresses: ["0xabc...", "0xdef..."]
  })
}
```

### 9. Blockchain Event Indexing (Indexer → Backend API)
```
CollectionCreated event emitted
  ↓
Ponder indexer catches event
  ↓
Webhook POST /webhooks/collection-created
  ↓
Backend verifies and confirms deployment
```

### 10. User Views Collection (FE)
```
/collections/[slug] page shows:
- Collection info (name, description, images)
- Stats (items, owners, floor price, volume)
- Items grid (NFTs in collection)
- Activity feed
```

---

## 🎯 MVP Scope (Minimum Viable Product)

### Must Have
- [x] Create ERC-721 collection
- [x] Upload images to IPFS
- [x] Deploy smart contract on testnet (Sepolia)
- [x] Store collection in database
- [x] View collection details
- [x] Edit collection metadata (description, images)

### Should Have (V1)
- [ ] Create ERC-1155 collection
- [ ] Allowlist management
- [ ] Mint stages (INACTIVE → ALLOWLIST → PUBLIC)
- [ ] Royalty configuration
- [ ] Collection stats (floor, volume)
- [ ] Search & filter collections

### Nice to Have (V2)
- [ ] Batch collection creation
- [ ] Collection templates
- [ ] Analytics dashboard
- [ ] Multi-chain support
- [ ] Collection verification badge

---

## 📊 Success Metrics

### Technical
- Collection creation success rate > 95%
- Contract deployment time < 2 minutes
- IPFS pinning time < 30 seconds
- API response time < 200ms (p95)

### User Experience
- Form completion rate > 80%
- Time to create collection < 5 minutes
- User satisfaction score > 4.5/5

---

## 🔗 Related Repositories

| Repository | Purpose | Role in Feature |
|------------|---------|-----------------|
| **zuno-marketplace-api** | Backend API | Store collections, provide GraphQL API |
| **zuno-marketplace-ui** | Frontend | Create collection UI, wallet integration |
| **zuno-marketplace-sdk** | TypeScript SDK | Smart contract interaction wrapper |
| **zuno-marketplace-contracts** | Solidity contracts | Collection factory & templates |
| **zuno-marketplace-metadata** | Metadata service | IPFS storage, image hosting |
| **zuno-marketplace-indexer** | Event indexer | Sync blockchain events to database |
| **zuno-marketplace-abis** | ABI provider | Serve contract ABIs to SDK |

---

## 🚀 Getting Started

### For Backend Developers
1. Read [Backend Plan](./01-BACKEND-PLAN.md)
2. Start with Phase 1: Database Schema Design
3. Follow the sequential phases

### For Frontend Developers
1. Read [Frontend Plan](./02-FRONTEND-PLAN.md)
2. Wait for Backend Phase 3 (GraphQL Gateway) to complete
3. Start with Phase 1: GraphQL Integration

### For Integration Testing
1. Read [Integration Guide](./03-INTEGRATION.md)
2. Ensure all services are running
3. Follow the end-to-end test scenarios

---

## 📝 Notes

- **Backend takes priority**: Frontend cannot proceed without backend GraphQL schema
- **Testnet first**: All development on Sepolia testnet, migrate to mainnet later
- **Incremental deployment**: Deploy MVP first, then add V1/V2 features
- **Security**: Validate all user inputs, verify contract ownership, prevent SQL injection

---

## 🤝 Team Coordination

### Backend Team
- Implement collection service (gRPC microservice)
- Create GraphQL schema & resolvers
- Integrate with metadata service
- Set up blockchain event indexing

### Frontend Team
- Build create collection form
- Integrate Zuno SDK for contract deployment
- Connect to metadata service for uploads
- Build collection management UI

### DevOps Team
- Deploy metadata service to production
- Configure IPFS (Pinata) credentials
- Set up monitoring & alerts
- Configure environment variables

---

**Last Updated**: 2025-11-20
**Status**: Planning Phase
**Next Steps**: Begin Backend Phase 1 (Database Schema Design)
