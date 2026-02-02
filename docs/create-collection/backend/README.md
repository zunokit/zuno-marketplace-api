# BACKEND IMPLEMENTATION - DETAILED GUIDES

This folder contains step-by-step implementation guides for backend development.

---

## 📂 Phase Files

| Phase | File | Status | Complexity |
|-------|------|--------|------------|
| **Phase 1** | [01-DATABASE-SCHEMA.md](./01-DATABASE-SCHEMA.md) | ✅ Complete | Medium |
| **Phase 2** | [02-COLLECTION-SERVICE.md](./02-COLLECTION-SERVICE.md) | 📝 In Progress | High |
| **Phase 3** | [03-GRAPHQL-GATEWAY.md](./03-GRAPHQL-GATEWAY.md) | 📝 In Progress | Medium |
| **Phase 4** | [04-UPLOAD-PROXY.md](./04-UPLOAD-PROXY.md) | ✅ Complete | Low |
| **Phase 5** | [05-BLOCKCHAIN-INDEXER.md](./05-BLOCKCHAIN-INDEXER.md) | 📝 Planned | Medium |
| **Phase 6** | [06-TESTING.md](./06-TESTING.md) | 📝 Planned | High |

---

## 🔄 Development Workflow

### 1. Sequential Phases (Must Complete in Order)

```
Phase 1: Database Schema
  ↓ MUST complete
Phase 2: Collection Service (gRPC)
  ↓ MUST complete
Phase 3: GraphQL Gateway
  ↓ Can parallelize after this
```

### 2. Parallel Work (After Phase 3)

```
Phase 3 Complete
  ├─→ Phase 4: Upload Proxy (Backend Dev A)
  └─→ Phase 5: Blockchain Indexer (Backend Dev B)
       ↓ Both complete
     Phase 6: Testing
```

---

## 🎯 Quick Links

### Phase 1: Database Schema
**What**: Create 6 tables (collections, metadata, items, stats, activity, allowlist)
**Output**: Migrations applied, schema tested
**[→ View Guide](./01-DATABASE-SCHEMA.md)**

---

### Phase 2: Collection Service (gRPC)
**What**: Build collection microservice on :50054
**Includes**:
- Proto definitions (`proto/collection.proto`)
- Repository layer (GORM)
- Service layer (business logic)
- gRPC server
- Docker integration

**Output**: Collection service running

**File Structure**:
```
services/collection-service/
├── cmd/main.go
├── internal/
│   ├── models/
│   ├── repository/
│   ├── service/
│   └── server/
└── go.mod
```

**[→ See 01-BACKEND-PLAN.md for details](./../01-BACKEND-PLAN.md)**

---

### Phase 3: GraphQL Gateway Integration
**What**: Add collection queries/mutations to GraphQL API
**Includes**:
- GraphQL schema (`collection.graphqls`)
- Resolvers (call collection service via gRPC)
- Client setup

**Output**: GraphQL API at `/graphql`

**Example Query**:
```graphql
query MyCollections {
  myCollections(page: 1, limit: 10) {
    items {
      id
      name
      status
      contractAddress
      stats { totalItems }
    }
  }
}
```

**[→ See 01-BACKEND-PLAN.md for details](./../01-BACKEND-PLAN.md)**

---

### Phase 4: Upload Proxy
**What**: Add upload endpoints to Gateway for secure metadata uploads
**Output**: `/api/upload/media`, `/api/upload/batch`
**[→ View Guide](./04-UPLOAD-PROXY.md)** ✅ **COMPLETE**

**Key Points**:
- ✅ No new microservice needed
- ✅ Thin proxy (~300 LOC)
- ✅ API key secured server-side
- ✅ Auth required (JWT)

---

### Phase 5: Blockchain Indexer Integration
**What**: Webhook endpoints to receive blockchain events
**Includes**:
- `/webhooks/collection-created`
- `/webhooks/nft-minted`
- HMAC signature verification
- Idempotency handling
- Collection status updates

**Output**: Backend syncs with blockchain via indexer

**Event Flow**:
```
Blockchain → Ponder Indexer → Webhook → Backend API → Database
```

**[→ To be created](./05-BLOCKCHAIN-INDEXER.md)**

---

### Phase 6: Testing & Documentation
**What**: Comprehensive testing suite
**Includes**:
- Unit tests (repositories, services)
- Integration tests (gRPC, GraphQL)
- E2E tests (full flow)
- Load testing (1000 concurrent users)
- API documentation

**Output**: >80% coverage, documented

**[→ To be created](./06-TESTING.md)**

---

## 🚀 Getting Started

### For Developers New to This Codebase

1. **Read Overview**:
   ```bash
   cat ../00-OVERVIEW.md
   cat ../01-BACKEND-PLAN.md
   ```

2. **Start with Phase 1**:
   ```bash
   cat 01-DATABASE-SCHEMA.md
   # Follow step-by-step instructions
   ```

3. **Move to Phase 2**:
   ```bash
   # After Phase 1 complete
   cat 02-COLLECTION-SERVICE.md  # (to be created)
   ```

### For Developers Joining Mid-Project

1. **Check current status** (ask team)
2. **Jump to active phase** (e.g., Phase 4)
3. **Follow checklist** to ensure prerequisites met

---

## ✅ Phase Completion Criteria

### Phase 1: Database ✓
- [ ] Migration applied
- [ ] All tables exist
- [ ] Triggers working
- [ ] Test data loaded

### Phase 2: Collection Service ✓
- [ ] gRPC server running :50054
- [ ] Proto compiled
- [ ] All RPC methods implemented
- [ ] Tests passing

### Phase 3: GraphQL Gateway ✓
- [ ] Schema generated
- [ ] Resolvers working
- [ ] Queries testable in playground
- [ ] Mutations working

### Phase 4: Upload Proxy ✓
- [x] Metadata client implemented
- [x] Upload endpoints working
- [x] Auth enforced
- [x] File validation working

### Phase 5: Blockchain Indexer ✓
- [ ] Webhook endpoints implemented
- [ ] Signature verification working
- [ ] Events processed correctly
- [ ] Idempotency tested

### Phase 6: Testing ✓
- [ ] Unit tests >80% coverage
- [ ] Integration tests passing
- [ ] E2E tests passing
- [ ] Documentation complete

---

## 📊 Estimated Effort Summary

| Phase | Lines of Code | Complexity | Time |
|-------|---------------|------------|------|
| Phase 1 | ~500 (SQL) | Medium | Foundation |
| Phase 2 | ~2000 (Go) | High | Core Work |
| Phase 3 | ~800 (Go+GraphQL) | Medium | Integration |
| Phase 4 | ~300 (Go) | Low | Quick Add |
| Phase 5 | ~500 (Go) | Medium | Event Handling |
| Phase 6 | ~1500 (Tests) | High | QA |
| **Total** | **~5600 LOC** | **-** | **TDD Approach** |

---

## 🔗 Related Documentation

- [../00-OVERVIEW.md](../00-OVERVIEW.md) - Feature overview
- [../01-BACKEND-PLAN.md](../01-BACKEND-PLAN.md) - Backend plan overview
- [../02-FRONTEND-PLAN.md](../02-FRONTEND-PLAN.md) - Frontend plan
- [../03-INTEGRATION.md](../03-INTEGRATION.md) - Integration workflows
- [../04-COMPLETION-CRITERIA.md](../04-COMPLETION-CRITERIA.md) - Acceptance criteria

---

**Last Updated**: 2025-11-20
**Current Phase**: Phase 1 & 4 complete
**Next Steps**: Complete Phase 2 & 3 documentation
