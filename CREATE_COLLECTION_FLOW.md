# Create Collection Flow - Zuno Marketplace

## 📋 Tổng Quan Hệ Thống

Hệ thống Zuno Marketplace bao gồm 11 repositories chính, tương tác với nhau để tạo một flow hoàn chỉnh cho việc tạo NFT collection.

### Danh Sách Repositories

| Repository | Vai Trò | Tech Stack | Port/URL |
|------------|---------|------------|----------|
| **zuno-marketplace-ui** | Frontend chính cho end-users | Next.js 15, React 19, TypeScript | :3000 |
| **zuno-marketplace-api** | Backend API (GraphQL + gRPC) | Go, gRPC, PostgreSQL | :8081 (GraphQL) |
| **zuno-marketplace-contracts** | Smart contracts NFT | Solidity, Foundry | Blockchain |
| **zuno-marketplace-sdk** | TypeScript SDK cho Web3 | TypeScript, Wagmi, Ethers | NPM Package |
| **zuno-marketplace-metadata** | Metadata & IPFS service | Next.js 16, Drizzle ORM, Pinata | :3001 |
| **zuno-marketplace-indexer** | Blockchain event indexer | Ponder Framework, PostgreSQL | :42069 |
| **zuno-marketplace-event-stream** | Real-time event display | Next.js, TanStack Query | :3002 |
| **zuno-marketplace-notifications** | Notification service | Next.js, Prisma, BullMQ | :3003 |
| **zuno-marketplace-abis** | ABI registry service | Next.js, Drizzle ORM | :3004 |
| **zuno-marketplace-admin** | Admin dashboard | Next.js, Drizzle Studio | :3005 |
| **zuno-marketplace-mini** | Testing/demo app | Next.js, Redux | :3006 |

---

## 🔄 Flow Tổng Quan - 8 Bước Chính

```
┌─────────────────────────────────────────────────────────────────────┐
│                        CREATE COLLECTION FLOW                        │
└─────────────────────────────────────────────────────────────────────┘

[1] User Input (UI)
      ↓
[2] Upload Images (Metadata Service)
      ↓
[3] Pin to IPFS (Pinata Background Job)
      ↓
[4] Deploy Smart Contract (Blockchain)
      ↓
[5] Index Events (Indexer)
      ↓
[6] Sync to Database (API Service) ⚠️ NOT IMPLEMENTED
      ↓
[7] Real-time Update (Event Stream)
      ↓
[8] Notify Users (Notifications)
```

---

## 📍 BƯỚC 1: User Input - Frontend UI

### Repository: `zuno-marketplace-ui`

**Entry Points:**
- Simple Form: `/create/collection`
- Advanced Form: `/mint/create`

**Component Files:**
- `src/modules/create/components/CreateCollectionForm.tsx` - Simple 4-tab form
- `src/modules/mint/create-form/components/CollectionForm.tsx` - Advanced form with Zod validation

**Form Data Thu Thập:**

| Nhóm | Field | Required | Validation |
|------|-------|----------|------------|
| **Basic Info** | name | ✅ | 1-100 chars |
| | symbol | ✅ | 3-10 chars uppercase |
| | description | ❌ | Max 2000 chars |
| | category | ❌ | Dropdown (Art, Gaming, Music, etc.) |
| **Token Type** | artworkMode | ✅ | "ERC721" or "ERC1155" |
| | chain | ✅ | CAIP-2 format (eip155:1, eip155:137) |
| **Media** | collectionImage | ✅ | File (800x800px recommended) |
| | bannerImage | ❌ | File (1400x400px recommended) |
| **Mint Config** | mintStartAt | ✅ | ISO 8601 datetime |
| | mintPrice | ❌ | Regex: `^(?:0\|[1-9]\d*)(?:\.\d{1,18})?$` |
| | maxSupply | ❌ | Positive integer or null (unlimited) |
| | mintLimitPerWallet | ❌ | Positive integer |
| **Royalty** | royaltyPercentage | ❌ | 0-10% (dropdown) |
| | royaltyAddress | ❌ | EVM address |
| **Stages** | presale.duration | ❌ | `{ days: number, hours: number }` |
| | presale.allowlistAddresses | ❌ | Array 1-5000 EVM addresses |
| | public.duration | ❌ | Nullable |
| **Metadata** | metadataBaseUrl | ✅* | HTTPS or IPFS URL (required for ERC721) |
| | sameArtworkImage | ✅* | File (required for ERC1155) |

**Validation Schema:**
- Location: `src/shared/types/mint.ts`
- Schema: `MintTerminalCreateFormSchema` (Zod)
- Validates: address format, price format, stage logic, allowlist uniqueness

**Current Status:**
- ✅ UI hoàn chỉnh với 2 forms
- ✅ Client-side validation
- ❌ Form submit chỉ `console.log()`, chưa call API
- ❌ Chưa tích hợp SDK
- ❌ Chưa có transaction signing flow

**Output từ bước này:**
- Form data object ready cho processing
- User đã authenticated (JWT token có sẵn)

---

## 📍 BƯỚC 2: Upload Images - Metadata Service

### Repository: `zuno-marketplace-metadata`

**API Endpoints:**
- `POST /api/media` - Upload single file
- `POST /api/media/batch` - Upload multiple files

**Processing Flow:**

### 2.1. Upload to CDN (ImageKit)

**Service:** `src/infrastructure/services/imagekit.service.ts`

**Flow:**
1. Receive file từ frontend (multipart/form-data)
2. Validate file:
   - Type: IMAGE, VIDEO, GIF, MODEL_3D
   - Size limits:
     - Images: 50MB max
     - Videos: 100MB max
     - GIFs: 25MB max
     - 3D Models: 100MB max
3. Upload to ImageKit CDN
4. Auto-generate thumbnail (for images/videos)
5. Create optimized URLs with transformations

**Output:**
- CDN URL: `https://ik.imagekit.io/zuno/collections/abc123.png`
- Thumbnail URL: `https://ik.imagekit.io/zuno/collections/tr:w-350/abc123.png`
- File metadata (width, height, size, mimeType)

### 2.2. Save to Database

**Schema:** `src/infrastructure/database/drizzle/schema/media.schema.ts`

**Table:** `media`

**Fields Stored:**
- `id` - UUID
- `user_id` - Owner (from JWT token)
- `file_name` - Original filename
- `file_size` - Bytes
- `mime_type` - e.g., "image/png"
- `media_type` - IMAGE, VIDEO, GIF, MODEL_3D
- `url` - ImageKit CDN URL
- `thumbnail_url` - Generated thumbnail
- `width`, `height` - Dimensions (for images)
- `is_pinned` - Initially `false`
- `ipfs_hash`, `ipfs_url` - Initially `null`
- `created_at`, `updated_at`

### 2.3. Queue IPFS Pinning Job

**Queue:** BullMQ with Redis (Upstash)

**Job Details:**
- Queue name: `media-ipfs-pin`
- Job data: `{ mediaId: "uuid", url: "https://..." }`
- Concurrency: 10 jobs simultaneously
- Retry: Automatic on failure

**Why Background Job:**
- IPFS upload slow (5-30 seconds)
- Non-blocking API response
- User gets immediate feedback

**Output từ bước này:**
- Media record created with CDN URL
- Background job queued for IPFS
- Return to frontend: `{ id, url, thumbnailUrl }`

---

## 📍 BƯỚC 3: Pin to IPFS - Background Processing

### Repository: `zuno-marketplace-metadata`

**Worker:** `src/infrastructure/queue/workers/media-ipfs-pin.worker.ts`

**Processing Flow:**

### 3.1. Worker Picks Up Job

**When:**
- Runs continuously: `pnpm workers`
- Processes jobs from queue as they arrive

**What it does:**
1. Fetch media record từ database by `mediaId`
2. Download file từ ImageKit CDN URL
3. Upload to Pinata IPFS

### 3.2. Pinata Upload

**Service:** `src/infrastructure/services/pinata/pinata.service.ts`

**Flow:**
1. Call Pinata API:
   - Endpoint: `https://api.pinata.cloud/pinning/pinFileToIPFS`
   - Auth: JWT token
   - Body: File stream
   - Metadata: `{ name, keyvalues: { mediaId, userId } }`
   - Group: `PINATA_GROUPS.MEDIA`

2. Receive response:
   - `IpfsHash`: "Qm...abc123"
   - `PinSize`: Size in bytes
   - `Timestamp`: Pin time

3. Construct IPFS URLs:
   - Hash format: `Qm...abc123`
   - IPFS URI: `ipfs://Qm...abc123`
   - Gateway URL: `https://gateway.pinata.cloud/ipfs/Qm...abc123`

### 3.3. Update Database

**SQL Operation:**
```
UPDATE media
SET ipfs_hash = 'Qm...abc123',
    ipfs_url = 'ipfs://Qm...abc123',
    is_pinned = true,
    pinned_at = NOW()
WHERE id = 'mediaId'
```

**Cache Invalidation:**
- Clear Redis cache for this media item
- Clear user's media list cache

### 3.4. Create Collection Metadata JSON

**API Endpoint:** `POST /api/metadata`

**Request Body:**
```
{
  name: "My Cool NFT Collection",
  description: "...",
  symbol: "COOL",
  image: "ipfs://Qm...abc123",        // từ bước 3.3
  bannerImage: "ipfs://Qm...def456",  // từ bước 3.3
  externalUrl: "https://mycollection.com",
  sellerFeeBasisPoints: 750,  // 7.5% royalty
  feeRecipient: "0x...",
  attributes: [
    { traitType: "Category", value: "Art" }
  ]
}
```

**Validation:** `src/shared/lib/validation/metadata.schemas.ts`
- Zod schema validates OpenSea compatibility
- Checks required fields (name, image)
- Validates attribute formats
- Ensures no duplicate trait types

**Database Storage:**
- Table: `metadata`
- Fields: name, symbol, description, image, attributes (JSONB), etc.
- Ownership: Linked to `user_id` (IDOR protection)

### 3.5. Pin Metadata JSON to IPFS

**Worker:** `src/infrastructure/queue/workers/metadata-ipfs-pin.worker.ts`

**Flow:**
1. Queue job: `metadata-ipfs-pin`
2. Worker converts metadata to OpenSea-compatible JSON:
   ```
   {
     "name": "My Cool NFT Collection",
     "description": "...",
     "image": "ipfs://Qm...abc123",
     "external_url": "https://...",
     "seller_fee_basis_points": 750,
     "fee_recipient": "0x...",
     "attributes": [...]
   }
   ```
3. Upload JSON to Pinata:
   - Group: `PINATA_GROUPS.METADATA`
   - Returns: `IpfsHash` for metadata JSON
4. Update database:
   ```
   UPDATE metadata
   SET ipfs_hash = 'Qm...metadata789',
       ipfs_url = 'ipfs://Qm...metadata789',
       is_pinned = true
   ```

**Output từ bước này:**
- ✅ Collection logo on IPFS: `ipfs://Qm...abc123`
- ✅ Collection banner on IPFS: `ipfs://Qm...def456`
- ✅ Collection metadata JSON on IPFS: `ipfs://Qm...metadata789`
- ✅ All data có CDN fallback URLs (ImageKit)

---

## 📍 BƯỚC 4: Deploy Smart Contract - Blockchain

### Repositories:
- `zuno-marketplace-contracts` (Smart contracts)
- `zuno-marketplace-sdk` (TypeScript SDK)
- `zuno-marketplace-ui` (Frontend integration)

### 4.1. Frontend Prepares Transaction

**Hook Used:** `useCollection()` from `zuno-marketplace-sdk/react`

**Function Called:** `createERC721Collection()` or `createERC1155Collection()`

**Parameters Mapping:**

| Frontend Field | Smart Contract Param | Conversion |
|----------------|---------------------|------------|
| name | `params.name` | String |
| symbol | `params.symbol` | String uppercase |
| userAddress | `params.owner` | Address from wallet |
| description | `params.description` | String |
| maxSupply | `params.maxSupply` | BigInt |
| mintLimitPerWallet | `params.mintLimitPerWallet` | BigInt |
| mintStartAt (Date) | `params.mintStartTime` | Unix timestamp (seconds) |
| mintPrice | `params.publicMintPrice` | Wei (ethers.parseEther) |
| presale.price | `params.allowlistMintPrice` | Wei |
| presale.duration | `params.allowlistStageDuration` | Seconds (days*86400 + hours*3600) |
| royaltyPercentage | `params.royaltyFee` | Basis points (7.5% → 750) |
| royaltyAddress | `params.royaltyFee` + separate Fee contract | Address |
| metadataBaseUrl | `params.tokenURI` | IPFS or HTTPS URL with trailing "/" |

### 4.2. SDK Gets Factory Address

**Contract Hierarchy:**
```
UserHub (Router)
  ├─ getFactoryFor("ERC721") → ERC721CollectionFactory address
  └─ getFactoryFor("ERC1155") → ERC1155CollectionFactory address

CollectionRegistry
  └─ Maps "ERC721" → Factory address (set by admin)

AdminHub
  └─ registerCollectionFactory() (admin only)
```

**SDK Flow:**
1. Connect to UserHub contract
2. Call `getFactoryFor("ERC721")`
3. Receive factory address: `0xFactory123...`
4. Get ABI from `zuno-marketplace-abis` service or cache
5. Create contract instance: `new ethers.Contract(factoryAddress, abi, signer)`

**ABI Caching:**
- Location: `src/core/ContractRegistry.ts`
- Uses TanStack Query for cache (5 min stale time)
- Fetches from: `GET https://abis.zuno.com/api/abis/ERC721CollectionFactory`
- Falls back to local ABIs if service unavailable

### 4.3. Smart Contract Execution

**Contract:** `src/core/factory/ERC721CollectionFactory.sol`

**Function:** `createERC721Collection(CollectionParams memory params)`

**Internal Flow:**

#### Step 1: Parameter Validation
- File: `src/common/BaseCollectionFactory.sol`
- Function: `_validateCollectionParams(params)`
- Checks:
  - ✅ `params.owner != address(0)`
  - ✅ `bytes(params.name).length > 0`
  - ✅ `bytes(params.symbol).length > 0`
  - ✅ `params.maxSupply > 0`
  - ✅ `params.royaltyFee <= 1000` (max 10%)
- Reverts with custom errors if fails

#### Step 2: Deploy Minimal Proxy
- File: `src/core/factory/ERC721CollectionFactory.sol`
- Function: `_deployCollectionProxy()`
- Uses: OpenZeppelin's `Clones.clone(implementationAddress)`
- Gas cost: ~45,000 gas (vs ~3,000,000 for full deployment)
- Returns: New collection address

**Implementation Address:**
- Set in constructor
- Points to: `ERC721CollectionImplementation.sol`
- Deployed once, reused for all collections (proxy pattern)

#### Step 3: Initialize Collection
- Function: `initialize(CollectionParams memory params)`
- Location: `src/core/collection/ERC721Collection.sol`
- Protected by: `initializer` modifier (can only call once)

**What gets initialized:**
1. **ERC721 Base:**
   - Name: `params.name`
   - Symbol: `params.symbol`
   - Inherits from OpenZeppelin ERC721

2. **Ownership:**
   - Transfer ownership to `params.owner`
   - Uses OpenZeppelin Ownable

3. **Mint Configuration:**
   - `s_maxSupply = params.maxSupply`
   - `s_mintLimitPerWallet = params.mintLimitPerWallet`
   - `s_mintStartTime = params.mintStartTime`
   - `s_allowlistMintPrice = params.allowlistMintPrice`
   - `s_publicMintPrice = params.publicMintPrice`
   - `s_allowlistStageDuration = params.allowlistStageDuration`
   - `s_totalMinted = 0`

4. **Metadata:**
   - `s_baseTokenURI = params.tokenURI`
   - Used in `tokenURI(tokenId)` function

5. **Royalty (EIP-2981):**
   - Creates separate `Fee.sol` contract
   - Fee contract stores: `royaltyFee`, `royaltyReceiver`
   - Collection implements `royaltyInfo(tokenId, salePrice)`
   - Returns: `(receiver, royaltyAmount)`

6. **Mint Stage:**
   - Initial stage: `MintStage.INACTIVE`
   - Auto-transitions based on timestamps:
     - Before `mintStartTime`: INACTIVE
     - `mintStartTime` to `mintStartTime + allowlistStageDuration`: ALLOWLIST
     - After allowlist ends: PUBLIC

#### Step 4: Register Collection
- Function: `_registerCollection(collectionAddress)`
- Storage: `mapping(address => bool) s_collections`
- Marks: `s_collections[newCollection] = true`
- Increment: `s_totalCollections++`
- Purpose: Verify collection created by factory

#### Step 5: Emit Event
- Event: `ERC721CollectionCreated(address indexed collectionAddress, address indexed creator)`
- Event location: `src/events/CollectionEvents.sol`
- Indexed params: Both `collectionAddress` and `creator` (for efficient filtering)

**Event Data:**
```
{
  collectionAddress: "0xNewCollection123...",
  creator: "0xUserWallet456...",
  blockNumber: 12345,
  transactionHash: "0xTxHash789...",
  logIndex: 5
}
```

### 4.4. Transaction Receipt Processing

**SDK Extracts Collection Address:**
- Function: `extractCollectionAddress(receipt)`
- Location: `src/modules/CollectionModule.ts`
- Method:
  1. Parse transaction receipt logs
  2. Find log matching `ERC721CollectionCreated` event signature
  3. Decode log data using ABI
  4. Extract `collectionAddress` parameter
  5. Return address to frontend

**Frontend Receives:**
```
{
  collectionAddress: "0xNewCollection123...",
  transactionHash: "0xTxHash789...",
  blockNumber: 12345,
  gasUsed: "650000"
}
```

**Gas Costs (Estimated):**
- ERC721 Collection: ~500,000 - 700,000 gas
- ERC1155 Collection: ~450,000 - 650,000 gas
- At 50 gwei: ~$15-30 USD equivalent

**Output từ bước này:**
- ✅ Collection deployed on-chain
- ✅ Collection address: `0xNewCollection123...`
- ✅ Ownership transferred to user
- ✅ Mint configuration set
- ✅ Royalty configured
- ✅ Event emitted for indexing

---

## 📍 BƯỚC 5: Index Events - Blockchain Indexer

### Repository: `zuno-marketplace-indexer`

**Technology:** Ponder Framework (v4.0 - Event-First Architecture)

### 5.1. Event Listener Configuration

**File:** `ponder.config.ts`

**Contract Registration:**
```
networks: {
  anvil: {
    chainId: 31337,
    rpcUrl: "http://127.0.0.1:8545"
  },
  sepolia: {
    chainId: 11155111,
    rpcUrl: process.env.SEPOLIA_RPC_URL
  }
}

contracts: {
  ERC721CollectionFactory: {
    address: "0xFactory123...",
    abi: ERC721CollectionFactoryABI,
    network: "sepolia",
    startBlock: 0,  // or specific block number
    eventFilter: {
      event: "ERC721CollectionCreated"
    }
  },
  ERC1155CollectionFactory: {
    address: "0xFactory456...",
    abi: ERC1155CollectionFactoryABI,
    network: "sepolia",
    eventFilter: {
      event: "ERC1155CollectionCreated"
    }
  }
}
```

**Polling Mechanism:**
- Ponder polls RPC every 1-5 seconds (configurable)
- Fetches new blocks
- Parses logs matching registered contracts
- Calls appropriate handlers

### 5.2. Event Handler Execution

**Handler Registration:**
- File: `src/domain/collection/index.ts`
- Function: `registerCollectionHandlers(ponder)`

**Handler Files:**
- `src/domain/collection/handlers/erc721-created.handler.ts`
- `src/domain/collection/handlers/erc1155-created.handler.ts`

**Handler Signature:**
```
Event Name: erc721collectionfactory_anvil:ERC721CollectionCreated
Handler: handleERC721Created({ event, context })
```

**Event Object Structure:**
```
event: {
  args: {
    collectionAddress: "0x...",
    creator: "0x..."
  },
  block: {
    number: 12345,
    timestamp: 1699999999,
    hash: "0x..."
  },
  transaction: {
    hash: "0x...",
    from: "0x...",
    to: "0xFactory...",
    gasUsed: "650000"
  },
  log: {
    address: "0xFactory...",
    logIndex: 5,
    data: "0x...",
    topics: ["0x..."]
  }
}

context: {
  db: DatabaseClient,
  network: "anvil",
  contracts: { ... }
}
```

### 5.3. Handler Processing Logic

**Flow trong Handler:**

#### 5.3.1. Extract Event Data
- Parse `event.args` cho collection address & creator
- Extract block timestamp, transaction hash, log index

#### 5.3.2. Prepare Event Data
**Current Implementation:**
- Static defaults: `{ name: "ERC721 Collection", symbol: "ERC721", tokenType: "ERC721" }`

**Future Enhancement (TODO):**
- Fetch on-chain data via RPC:
  - Call `collection.name()`
  - Call `collection.symbol()`
  - Call `collection.totalSupply()`
  - Call `collection.maxSupply()`
- Adds 3-4 RPC calls per event (~200ms latency)

#### 5.3.3. Validate Data
- Use Zod schema: `validateEventData("collection_created", data)`
- Location: `src/shared/schemas/event.schemas.ts`
- Ensures data matches expected format

#### 5.3.4. Store in Event Table
**Repository:** `EventRepository`
**Method:** `createEvent()`

**Database:** PostgreSQL
**Table:** `event`
**Schema:** JSONB-based flexible schema

**Columns:**
- `id` (TEXT, PK): `"{txHash}:{logIndex}"`
- `eventType` (TEXT): `"collection_created"`
- `category` (TEXT): `"collection"`
- `actor` (TEXT): Creator address
- `collection` (TEXT): Collection address
- `tokenId` (TEXT, nullable): null for collection events
- `data` (JSONB): Flexible event-specific data
- `contractAddress` (TEXT): Factory address
- `contractName` (TEXT): "ERC721CollectionFactory"
- `blockNumber` (BIGINT): 12345
- `blockTimestamp` (BIGINT): Unix timestamp
- `transactionHash` (TEXT): Transaction hash
- `logIndex` (INTEGER): Position in transaction logs
- `chainId` (INTEGER): 11155111 (Sepolia)
- `processedAt` (BIGINT): Indexer processing timestamp
- `version` (TEXT): "3.0"

**JSONB Data Structure:**
```
{
  "name": "ERC721 Collection",
  "symbol": "ERC721",
  "tokenType": "ERC721",
  "maxSupply": null
}
```

**Indexes for Performance:**
- `idx_event_type` on `eventType`
- `idx_category` on `category`
- `idx_actor` on `actor`
- `idx_collection` on `collection`
- `idx_block_timestamp` on `blockTimestamp DESC`
- `idx_chain_id` on `chainId`
- Composite: `(category, eventType, blockTimestamp DESC)`

#### 5.3.5. Update Account Cache
**Repository:** `AccountRepository`
**Methods:**
- `getOrCreate(address, timestamp)` - Ensure user record exists
- `incrementActivity(address, timestamp)` - Update activity

**Table:** `account`

**Fields:**
- `address` (TEXT, PK): User wallet address
- `firstSeenAt` (BIGINT): First interaction timestamp
- `lastActiveAt` (BIGINT): Most recent activity
- `eventCount` (INTEGER): Total events participated

**Purpose:**
- Fast user lookups
- Activity tracking
- No complex aggregations needed (calculated from events on-demand)

### 5.4. REST API Exposure

**File:** `src/api/index.ts`

**Server:** Hono (lightweight Express alternative)

**Endpoints:**

#### 5.4.1. Get Activity Feed
```
GET /api/activity?limit=50
```

**Query Params:**
- `limit` (number, default: 50, max: 100)
- `offset` (number, default: 0)
- `category` (string, optional): "collection", "listing", "sale"
- `eventType` (string, optional): "collection_created"

**Response:**
```
{
  "success": true,
  "data": [
    {
      "id": "0xtx:5",
      "eventType": "collection_created",
      "category": "collection",
      "actor": "0xCreator...",
      "collection": "0xNewCollection...",
      "tokenId": null,
      "data": {
        "name": "ERC721 Collection",
        "symbol": "ERC721",
        "tokenType": "ERC721"
      },
      "contractAddress": "0xFactory...",
      "contractName": "ERC721CollectionFactory",
      "blockNumber": "12345",
      "blockTimestamp": "1699999999",
      "transactionHash": "0xtx",
      "logIndex": 5,
      "chainId": 31337
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 150
  }
}
```

#### 5.4.2. Get Events
```
GET /api/events?category=collection&limit=50
```

**Query Params:**
- Same as `/api/activity`
- Additional filters: `actor`, `collection`

**Use Case:**
- More specific filtering
- Developer API access

### 5.5. Real-time Updates via Polling

**Polling Frequency:**
- Ponder: 1-5 seconds (blockchain → indexer)
- Frontend: 30 seconds (indexer → UI)

**Latency:**
- Block confirmation: ~12 seconds (Ethereum)
- Indexer processing: <1 second
- Frontend refresh: 0-30 seconds
- **Total: ~12-42 seconds** from transaction to UI update

**Output từ bước này:**
- ✅ Event stored in PostgreSQL
- ✅ Account cache updated
- ✅ REST API exposes event data
- ✅ Ready for frontend consumption

---

## 📍 BƯỚC 6: Sync to Database - API Service

### Repository: `zuno-marketplace-api`

### ⚠️ CURRENT STATUS: NOT IMPLEMENTED

**What SHOULD Exist:**
- Collection Service (gRPC)
- Proto definition (`proto/collection.proto`)
- Database migration (`db/migrations/000003_add_collections.up.sql`)
- RabbitMQ consumer
- GraphQL schema

### 6.1. Architecture Design (Planned)

**Service Structure:**
```
services/collection-service/
├── cmd/
│   └── main.go                    # Service entrypoint (:50054)
├── internal/
│   ├── config/
│   │   └── config.go             # Load from env
│   ├── models/
│   │   └── collection.go         # GORM models
│   ├── repository/
│   │   ├── collection_repository.go       # Interface
│   │   └── collection_repository_impl.go  # PostgreSQL implementation
│   ├── consumer/
│   │   └── collection_consumer.go # RabbitMQ event handler
│   ├── server/
│   │   └── collection_server.go  # gRPC service implementation
│   └── client/
│       └── user_client.go        # Call User Service
```

### 6.2. Event Flow (Planned)

**RabbitMQ Integration:**

#### 6.2.1. Indexer Publishes Event
**After storing in event table:**
- Exchange: `marketplace.events`
- Routing Key: `collection.created`
- Message:
```
{
  "eventType": "collection_created",
  "collectionAddress": "0xNewCollection...",
  "creator": "0xCreator...",
  "chainId": 11155111,
  "transactionHash": "0xtx",
  "blockNumber": 12345,
  "blockTimestamp": 1699999999,
  "data": {
    "name": "ERC721 Collection",
    "symbol": "ERC721",
    "tokenType": "ERC721"
  }
}
```

#### 6.2.2. Collection Service Consumes
**Consumer:** `internal/consumer/collection_consumer.go`

**Processing Steps:**

**1. Validate Message:**
- Check required fields
- Validate addresses
- Ensure not duplicate (check if collection exists)

**2. Fetch On-Chain Data:**
- RPC calls to collection contract:
  - `name()` → "My Cool NFT Collection"
  - `symbol()` → "COOL"
  - `totalSupply()` → 0 (initially)
  - `maxSupply()` → 10000
  - `owner()` → "0xCreator..."
  - `royaltyInfo(0, 1000000)` → (receiver, fee)

**Why fetch again?**
- Indexer only stores event data (minimal)
- API needs complete collection info
- Single source of truth: blockchain

**3. Lookup Creator User ID:**
- Call User Service gRPC:
  - `GetUserByWallet(address: "0xCreator...")`
  - Returns: `{ userId: "uuid-123", status: "active" }`
- If user not found: Create user via `CreateUserFromWallet()`

**4. Generate Collection Slug:**
- Function: `generateSlug(name string) string`
- Logic:
  - Lowercase: "My Cool NFT Collection" → "my cool nft collection"
  - Replace spaces with hyphens: "my-cool-nft-collection"
  - Remove special chars: Keep only `[a-z0-9-]`
  - Ensure uniqueness: Append number if exists ("my-cool-nft-collection-2")
  - Validate: 3-100 chars

**5. Store in Database:**

**Table:** `collections`

**Schema:**
```
id                  UUID PRIMARY KEY DEFAULT gen_random_uuid()
chain_id            TEXT NOT NULL
contract_address    TEXT NOT NULL
slug                TEXT UNIQUE NOT NULL
name                TEXT NOT NULL
symbol              TEXT NOT NULL
description         TEXT
category            TEXT
image_url           TEXT
banner_url          TEXT
website_url         TEXT
social_links_json   JSONB
is_verified         BOOLEAN DEFAULT false
is_hidden           BOOLEAN DEFAULT false
source              TEXT DEFAULT 'factory'  -- 'factory' or 'external'
total_supply        INTEGER DEFAULT 0
max_supply          INTEGER
royalty_bps         INTEGER
royalty_receiver    TEXT
metadata_standard   TEXT  -- 'ERC721' or 'ERC1155'
status              TEXT DEFAULT 'active'  -- 'active', 'hidden', 'banned'
deployed_block      BIGINT
index_status        TEXT DEFAULT 'synced'
creator_id          UUID REFERENCES users(user_id)
created_at          TIMESTAMP DEFAULT NOW()
updated_at          TIMESTAMP DEFAULT NOW()

UNIQUE(chain_id, contract_address)
INDEX idx_collections_slug ON collections(slug)
INDEX idx_collections_creator ON collections(creator_id)
INDEX idx_collections_chain_address ON collections(chain_id, contract_address)
```

**Insert:**
```
INSERT INTO collections (
  chain_id, contract_address, slug, name, symbol,
  total_supply, max_supply, royalty_bps, royalty_receiver,
  metadata_standard, source, deployed_block, creator_id, status
) VALUES (
  'eip155:11155111',
  '0xNewCollection...',
  'my-cool-nft-collection',
  'My Cool NFT Collection',
  'COOL',
  0, 10000, 750, '0xRoyaltyReceiver...',
  'ERC721', 'factory', 12345,
  'uuid-creator', 'active'
)
```

**6. Update User Stats:**
```
UPDATE user_stats
SET collections_count = collections_count + 1
WHERE user_id = 'uuid-creator'
```

**7. Acknowledge Message:**
- RabbitMQ ACK to remove from queue
- If error: NACK and requeue (with retry limit)

### 6.3. gRPC Service Interface (Planned)

**Proto Definition:** `proto/collection.proto`

**Service Methods:**
```
service CollectionService {
  // Create collection record (called by consumer)
  rpc CreateCollection(CreateCollectionRequest) returns (CreateCollectionResponse);

  // Get collection by address or slug
  rpc GetCollection(GetCollectionRequest) returns (GetCollectionResponse);

  // List collections with filters
  rpc ListCollections(ListCollectionsRequest) returns (ListCollectionsResponse);

  // Update collection metadata (owner only)
  rpc UpdateCollection(UpdateCollectionRequest) returns (UpdateCollectionResponse);

  // Get user's collections
  rpc GetUserCollections(GetUserCollectionsRequest) returns (GetUserCollectionsResponse);
}
```

**Message Types:**
```
message Collection {
  string id = 1;
  string chain_id = 2;
  string contract_address = 3;
  string slug = 4;
  string name = 5;
  string symbol = 6;
  string description = 7;
  string creator_id = 8;
  int32 total_supply = 9;
  int32 max_supply = 10;
  string status = 11;
  string metadata_standard = 12;
  int64 created_at = 13;
}
```

### 6.4. GraphQL Integration (Planned)

**Schema File:** `services/graphql-gateway/graph/schemas/collection.graphqls`

**Types:**
```
type Collection {
  id: ID!
  chainId: String!
  contractAddress: String!
  slug: String!
  name: String!
  symbol: String!
  description: String
  category: String
  imageUrl: String
  bannerUrl: String
  websiteUrl: String
  socialLinks: SocialLinks
  isVerified: Boolean!
  totalSupply: Int!
  maxSupply: Int
  royaltyBps: Int
  royaltyReceiver: String
  metadataStandard: String!
  status: String!
  creator: User!
  createdAt: String!
  updatedAt: String!
}

type SocialLinks {
  twitter: String
  discord: String
  instagram: String
}

input GetCollectionInput {
  id: ID
  slug: String
  contractAddress: String
}

extend type Query {
  # Get single collection
  getCollection(input: GetCollectionInput!): Collection

  # Get user's collections
  myCollections(limit: Int = 20, offset: Int = 0): [Collection!]!

  # List all collections with filters
  listCollections(
    category: String
    isVerified: Boolean
    sortBy: String
    limit: Int = 20
    offset: Int = 0
  ): CollectionConnection!
}

type CollectionConnection {
  nodes: [Collection!]!
  totalCount: Int!
  pageInfo: PageInfo!
}
```

**Resolver Implementation:**
- File: `services/graphql-gateway/graph/schema.resolvers.go`
- Calls: Collection Service gRPC client
- Auth: JWT middleware injects `user_id`
- Error handling: Convert gRPC errors to GraphQL errors

**Example Query:**
```
query GetCollection {
  getCollection(input: { slug: "my-cool-nft-collection" }) {
    id
    name
    symbol
    contractAddress
    chainId
    totalSupply
    maxSupply
    creator {
      id
      profile {
        username
        displayName
      }
    }
    createdAt
  }
}
```

**Output từ bước này (when implemented):**
- ✅ Collection stored in API database
- ✅ User stats updated
- ✅ GraphQL query available
- ✅ gRPC service for internal calls
- ✅ Collection indexed and searchable

---

## 📍 BƯỚC 7: Real-time Update - Event Stream

### Repository: `zuno-marketplace-event-stream`

**Purpose:** Display real-time marketplace events in UI

### 7.1. Frontend Polling Architecture

**Technology:**
- TanStack Query (React Query)
- Next.js App Router
- WebSocket (future enhancement)

**Hook:** `src/hooks/use-events.ts`

**Usage:**
```
Component renders
  → useEvents({ limit: 50 })
  → TanStack Query queries
  → Polls every 30 seconds
  → Updates UI on new data
```

### 7.2. Polling Configuration

**Constants:** `src/lib/constants.ts`

**Settings:**
- `POLLING_INTERVAL_MS: 30000` (30 seconds)
- `CACHE_STALE_TIME_MS: 25000` (25 seconds)
- `RETRY_COUNT: 3`
- `RETRY_DELAY_MS: 1000` (exponential backoff)

**Why 30 seconds?**
- Balance freshness vs server load
- Blockchain finality: ~12 seconds
- Indexer latency: ~1-5 seconds
- Total latency: ~15-20 seconds
- 30s gives buffer for processing

### 7.3. HTTP Client Implementation

**File:** `src/lib/services/ponder-client.ts`

**Class:** `PonderClient`

**Methods:**

#### 7.3.1. Get Activity
```
async getActivity(limit: number = 50): Promise<Event[]>
```

**Flow:**
1. Build URL: `${INDEXER_URL}/api/activity?limit=${limit}`
2. Add headers:
   - `Accept: application/json`
   - `If-None-Match: ${etag}` (if cached)
3. Fetch with timeout (10s)
4. Handle response:
   - `200 OK`: Parse JSON, update cache, return events
   - `304 Not Modified`: Return cached data (bandwidth saved)
   - `4xx/5xx`: Throw error, trigger retry

**ETag Caching:**
- Server sends: `ETag: "abc123"`
- Client stores in memory
- Next request: `If-None-Match: "abc123"`
- If no changes: `304 Not Modified` (no body)
- **Bandwidth savings:** 50-90% on repeated requests

#### 7.3.2. Get Events (Filtered)
```
async getEvents(options: {
  category?: string
  eventType?: string
  limit?: number
}): Promise<Event[]>
```

**Use Case:** Filter to collection events only

**Example:**
```
ponderClient.getEvents({
  category: "collection",
  limit: 20
})
```

### 7.4. React Component Integration

**Component Hierarchy:**
```
EventPage
  └─ EventFeed
      ├─ EventTicker (auto-scrolling)
      └─ EventItem (individual event card)
```

**Component Files:**
- `src/components/events/event-feed.tsx`
- `src/components/events/event-ticker.tsx`
- `src/components/events/event-item.tsx`

### 7.5. Event Display Logic

**Component:** `EventItem`

**Props:**
```
{
  id: "0xtx:5",
  eventType: "collection_created",
  category: "collection",
  actor: "0xCreator...",
  collection: "0xNewCollection...",
  data: {
    name: "My Cool NFT Collection",
    symbol: "COOL",
    tokenType: "ERC721"
  },
  blockTimestamp: "1699999999",
  transactionHash: "0xtx"
}
```

**Rendering Logic:**

**1. Icon Selection:**
```
category = "collection" → 🖼️
category = "listing" → 🏷️
category = "sale" → 💰
category = "offer" → 📝
```

**2. Color Theme:**
```
category = "collection" → text-orange-500
category = "listing" → text-blue-500
category = "sale" → text-green-500
category = "offer" → text-purple-500
```

**3. Event Title:**
```
eventType = "collection_created" → "Collection Created"
eventType = "listing_created" → "New Listing"
eventType = "sale_completed" → "Sale Completed"
```
- Function: `formatEventType(type)`
- Logic: Split by `_`, capitalize each word

**4. Address Formatting:**
```
"0x1234567890abcdef" → "0x1234...cdef"
```
- Function: `shortenAddress(address)`
- Takes first 6 and last 4 chars

**5. Timestamp Display:**
```
blockTimestamp = 1699999999 → "2 minutes ago"
```
- Library: `date-fns`
- Function: `formatDistanceToNow(timestamp)`
- Updates: Re-renders every 60 seconds

**6. Transaction Link:**
```
<a href="https://sepolia.etherscan.io/tx/{transactionHash}">
  View on Etherscan →
</a>
```

### 7.6. Real-time Animation

**Auto-scroll Ticker:**
- Component: `EventTicker`
- Animation: CSS `@keyframes scroll`
- Speed: 50px/second
- Infinite loop
- Pauses on hover

**Fade-in Animation:**
- New events slide in from right
- Transition: `transform 0.3s ease-out`
- Opacity: 0 → 1

**Skeleton Loading:**
- While loading: Show shimmer skeleton
- Component: `EventSkeleton`
- Prevents layout shift

### 7.7. Error Handling

**States:**
- `isLoading`: Initial fetch
- `isRefetching`: Background refresh
- `isError`: Fetch failed
- `error`: Error details

**Error Display:**
```
if (isError) {
  return (
    <ErrorBanner>
      Failed to load events. <button>Retry</button>
    </ErrorBanner>
  )
}
```

**Retry Logic:**
- Automatic: 3 retries with exponential backoff
- Manual: "Retry" button
- Fallback: Show cached data with warning

### 7.8. Performance Optimizations

**1. Request Deduplication:**
- TanStack Query dedupes concurrent requests
- If 5 components use `useEvents(50)`, only 1 API call

**2. Background Refetch:**
- `refetchInterval: 30000`
- `refetchOnWindowFocus: false` (avoid spam)
- `refetchOnReconnect: true`

**3. Cache Management:**
- `staleTime: 25000` (data fresh for 25s)
- `cacheTime: 300000` (keep cache 5 min)
- Automatic garbage collection

**4. Conditional Rendering:**
- Only render visible events (virtualization possible)
- Limit: Max 50 events displayed
- Pagination: "Load more" button

**Output từ bước này:**
- ✅ Events displayed in UI within ~30 seconds
- ✅ Auto-refresh every 30 seconds
- ✅ Smooth animations
- ✅ Optimized bandwidth (ETag caching)
- ✅ Error handling with retry

---

## 📍 BƯỚC 8: Notify Users - Notification Service

### Repository: `zuno-marketplace-notifications`

**Purpose:** Alert users about collection drops, mints, and activity

### 8.1. Drop Management System

**Database Schema:** `src/infrastructure/database/prisma/schema.prisma`

**Core Tables:**

#### Table: `Drop`
```
id              UUID PRIMARY KEY
organizationId  UUID
collectionId    TEXT  -- Collection contract address
dropName        TEXT
description     TEXT
startTime       TIMESTAMP
endTime         TIMESTAMP (nullable)
status          TEXT  -- 'announced', 'live', 'ended'
totalSupply     INTEGER (nullable)
pricePerNFT     DECIMAL (nullable)
imageUrl        TEXT
mintUrl         TEXT
createdAt       TIMESTAMP
updatedAt       TIMESTAMP
```

**Purpose:** Track collection mint campaigns

#### Table: `Watchlist`
```
id        UUID PRIMARY KEY
userId    UUID  -- References User
itemType  TEXT  -- 'collection', 'nft', 'creator'
itemId    TEXT  -- Collection address or NFT ID
createdAt TIMESTAMP
```

**Purpose:** Users subscribe to collections for alerts

#### Table: `WhitelistEntry`
```
id              UUID PRIMARY KEY
dropId          UUID  -- References Drop
userId          UUID
organizationId  UUID
spots           INTEGER  -- How many mint slots
isApproved      BOOLEAN
hasNotified     BOOLEAN  -- Already sent approval notification
approvedAt      TIMESTAMP (nullable)
createdAt       TIMESTAMP
```

**Purpose:** Manage allowlist for presale

### 8.2. Notification Triggers

**Service:** `src/core/services/drop-notification.service.ts`

**Event Sources:**
1. RabbitMQ: Collection created event
2. Cron jobs: Time-based triggers (5min before drop, drop live)
3. Manual: Admin announces drop

### 8.3. Drop Announcement Flow

**Trigger:** Collection created event consumed from RabbitMQ

**Flow:**

#### 8.3.1. Create Drop Record
```
Function: announceDrop(data: AnnounceDropDto)

Input:
{
  collectionId: "0xNewCollection...",
  organizationId: "org-uuid",
  dropName: "My Cool NFT Collection",
  description: "1000 unique NFTs",
  startTime: "2025-11-17T14:00:00Z",
  totalSupply: 1000,
  pricePerNFT: "0.01",
  imageUrl: "ipfs://Qm...abc123",
  mintUrl: "https://zuno.com/mint/0xNewCollection..."
}

Process:
1. Validate input (Zod schema)
2. Create Drop record in database
3. Set status: "announced"
4. Schedule pre-drop notifications
5. Find watchlist users
6. Send notifications
```

#### 8.3.2. Find Watchlist Users
```
Query:
SELECT user_id FROM watchlist
WHERE item_type = 'collection'
  AND item_id = '0xNewCollection...'

OR

WHERE item_type = 'creator'
  AND item_id = '{creator_address}'
```

**Watchers Can Be:**
- Users watching this specific collection
- Users watching the creator's profile
- Users subscribed to category (e.g., "Art")

#### 8.3.3. Send Notifications

**Channels:**

**1. Email Notification:**
- Service: NodeMailer
- Template: `templates/drop-announced.hbs` (Handlebars)
- Subject: "🚀 New Drop: {dropName}"
- Content:
  - Collection image
  - Drop name, description
  - Start time (user's timezone)
  - Price, total supply
  - CTA: "View Drop" button → mintUrl

**2. WebSocket Push (Real-time):**
- For users currently online
- Socket.io room: `user:{userId}`
- Payload:
```
{
  type: "drop_announced",
  data: {
    dropId: "drop-uuid",
    dropName: "My Cool NFT Collection",
    imageUrl: "ipfs://...",
    startTime: "2025-11-17T14:00:00Z",
    mintUrl: "https://..."
  }
}
```
- Frontend: Toast notification appears

**3. Push Notification (Mobile - if integrated):**
- Service: Firebase Cloud Messaging
- Title: "New Drop: {dropName}"
- Body: "Starts in {timeUntil}"
- Action: Open app to mint page

**4. Discord Webhook (if configured):**
- Organization's Discord server
- Channel: #drops
- Embed:
  - Title: Drop name
  - Description: Description
  - Image: Collection image
  - Fields: Price, supply, start time
  - Button: "Mint Now"

### 8.4. Time-based Notifications

**Cron Jobs:** BullMQ scheduled jobs

#### 8.4.1. Drop Starting Soon (5 minutes before)
```
Schedule: Check every 1 minute

Query:
SELECT * FROM drops
WHERE status = 'announced'
  AND start_time BETWEEN NOW() AND NOW() + INTERVAL '6 minutes'
  AND start_time > NOW() + INTERVAL '4 minutes'

For each drop:
  1. Update status: 'starting_soon'
  2. Find watchlist users
  3. Send notification:
     - Subject: "⏰ Drop starts in 5 minutes!"
     - Content: Reminder to prepare wallet, gas fees
     - CTA: "Get Ready"
```

#### 8.4.2. Drop Live
```
Schedule: Check every 30 seconds

Query:
SELECT * FROM drops
WHERE status IN ('announced', 'starting_soon')
  AND start_time <= NOW()

For each drop:
  1. Update status: 'live'
  2. Find watchlist users + whitelist users
  3. Send notification:
     - Subject: "🔥 Drop is LIVE!"
     - Content: Mint now before sold out
     - CTA: "Mint Now" → direct to mint page
```

#### 8.4.3. Drop Ended
```
Schedule: Check every 5 minutes

Query:
SELECT * FROM drops
WHERE status = 'live'
  AND (end_time <= NOW() OR current_supply >= total_supply)

For each drop:
  1. Update status: 'ended'
  2. Calculate stats: Total minted, revenue, unique minters
  3. Send summary to creator
```

### 8.5. Whitelist Notifications

#### 8.5.1. Whitelist Approved
```
Trigger: Admin approves whitelist entry

Function: notifyWhitelistApproved(entryId: string)

Flow:
1. Get WhitelistEntry by ID
2. Get Drop details
3. Get User email/preferences
4. Send notification:
   - Subject: "✅ You're on the whitelist!"
   - Content:
     - Drop name
     - Your spots: {spots}
     - Presale start time
     - Presale price (if different)
   - CTA: "View Details"
5. Update: hasNotified = true
```

#### 8.5.2. Whitelist Presale Reminder
```
Schedule: 1 hour before presale starts

For whitelist users:
  - Reminder to connect wallet
  - Check gas fees
  - Prepare payment
```

### 8.6. Mint Success/Failure Notifications

#### 8.6.1. Mint Success
```
Trigger: Indexer detects Minted event

Function: notifyMintSuccess(data: {
  userId: string
  collectionId: string
  tokenId: string
  transactionHash: string
})

Notification:
  - Subject: "🎉 Mint successful!"
  - Content:
    - Collection name
    - Token ID: #{tokenId}
    - View NFT button → marketplace
    - Share on social media options
```

#### 8.6.2. Mint Failed
```
Trigger: Frontend reports failed transaction

Function: notifyMintFailed(data: {
  userId: string
  collectionId: string
  errorMessage: string
})

Notification:
  - Subject: "❌ Mint failed"
  - Content:
    - Error message (user-friendly)
    - Troubleshooting tips
    - Support link
  - CTA: "Try Again"
```

### 8.7. User Preferences

**Table:** `NotificationPreferences`
```
userId          UUID PRIMARY KEY
emailEnabled    BOOLEAN DEFAULT true
pushEnabled     BOOLEAN DEFAULT true
discordEnabled  BOOLEAN DEFAULT false
categories      JSONB  -- ["drops", "sales", "offers"]
frequency       TEXT   -- "instant", "daily_digest", "weekly_digest"
```

**Respect User Settings:**
- Check preferences before sending
- Allow opt-out per category
- Unsubscribe link in all emails
- GDPR compliance: Right to be forgotten

### 8.8. Notification Queue System

**Technology:** BullMQ with Redis

**Queues:**
- `drop-notifications` (high priority)
- `email-notifications` (medium priority)
- `digest-notifications` (low priority, batched)

**Why Queue?**
- Avoid overwhelming email service (rate limits)
- Batch processing for efficiency
- Retry failed sends
- Track delivery status

**Worker Configuration:**
- Concurrency: 10 emails simultaneously
- Rate limit: 100 emails/minute (SendGrid limit)
- Retry: 3 attempts with exponential backoff
- Dead letter queue: Failed notifications for manual review

**Output từ bước này:**
- ✅ Drop announced to watchlist users
- ✅ Time-based alerts (5min before, live)
- ✅ Whitelist notifications
- ✅ Mint success/failure alerts
- ✅ Multi-channel delivery (email, websocket, push)
- ✅ User preferences respected

---

## 🔗 LIÊN KẾT GIỮA CÁC REPOSITORIES

### Data Flow Map

```
┌─────────────────────────────────────────────────────────────────┐
│                         USER ACTION                              │
└────────────┬────────────────────────────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────────────────────────┐
│ zuno-marketplace-ui                                             │
│ - Component: CreateCollectionForm / CollectionForm             │
│ - Validation: Zod schema (MintTerminalCreateFormSchema)        │
│ - State: React Hook Form                                       │
│ - Output: Form data object                                     │
└────────────┬───────────────────────────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────────────────────────┐
│ zuno-marketplace-metadata                                       │
│ - Endpoint: POST /api/media, POST /api/metadata                │
│ - Storage: ImageKit (CDN) + Database (PostgreSQL)              │
│ - Queue: BullMQ → media-ipfs-pin, metadata-ipfs-pin            │
│ - IPFS: Pinata (background worker)                             │
│ - Output: { imageUrl, ipfsUrl, metadataUrl }                   │
└────────────┬───────────────────────────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────────────────────────┐
│ zuno-marketplace-sdk                                            │
│ - Module: CollectionModule                                     │
│ - Hook: useCollection() [React]                                │
│ - Function: createERC721Collection(params)                     │
│ - Registry: ContractRegistry (ABI caching)                     │
│ - Client: Calls UserHub.getFactoryFor("ERC721")                │
│ - Output: { collectionAddress, transactionHash, receipt }      │
└────────────┬───────────────────────────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────────────────────────┐
│ zuno-marketplace-abis                                           │
│ - Endpoint: GET /api/abis/{contractType}                       │
│ - Purpose: Provide ABI for SDK                                 │
│ - Cache: TanStack Query (5 min stale time)                     │
│ - Output: ABI JSON for contract interaction                    │
└────────────┬───────────────────────────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────────────────────────┐
│ zuno-marketplace-contracts                                      │
│ - Contract: ERC721CollectionFactory / ERC1155CollectionFactory │
│ - Function: createERC721Collection(CollectionParams)           │
│ - Pattern: Minimal Proxy (Clones.clone)                        │
│ - Storage: Collection state initialized                        │
│ - Event: ERC721CollectionCreated(address, address)             │
│ - Output: Collection deployed on blockchain                    │
└────────────┬───────────────────────────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────────────────────────┐
│ zuno-marketplace-indexer                                        │
│ - Framework: Ponder v4.0                                       │
│ - Listener: ERC721CollectionCreated event                      │
│ - Handler: handleERC721Created()                               │
│ - Storage: PostgreSQL (event table + account cache)            │
│ - API: GET /api/activity, GET /api/events                      │
│ - Queue: RabbitMQ publish (collection.created)                 │
│ - Output: Event data available via REST API                    │
└────────────┬───────────────────────────────────────────────────┘
             │
             ├─────────────────────────────────┐
             │                                 │
             ▼                                 ▼
┌─────────────────────────────┐   ┌──────────────────────────────┐
│ zuno-marketplace-api        │   │ zuno-marketplace-event-stream│
│ ⚠️ NOT IMPLEMENTED          │   │ - Hook: useEvents()          │
│                             │   │ - Client: PonderClient       │
│ - Consumer: RabbitMQ        │   │ - Polling: Every 30s         │
│ - Service: CollectionService│   │ - Cache: ETag-based          │
│ - gRPC: :50054              │   │ - Component: EventFeed       │
│ - GraphQL: collection.graphqls  │ - Display: 🖼️ orange card   │
│ - Database: collections table   │ - Animation: Fade-in + scroll│
│ - Output: GraphQL queries   │   │ - Output: Real-time UI       │
└─────────────────────────────┘   └──────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────────────────────────┐
│ zuno-marketplace-notifications                                  │
│ - Consumer: RabbitMQ (collection.created)                      │
│ - Service: DropNotificationService                             │
│ - Functions: announceDrop(), notifyDropLive()                  │
│ - Channels: Email, WebSocket, Push, Discord                    │
│ - Queue: BullMQ (email queue)                                  │
│ - Database: Drop, Watchlist, WhitelistEntry                    │
│ - Output: Users notified about new collection                  │
└────────────────────────────────────────────────────────────────┘
```

### Repository Dependencies

**zuno-marketplace-ui:**
- Depends on: `zuno-marketplace-sdk` (npm package)
- Depends on: `zuno-marketplace-metadata` (API calls)
- Depends on: `zuno-marketplace-api` (GraphQL - future)
- Uses: Wagmi, RainbowKit, SIWE for wallet

**zuno-marketplace-sdk:**
- Depends on: `zuno-marketplace-abis` (ABI registry)
- Depends on: `zuno-marketplace-contracts` (deployed addresses)
- Uses: Wagmi, Ethers, TanStack Query

**zuno-marketplace-metadata:**
- Standalone service
- External: ImageKit, Pinata, Redis
- Database: PostgreSQL (Drizzle ORM)

**zuno-marketplace-contracts:**
- Standalone (deployed on blockchain)
- Framework: Foundry

**zuno-marketplace-indexer:**
- Depends on: `zuno-marketplace-contracts` (ABIs, events)
- Publishes to: RabbitMQ
- Database: PostgreSQL
- Framework: Ponder

**zuno-marketplace-api:**
- Consumes: RabbitMQ (from indexer)
- Calls: User Service, Wallet Service (gRPC)
- Database: PostgreSQL
- Exposes: GraphQL, gRPC

**zuno-marketplace-event-stream:**
- Depends on: `zuno-marketplace-indexer` (REST API)
- Framework: Next.js, TanStack Query

**zuno-marketplace-notifications:**
- Consumes: RabbitMQ (from indexer)
- Database: PostgreSQL (Prisma)
- External: NodeMailer, Firebase, Discord

**zuno-marketplace-abis:**
- Standalone ABI registry
- Database: PostgreSQL (Drizzle ORM)

**zuno-marketplace-admin:**
- Generic admin UI (Drizzle Studio wrapper)

**zuno-marketplace-mini:**
- Testing app using `zuno-marketplace-sdk`
- Redux state management

---

## 🎯 IMPLEMENTATION GAPS & NEXT STEPS

### ✅ Đã Hoàn Thành

1. **Frontend UI**: 2 forms (simple + advanced) với validation đầy đủ
2. **Metadata Service**: Upload images + IPFS pinning hoàn chỉnh
3. **Smart Contracts**: Factory + Collection contracts deployed và tested
4. **SDK**: TypeScript SDK với React hooks
5. **Indexer**: Event listening + storage + REST API
6. **Event Stream**: Real-time UI updates
7. **Notifications**: Drop announcements + alerts

### ❌ Còn Thiếu (Critical)

1. **Collection Service (API)**:
   - Tạo `proto/collection.proto`
   - Implement gRPC service
   - Database migration (collections table)
   - RabbitMQ consumer
   - GraphQL schema

2. **Frontend Integration**:
   - Wire form submit to SDK
   - Transaction signing flow
   - ProcessDialog state management
   - Error handling + retry

3. **End-to-End Testing**:
   - Integration tests
   - Gas cost optimization
   - Security audit

### 📋 Implementation Checklist

**Phase 1: Collection Service (Backend)** [~2-3 days]
- [ ] Define `proto/collection.proto`
- [ ] Run `make proto`
- [ ] Create migration `000003_add_collections.up.sql`
- [ ] Implement `collection-service` (following user-service pattern)
- [ ] Add RabbitMQ consumer
- [ ] Test with manual events

**Phase 2: GraphQL Integration** [~1 day]
- [ ] Create `collection.graphqls` schema
- [ ] Generate GraphQL types
- [ ] Implement resolvers
- [ ] Add gRPC client to gateway
- [ ] Test queries/mutations

**Phase 3: Frontend Integration** [~2 days]
- [ ] Wire SDK to form submit
- [ ] Add wallet signature confirmation
- [ ] Implement ProcessDialog states
- [ ] Error handling
- [ ] Success redirect to collection page

**Phase 4: Testing & Optimization** [~2 days]
- [ ] Unit tests (service layer)
- [ ] Integration tests (full flow)
- [ ] Gas optimization review
- [ ] Security audit
- [ ] Load testing

**Phase 5: Documentation** [~1 day]
- [ ] Update API docs
- [ ] Frontend integration guide
- [ ] Video tutorial

**Total Estimated Time:** ~8-10 days

---

## 📊 METRICS & MONITORING

### Performance Metrics

**Transaction Times:**
- Smart contract deployment: ~650k gas (~15-30 seconds)
- IPFS upload: ~5-30 seconds (background)
- Indexer latency: ~1-5 seconds
- API sync: ~2-10 seconds
- Frontend update: ~30 seconds (polling)

**Total Time (User Perspective):**
- Click "Deploy" → Transaction confirmed: ~15-30 seconds
- Confirmed → Appears in event feed: ~30-60 seconds
- **Total:** ~45-90 seconds end-to-end

### Cost Breakdown

**Gas Costs:**
- ERC721 Collection: ~500k-700k gas
- ERC1155 Collection: ~450k-650k gas
- At 50 gwei: ~$15-30 USD

**Infrastructure Costs:**
- Pinata IPFS: ~$0.10/GB/month
- ImageKit CDN: ~$0.05/GB bandwidth
- RabbitMQ: Minimal (self-hosted)
- PostgreSQL: Minimal (self-hosted)
- Indexer RPC: ~$0.01/1000 requests

### Monitoring Points

**Track These Metrics:**
1. Collection creation success rate
2. Average gas cost
3. IPFS pin time (median, p95)
4. Indexer sync latency
5. API response time
6. Event stream refresh rate
7. Notification delivery rate
8. Error rate by step

---

## 🔒 SECURITY CONSIDERATIONS

### Smart Contract Security

1. **Initializer Protection**: OpenZeppelin's `initializer` prevents re-initialization
2. **Access Control**: Owner-only functions (Ownable)
3. **Royalty Limits**: Max 10% (1000 bps) enforced
4. **Parameter Validation**: Revert on invalid inputs
5. **Audit**: Recommend professional audit before mainnet

### API Security

1. **Authentication**: JWT tokens (SIWE)
2. **Authorization**: User can only modify own collections
3. **IDOR Protection**: User-scoped queries
4. **Rate Limiting**: API key-based limits
5. **Input Validation**: Zod schemas throughout
6. **SQL Injection**: Parameterized queries (GORM, Drizzle)

### IPFS Security

1. **Content Addressing**: Immutable hashes
2. **Pinning**: Ensure data persists (Pinata)
3. **Gateway Fallback**: Multiple IPFS gateways
4. **CDN Backup**: ImageKit as fallback

---

## 🎓 KEY LEARNINGS

### Architectural Patterns

1. **Event-First Architecture**: Single source of truth (event table)
2. **Minimal Proxy Pattern**: 98% gas savings
3. **Background Jobs**: Non-blocking user experience
4. **Polling vs WebSocket**: Polling sufficient for 30s refresh
5. **ETag Caching**: 50-90% bandwidth savings
6. **JSONB Storage**: Flexible schema for events

### Technology Choices

1. **Ponder Framework**: Best-in-class indexer (vs The Graph)
2. **TanStack Query**: Excellent caching + polling
3. **BullMQ**: Reliable job queue
4. **Zod**: Type-safe validation across stack
5. **gRPC**: Efficient inter-service communication
6. **GraphQL**: Flexible frontend queries

### Scalability Considerations

1. **Horizontal Scaling**: Indexer + API can scale independently
2. **Database Indexing**: Critical for query performance
3. **Cache Strategy**: Redis for hot data, PostgreSQL for cold
4. **Job Queue**: Prevents overwhelming external services
5. **CDN**: Reduces origin server load (ImageKit)

---

## 🤝 SUPPORT & TROUBLESHOOTING

### Common Issues

**Issue 1: Transaction Fails**
- Check: Wallet has sufficient ETH for gas
- Check: Parameters valid (maxSupply > 0, etc.)
- Check: Factory address correct for network
- Solution: Validate inputs, check network

**Issue 2: IPFS Upload Slow**
- Expected: 5-30 seconds for large files
- Solution: Background job, user gets immediate response
- Fallback: CDN URL available immediately

**Issue 3: Event Not Appearing in Feed**
- Check: Indexer running and synced
- Check: RPC connection healthy
- Check: Event emitted (check Etherscan)
- Solution: Wait 30-60 seconds, refresh manually

**Issue 4: Notifications Not Sent**
- Check: User subscribed to collection/creator
- Check: Notification preferences enabled
- Check: Email service (SendGrid) operational
- Solution: Check worker logs, retry failed jobs

---

## 📚 REFERENCES

### Documentation Links

- **Ponder Framework**: https://ponder.sh/docs
- **OpenZeppelin Contracts**: https://docs.openzeppelin.com/contracts
- **Wagmi**: https://wagmi.sh
- **TanStack Query**: https://tanstack.com/query
- **Pinata IPFS**: https://docs.pinata.cloud
- **ImageKit**: https://docs.imagekit.io

### Internal Docs

- CLAUDE.md - Project guidelines
- database-schema.md - Complete schema reference
- Each service has README.md with setup instructions

---

**Document Version:** 1.0
**Last Updated:** 2025-11-18
**Author:** Claude Code Assistant
**Status:** Comprehensive flow documented, implementation pending