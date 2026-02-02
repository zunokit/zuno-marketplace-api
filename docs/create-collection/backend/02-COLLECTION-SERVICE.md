# PHASE 2: COLLECTION SERVICE (gRPC)

**Phase**: 2 of 6
**Dependencies**: Phase 1 (Database Schema) complete
**Estimated Effort**: ~2000 lines Go
**Output**: Collection service running on `:50054`

---

## 🎯 Overview

Create `collection-service` microservice with:
- ✅ gRPC server on port 50054
- ✅ Repository layer (GORM)
- ✅ Service layer (business logic)
- ✅ Proto definitions
- ✅ Docker integration

---

## 📋 Implementation Checklist

### Step 1: Proto Definitions ✅
**File**: `proto/collection.proto`
- [ ] Define enums (TokenStandard, CollectionStatus)
- [ ] Define messages (Collection, CollectionMetadata, CollectionStats)
- [ ] Define requests/responses (Create, Get, Update, List, Delete)
- [ ] Define CollectionService with 7 RPC methods
- [ ] Run `make proto` to generate Go code

**Key RPCs**:
```protobuf
service CollectionService {
  rpc CreateCollection(CreateCollectionRequest) returns (CreateCollectionResponse);
  rpc GetCollection(GetCollectionRequest) returns (GetCollectionResponse);
  rpc UpdateCollection(UpdateCollectionRequest) returns (UpdateCollectionResponse);
  rpc ListCollectionsByUser(ListCollectionsByUserRequest) returns (ListCollectionsResponse);
  rpc ListCollections(ListCollectionsRequest) returns (ListCollectionsResponse);
  rpc AddToAllowlist(AddToAllowlistRequest) returns (AddToAllowlistResponse);
  rpc DeleteCollection(DeleteCollectionRequest) returns (DeleteCollectionResponse);
}
```

---

### Step 2: Models (GORM) ✅
**File**: `services/collection-service/internal/models/collection.go`
- [ ] Define `Collection` struct matching database schema
- [ ] Define `CollectionMetadata` struct
- [ ] Define `CollectionStats` struct
- [ ] Define `JSONB` type for PostgreSQL
- [ ] Add GORM tags (types, indexes, constraints)

**Key structs**: `Collection`, `CollectionMetadata`, `CollectionStats`

---

### Step 3: Repository Layer ✅
**Files**: `services/collection-service/internal/repository/`

**Interfaces to implement**:
```go
type CollectionRepository interface {
    Create(ctx, *models.Collection) error
    Update(ctx, id string, updates map[string]interface{}) error
    GetByID(ctx, id string) (*models.Collection, error)
    GetByContractAddress(ctx, address, chainID string) (*models.Collection, error)
    ListByUser(ctx, userID string, page, limit int) ([]*models.Collection, int64, error)
    List(ctx, filters *ListFilters, page, limit int) ([]*models.Collection, int64, error)
    Delete(ctx, id string) error
}

type AllowlistRepository interface {
    AddWallets(ctx, collectionID string, wallets []*models.CollectionAllowlist) error
    GetByCollection(ctx, collectionID string) ([]*models.CollectionAllowlist, error)
    RemoveWallet(ctx, collectionID, walletAddress string) error
}
```

**Implementation**: `postgres_collection_repository.go`, `postgres_allowlist_repository.go`

---

### Step 4: Service Layer (Business Logic) ✅
**File**: `services/collection-service/internal/service/collection_service.go`

**Key methods**:
```go
type CollectionService struct {
    collectionRepo  repository.CollectionRepository
    allowlistRepo   repository.AllowlistRepository
    metadataRepo    repository.MetadataRepository
    activityRepo    repository.ActivityRepository
}

func (s *CollectionService) CreateCollection(ctx, req) (*models.Collection, error) {
    // 1. Validate user exists (optional: call User Service)
    // 2. Create collection with status=PENDING
    // 3. Triggers will auto-create stats and log activity
    // 4. Return collection
}

func (s *CollectionService) UpdateCollection(ctx, id, userID, updates) (*models.Collection, error) {
    // 1. Get existing collection
    // 2. Verify ownership (collection.UserID == userID)
    // 3. Apply updates
    // 4. Triggers will log STATUS_CHANGED activity
    // 5. Return updated collection
}
```

**Business rules**:
- Only owner can update/delete collection
- Cannot change status from DEPLOYED to PENDING
- Contract address required when status=DEPLOYED

---

### Step 5: gRPC Server Implementation ✅
**File**: `services/collection-service/internal/server/collection_server.go`

```go
type CollectionServer struct {
    pb.UnimplementedCollectionServiceServer
    service *service.CollectionService
    logger  *log.Logger
}

func (s *CollectionServer) CreateCollection(ctx, req) (*pb.CreateCollectionResponse, error) {
    // 1. Validate request
    // 2. Call service layer
    // 3. Convert model to proto
    // 4. Return response
}

// Implement all 7 RPC methods...
```

**Error handling**: Use gRPC status codes (NotFound, InvalidArgument, PermissionDenied, etc.)

---

### Step 6: Main Entrypoint ✅
**File**: `services/collection-service/cmd/main.go`

```go
func main() {
    // 1. Load config from .env
    // 2. Connect to PostgreSQL
    // 3. Initialize repositories
    // 4. Initialize service
    // 5. Create gRPC server
    // 6. Register CollectionServer
    // 7. Start listening on :50054
    // 8. Graceful shutdown
}
```

**Environment variables**:
```env
COLLECTION_GRPC_PORT=50054
DATABASE_URL=postgresql://...
```

---

### Step 7: Docker Integration ✅
**File**: `infra/development/docker/collection-service.Dockerfile`
**Update**: `docker-compose.yml`

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o collection-service ./services/collection-service/cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/collection-service /usr/local/bin/
CMD ["collection-service"]
```

---

## ✅ Completion Checklist

- [ ] Proto compiled successfully
- [ ] All models defined with GORM tags
- [ ] Repository layer implemented with tests
- [ ] Service layer implemented with business logic
- [ ] gRPC server implements all 7 RPCs
- [ ] Main entrypoint runs without errors
- [ ] Service starts on :50054
- [ ] Can create collection via gRPC call
- [ ] Can retrieve collection by ID
- [ ] Ownership validation working
- [ ] Docker container builds successfully
- [ ] Integration tests passing

---

## 🧪 Testing

**Unit tests**:
```bash
go test ./services/collection-service/internal/repository/...
go test ./services/collection-service/internal/service/...
```

**Integration test**:
```bash
go test ./services/collection-service/internal/server/...
```

**Manual test with grpcurl**:
```bash
grpcurl -plaintext \
  -d '{"user_id":"uuid","name":"Test","symbol":"TST","token_standard":1,"deployer_address":"0x123...","image_url":"https://..."}' \
  localhost:50054 \
  pb.CollectionService/CreateCollection
```

---

## 📖 Reference Implementation

See existing services for patterns:
- `services/auth-service/` - gRPC patterns
- `services/user-service/` - Repository patterns
- `services/wallet-service/` - Service patterns

---

## 🎯 Next Steps

After Phase 2 complete → **[Phase 3: GraphQL Gateway](./03-GRAPHQL-GATEWAY.md)**

---

**Last Updated**: 2025-11-20
**Status**: Outline ready for implementation
