# PHASE 3: GRAPHQL GATEWAY INTEGRATION

**Phase**: 3 of 6
**Dependencies**: Phase 2 (Collection Service) complete
**Estimated Effort**: ~800 lines (Go + GraphQL)
**Output**: GraphQL API at `/graphql`

---

## 🎯 Overview

Add collection queries/mutations to GraphQL Gateway:
- ✅ Define GraphQL schema (collection.graphqls)
- ✅ Generate Go code with gqlgen
- ✅ Implement resolvers (call collection service via gRPC)
- ✅ Add gRPC client for collection service
- ✅ Auth middleware integration

---

## 📋 Implementation Checklist

### Step 1: GraphQL Schema ✅
**File**: `services/graphql-gateway/graph/schemas/collection.graphqls`

```graphql
# Enums
enum TokenStandard { ERC721, ERC1155 }
enum CollectionStatus { PENDING, DEPLOYED, FAILED, ARCHIVED }

# Types
type Collection {
  id: ID!
  userId: ID!
  slug: String
  name: String!
  symbol: String!
  contractAddress: String
  chainId: String
  tokenStandard: TokenStandard!
  status: CollectionStatus!
  metadata: CollectionMetadata
  stats: CollectionStats
  createdAt: Time!
  deployedAt: Time
}

type CollectionConnection {
  items: [Collection!]!
  pageInfo: PageInfo!
}

# Inputs
input CreateCollectionInput { ... }
input UpdateCollectionInput { ... }

# Queries
extend type Query {
  collection(id: ID!): Collection
  myCollections(page: Int, limit: Int): CollectionConnection!
  collections(...): CollectionConnection!
}

# Mutations
extend type Mutation {
  createCollection(input: CreateCollectionInput!): Collection!
  updateCollection(id: ID!, input: UpdateCollectionInput!): Collection!
  addToAllowlist(input: AddToAllowlistInput!): Boolean!
  deleteCollection(id: ID!): Boolean!
}
```

**Generate code**:
```bash
cd services/graphql-gateway
go run github.com/99designs/gqlgen generate
```

---

### Step 2: gRPC Client Setup ✅
**File**: `services/graphql-gateway/internal/client/collection_client.go`

```go
type CollectionClient struct {
    conn   *grpc.ClientConn
    client pb.CollectionServiceClient
}

func NewCollectionClient(addr string) (*CollectionClient, error) {
    conn, err := grpc.Dial(addr, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    return &CollectionClient{
        conn:   conn,
        client: pb.NewCollectionServiceClient(conn),
    }, nil
}
```

**Add to main.go**:
```go
collectionClient, _ := client.NewCollectionClient(os.Getenv("COLLECTION_SERVICE_ADDR"))
```

---

### Step 3: Resolvers Implementation ✅
**File**: `services/graphql-gateway/graph/schema.resolvers.go`

**Query resolvers**:
```go
func (r *queryResolver) Collection(ctx, id) (*model.Collection, error) {
    // 1. Call collection service via gRPC
    resp, err := r.collectionClient.GetCollection(ctx, &pb.GetCollectionRequest{Id: id})
    // 2. Convert proto to GraphQL model
    return protoToGraphQL(resp.Collection), nil
}

func (r *queryResolver) MyCollections(ctx, page, limit) (*model.CollectionConnection, error) {
    // 1. Get user_id from context (auth middleware)
    userID := middleware.GetUserIDFromContext(ctx)
    // 2. Call collection service
    // 3. Build connection response with pagination
}
```

**Mutation resolvers**:
```go
func (r *mutationResolver) CreateCollection(ctx, input) (*model.Collection, error) {
    // 1. Get user_id from context (auth required!)
    userID := middleware.GetUserIDFromContext(ctx)
    if userID == "" {
        return nil, errors.New("unauthorized")
    }
    // 2. Convert GraphQL input to proto request
    // 3. Call collection service gRPC
    // 4. Return GraphQL model
}

func (r *mutationResolver) UpdateCollection(ctx, id, input) (*model.Collection, error) {
    // 1. Auth check
    // 2. Call collection service (ownership check happens in service)
    // 3. Return updated collection
}
```

---

### Step 4: Model Conversion ✅
**File**: `services/graphql-gateway/internal/mapper/collection_mapper.go`

```go
func ProtoToGraphQL(proto *pb.Collection) *model.Collection {
    return &model.Collection{
        ID:              proto.Id,
        UserID:          proto.UserId,
        Name:            proto.Name,
        Symbol:          proto.Symbol,
        TokenStandard:   mapTokenStandard(proto.TokenStandard),
        Status:          mapStatus(proto.Status),
        // ... map all fields
    }
}

func GraphQLToProto(input *model.CreateCollectionInput, userID string) *pb.CreateCollectionRequest {
    return &pb.CreateCollectionRequest{
        UserId:          userID,
        Name:            input.Name,
        Symbol:          input.Symbol,
        // ... map all fields
    }
}
```

---

### Step 5: Auth Middleware Integration ✅
**Verify**: `services/graphql-gateway/internal/middleware/auth.go`

**Protected resolvers**: All mutations + `myCollections` query
**Public resolvers**: `collection(id)`, `collections()` (list)

---

## ✅ Completion Checklist

- [ ] GraphQL schema defined
- [ ] gqlgen code generated
- [ ] Collection gRPC client initialized
- [ ] All query resolvers implemented
- [ ] All mutation resolvers implemented
- [ ] Proto ↔ GraphQL mappers working
- [ ] Auth middleware applied
- [ ] Ownership validation working
- [ ] Pagination working
- [ ] Error handling consistent
- [ ] GraphQL playground accessible
- [ ] Integration tests passing

---

## 🧪 Testing

**GraphQL Playground**: http://localhost:8081/playground

**Test query**:
```graphql
query {
  myCollections(page: 1, limit: 10) {
    items {
      id
      name
      status
      contractAddress
    }
    pageInfo {
      total
      hasNext
    }
  }
}
```

**Test mutation**:
```graphql
mutation {
  createCollection(input: {
    name: "My Collection"
    symbol: "MYC"
    tokenStandard: ERC721
    deployerAddress: "0x123..."
    imageUrl: "https://..."
  }) {
    id
    status
  }
}
```

---

## 🎯 Next Steps

After Phase 3 → **[Phase 4: Upload Proxy](./04-UPLOAD-PROXY.md)** (already complete!)

---

**Last Updated**: 2025-11-20
**Status**: Outline ready
