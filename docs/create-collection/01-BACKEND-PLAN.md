# BACKEND IMPLEMENTATION PLAN - OVERVIEW

**Repository**: `zuno-marketplace-api`
**Architecture**: Go microservices (gRPC + GraphQL BFF)
**New Service**: `collection-service` (gRPC :50054)

---

## 🎯 Overview

Backend API nhiệm vụ chính:
1. ✅ **Collection lifecycle management** (PENDING → DEPLOYED → FAILED)
2. ✅ **User-collection relationships** (ownership, favorites)
3. ✅ **Blockchain integration** (index events, sync on-chain data)
4. ✅ **Stats aggregation** (floor price, volume, owners)
5. ✅ **Search & discovery** (complex queries, filtering)
6. ✅ **Access control** (permissions, authorization)
7. ✅ **Activity tracking** (audit logs)
8. ✅ **Upload proxy** (secure metadata service integration)

---

## 🏗️ Architecture Additions

### New Components

```
services/
├── auth-service/          :50051 (existing)
├── user-service/          :50052 (existing)
├── wallet-service/        :50053 (existing)
├── collection-service/    :50054 ✨ NEW - Collection business logic
└── graphql-gateway/       :8081  (updated)
    ├── /graphql                  (existing)
    ├── /playground               (existing)
    └── /api/upload/              ✨ NEW - Upload proxy endpoints
        ├── /media
        └── /batch
```

### Integration Points

```
┌────────────────────────────────────────────────┐
│          GraphQL Gateway (:8081)               │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  GraphQL API + Upload Proxy                   │
└──────┬─────────────────────────┬───────────────┘
       │                         │
       │ gRPC                    │ HTTP (proxy)
       ▼                         ▼
┌──────────────────┐    ┌─────────────────────┐
│ Collection       │    │ Metadata Service    │
│ Service          │    │ (External)          │
│ :50054           │    │ :3001               │
└──────────────────┘    └─────────────────────┘
       │
       │ Events
       ▼
┌──────────────────┐
│ Blockchain       │
│ Indexer          │
│ (Ponder)         │
└──────────────────┘
```

---

## 📂 Implementation Phases (Detailed Docs)

Backend implementation chia thành **6 phases** chi tiết:

### Phase 1: Database Schema Design
**File**: [`backend/01-DATABASE-SCHEMA.md`](./backend/01-DATABASE-SCHEMA.md)
- Collections tables (main, metadata, items, stats, activity, allowlist)
- Triggers & functions
- Migrations
- Testing schema

**Output**: Migrations ready, schema tested

---

### Phase 2: Collection Service (gRPC)
**File**: [`backend/02-COLLECTION-SERVICE.md`](./backend/02-COLLECTION-SERVICE.md)
- Proto definitions
- Repository layer (GORM)
- Service layer (business logic)
- gRPC server implementation
- Docker integration

**Output**: Collection service running on :50054

---

### Phase 3: GraphQL Gateway Integration
**File**: [`backend/03-GRAPHQL-GATEWAY.md`](./backend/03-GRAPHQL-GATEWAY.md)
- GraphQL schema (collection.graphqls)
- Resolvers implementation
- gRPC client setup
- Query/Mutation handlers

**Output**: GraphQL API accessible at /graphql

---

### Phase 4: Upload Proxy (Metadata Service Integration)
**File**: [`backend/04-UPLOAD-PROXY.md`](./backend/04-UPLOAD-PROXY.md)
- **NEW**: Upload proxy endpoints in GraphQL Gateway
- Metadata Service HTTP client
- File upload handling (multipart/form-data)
- Security (API key management)
- Progress tracking

**Output**: Upload endpoints at /api/upload/*

**Key Decision**:
- ✅ Backend proxies uploads to Metadata Service
- ✅ API key secured server-side (not exposed to frontend)
- ✅ Thin proxy layer (~100 lines code)
- ✅ No new microservice needed

---

### Phase 5: Blockchain Indexer Integration
**File**: [`backend/05-BLOCKCHAIN-INDEXER.md`](./backend/05-BLOCKCHAIN-INDEXER.md)
- Webhook endpoints for indexer
- Event payload validation
- HMAC signature verification
- Idempotency handling
- Collection status updates

**Output**: Webhook endpoints receiving blockchain events

---

### Phase 6: Testing & Documentation
**File**: [`backend/06-TESTING.md`](./backend/06-TESTING.md)
- Unit tests (repositories, services)
- Integration tests (gRPC, GraphQL)
- E2E tests (full flow)
- Load testing
- API documentation

**Output**: >80% coverage, documented APIs

---

## 🔄 Development Sequence

### Sequential Dependencies

```
Phase 1 (Database)
  ↓ MUST complete first
Phase 2 (Collection Service)
  ↓ MUST complete
Phase 3 (GraphQL Gateway)
  ↓ Can parallelize below
  ├─→ Phase 4 (Upload Proxy)
  └─→ Phase 5 (Blockchain Indexer)
       ↓ All complete
     Phase 6 (Testing)
```

### Parallel Work Opportunities

After Phase 3 completes:
- ✅ Backend dev A: Work on Phase 4 (Upload Proxy)
- ✅ Backend dev B: Work on Phase 5 (Blockchain Indexer)
- ✅ Frontend team: Can start integration (Phase 1-3 complete)

---

## 🎯 Phase Completion Checklist

### Phase 1: Database ✅ COMPLETED
- [x] Migration files created (up + down)
- [x] All tables, indexes, constraints working
- [x] Triggers tested
- [x] Migration 000003_add_collections_tables deployed
- [x] Schema tested with collection service

### Phase 2: Collection Service ✅ COMPLETED
- [x] Proto compiled
- [x] Repository layer implemented
- [x] Service layer implemented
- [x] gRPC server implemented
- [x] Health check working
- [x] Unit tests added (converter, validator, service)
- [x] Service running on :50054

### Phase 3: GraphQL Gateway ✅ COMPLETED
- [x] Collection GraphQL schema created
- [x] Resolvers implemented
- [x] gRPC client connected
- [x] Queries working
- [x] Mutations working

### Phase 4: Upload Proxy ✅ COMPLETED
- [x] HTTP client for Metadata Service working
- [x] Upload endpoints implemented
- [x] Auth middleware applied
- [x] File validation working
- [x] Error handling tested
- [x] Unit tests for client (76.9% coverage)
- [x] Unit tests for handlers (69.9% coverage)

### Phase 5: Blockchain Indexer 📋 NOT STARTED
- [ ] Webhook endpoints implemented
- [ ] Signature verification working
- [ ] Idempotency tested
- [ ] Database updates working
- [ ] Indexer connected

### Phase 6: Testing 🔄 ONGOING
- [x] Unit tests for auth-service (jwt, siwe, login events)
- [x] Unit tests for wallet-service (repository 89.3%, server 97.5%)
- [x] Unit tests for user-service (repository 83.0%, server 88.7%)
- [x] Unit tests for graphql-gateway (context 100%, cookie 100%, helpers)
- [x] Unit tests for collection-service (converter, validator, service)
- [ ] Integration tests for collection service
- [ ] E2E tests for create collection flow
- [ ] Load tests
- [ ] API documentation complete

---

## 📊 Estimated Effort

| Phase | Complexity | Lines of Code | Time Estimate |
|-------|------------|---------------|---------------|
| Phase 1: Database | Medium | ~500 (SQL) | Foundation |
| Phase 2: Collection Service | High | ~2000 (Go) | Core work |
| Phase 3: GraphQL Gateway | Medium | ~800 (Go + GraphQL) | Integration |
| Phase 4: Upload Proxy | Low | ~300 (Go) | Quick add |
| Phase 5: Blockchain Indexer | Medium | ~500 (Go) | Event handling |
| Phase 6: Testing | High | ~1500 (Go tests) | Quality assurance |
| **Total** | **-** | **~5600 LOC** | **Follow TDD** |

---

## 🚀 Quick Start

### For Developers Starting Phase 1
```bash
cd docs/create-collection/backend
cat 01-DATABASE-SCHEMA.md

# Follow step-by-step instructions
make migrate-create NAME=add_collections_tables
# ... implement schema
make migrate-up
make test
```

### For Developers Starting Phase 2
```bash
cat backend/02-COLLECTION-SERVICE.md

# Create service structure
mkdir -p services/collection-service/{cmd,internal/{models,repository,service,server}}
# ... follow guide
```

### For Developers Starting Phase 3
```bash
cat backend/03-GRAPHQL-GATEWAY.md

# Add GraphQL schema
vim services/graphql-gateway/graph/schemas/collection.graphqls
# ... follow guide
```

---

## 🔗 Related Documentation

- [00-OVERVIEW.md](./00-OVERVIEW.md) - Feature overview & architecture
- [02-FRONTEND-PLAN.md](./02-FRONTEND-PLAN.md) - Frontend implementation
- [03-INTEGRATION.md](./03-INTEGRATION.md) - Integration workflows
- [04-COMPLETION-CRITERIA.md](./04-COMPLETION-CRITERIA.md) - Acceptance criteria

---

## 📝 Key Architectural Decisions

### Decision 1: Upload Proxy in Gateway (Not New Service)
**Context**: Need secure way to upload to Metadata Service
**Decision**: Add upload endpoints to GraphQL Gateway
**Rationale**:
- ✅ Gateway already HTTP server
- ✅ Thin proxy layer (~100 LOC)
- ✅ API key secured server-side
- ✅ No new service complexity

**Alternative Rejected**: Create dedicated Upload Service (over-engineering)

### Decision 2: Collection Service Naming
**Context**: Need name for new microservice
**Decision**: `collection-service`
**Rationale**:
- ✅ Consistent with existing pattern (auth-service, user-service)
- ✅ Clear responsibility
- ✅ Easy to understand

**Alternative Rejected**: nft-service, marketplace-service (too broad)

### Decision 3: Database Schema Organization
**Context**: How to structure collection data
**Decision**: Separate tables (collections, metadata, stats, items, activity, allowlist)
**Rationale**:
- ✅ Normalized schema
- ✅ Efficient queries
- ✅ Easy to extend
- ✅ Clear relationships

**Alternative Rejected**: JSONB blob (hard to query/aggregate)

---

## 📈 Current Progress Summary

**Completed Phases**:
- ✅ Phase 1: Database Schema Design (100%)
- ✅ Phase 2: Collection Service (100%)
- ✅ Phase 3: GraphQL Gateway Integration (100%)
- ✅ Phase 4: Upload Proxy (100%)

**In Progress**:
- 🔄 Phase 6: Testing (65% - unit tests complete, integration tests pending)

**Not Started**:
- 📋 Phase 5: Blockchain Indexer

**Overall Progress**: ~70% (4/6 phases complete)

---

**Last Updated**: 2025-11-22
**Status**: Phase 4 Complete → Phase 5 Ready to Start
**Recent Work**:
- Completed Phase 3: GraphQL Gateway Integration with collection resolvers
- Completed Phase 4: Upload Proxy with Metadata Service integration
- Added upload endpoints: /api/upload/media and /api/upload/batch
- Implemented MetadataServiceClient with 76.9% test coverage
- Implemented UploadHandler with 69.9% test coverage
- Auth middleware integration for secure uploads
**Next Step**: Begin Phase 5 (Blockchain Indexer Integration) - Implement webhook endpoints for indexer events
