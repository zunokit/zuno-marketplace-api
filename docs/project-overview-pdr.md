# Project Overview & Product Development Requirements

## Project Vision

Zuno NFT Marketplace API is a production-ready microservices backend designed for Web3 wallet authentication, user management, and multi-wallet support. The system enables secure, deterministic user identity across blockchain networks.

## Current Status

- **Version**: 0.1.0 (Clean Skeleton)
- **Release Date**: November 15, 2025
- **Status**: Ready for feature development
- **Branch**: `feature/hybird-serverless-and-servers-infrastructure`
- **Phase 3 Complete**: Application Configuration (INFRA_MODE support, URL-based config, comprehensive tests)

## Key Objectives

1. **Secure Authentication**: SIWE (Sign-In with Ethereum) EIP-4361 support with replay attack detection
2. **Multi-Wallet Support**: CAIP-10 account management with primary wallet designation
3. **Deterministic Identity**: UUID v5 user IDs from account addresses for idempotency
4. **Scalable Architecture**: gRPC microservices with GraphQL BFF API
5. **Production Ready**: 80% test coverage, comprehensive CI/CD pipeline

## Core Features

### Authentication Service (Port 50051)
- SIWE message signing verification
- JWT token generation (access + refresh)
- Refresh token rotation for replay detection
- Atomic nonce consumption
- Session tracking with device fingerprints
- Login event history

### User Service (Port 50052)
- Deterministic user profile creation
- User preference management
- Social follow system
- User statistics tracking
- Account linking

### Wallet Service (Port 50053)
- Multi-wallet management per user
- CAIP-10 standard compliance
- Primary wallet designation
- Wallet verification tracking
- Activity logging

### GraphQL Gateway (Port 8081)
- HTTP/GraphQL BFF API
- Cross-service GraphQL composition
- Authentication middleware integration
- Health check endpoints
- WebSocket subscription support (future)

## Technology Stack

### Backend & Runtime
- **Language**: Go 1.21+
- **gRPC**: Service-to-service communication
- **GraphQL**: Client-facing API (gqlgen)

### Databases & Cache
- **PostgreSQL 15**: Primary data store (ACID compliance)
- **Redis 7**: Session/cache layer (not yet utilized)
- **RabbitMQ 3**: Event messaging infrastructure (not yet integrated)

### Infrastructure
- **Containerization**: Docker with multi-stage Alpine builds
- **Orchestration**: Kubernetes + Tilt for development
- **Development**: Docker Compose for local setup

### Authentication
- **EIP-4361**: SIWE message signing
- **JWT**: HMAC-SHA256 tokens
- **Standards**: CAIP-10 account identifiers

### CI/CD & Quality
- **Automation**: GitHub Actions (lint, test, build, security)
- **Coverage**: 80% minimum threshold
- **Linting**: golangci-lint
- **Security**: gosec, go vet
- **Formatting**: gofmt

## Database Schema

### Auth Service (3 tables)
- **auth_nonces**: SIWE nonce generation and consumption
- **sessions**: User session tracking with device fingerprinting
- **login_events**: Login attempt history for audit trails

### User Service (5 tables)
- **users**: Core user accounts with deterministic UUID v5
- **profiles**: User profile information
- **user_follows**: Social follow relationships
- **user_preferences**: User settings and preferences
- **user_stats**: User activity statistics

### Wallet Service (3 tables)
- **wallet_links**: Multi-wallet associations
- **wallet_verifications**: SIWE verification status
- **wallet_activity**: Wallet transaction history

**Total**: 11 tables with proper indexing and constraints

## Design Decisions

### 1. Deterministic User Identity
**Decision**: UUID v5 generated from Ethereum account address
**Rationale**: Ensures same address always produces same user ID; enables idempotent operations
**Implementation**: SHA-1 namespace-based generation in user repository

### 2. Refresh Token Rotation
**Decision**: New refresh token issued on each access token refresh
**Rationale**: Prevents token family replay attacks; revokes compromised sessions
**Implementation**: Token family tracking with generation counter

### 3. Atomic Nonce Consumption
**Decision**: Database function for atomic read-consume operation
**Rationale**: Prevents race conditions in SIWE verification
**Implementation**: PostgreSQL transaction isolation level serializable

### 4. Idempotent Operations
**Decision**: EnsureUser, UpsertLink patterns
**Rationale**: Handles duplicate requests from network retries
**Implementation**: UPSERT with ON CONFLICT clauses

### 5. Fire-and-Forget Audit Logging
**Decision**: Async login event recording
**Rationale**: Prevents blocking authentication on logging failures
**Implementation**: RabbitMQ integration (future); currently synchronous

## Architecture Pattern

```
Frontend (Web3)
    ↓
GraphQL Gateway (HTTP/WS)
    ↓
gRPC Services (Auth, User, Wallet)
    ↓
PostgreSQL (Primary Data) + Redis (Cache) + RabbitMQ (Events)
```

**Layer Separation**:
- **Transport**: gRPC (service-to-service), HTTP/GraphQL (client-facing)
- **Service**: Business logic, validation, coordination
- **Repository**: Data access, query building, database interaction
- **Model**: Data representation, serialization

## Dependency Graph

```
GraphQL Gateway
├── Auth Service (gRPC)
│   └── User Service (internal dependency)
│   └── Wallet Service (internal dependency)
├── User Service (gRPC)
└── Wallet Service (gRPC)
    └── User Service (internal dependency)
```

## Development Roadmap

### Phase 1: Foundation (Complete - v0.1.0)
- ✅ Clean skeleton with tested schemas
- ✅ gRPC service definitions
- ✅ Database infrastructure
- ✅ CI/CD pipeline
- ✅ Development tooling (Docker Compose, Tilt, Makefile)

### Phase 2: Environment Configuration (Complete - v0.1.0)
- ✅ Environment templates (.env.development.example, .env.production.example)
- ✅ Interactive setup script (scripts/setup-env.sh)
- ✅ Serverless infrastructure support (Supabase, Upstash, CloudAMQP)
- ✅ Docker infrastructure support (local containers)
- ✅ Infrastructure mode detection and switching
- ✅ .gitignore updates for environment files

### Phase 3: Application Configuration (Complete - v0.1.0)
- ✅ INFRA_MODE environment variable support in all services
- ✅ URL-based configuration for Database, Redis, RabbitMQ
- ✅ GetDSN(), GetAddr(), GetURL() methods for connection strings
- ✅ Comprehensive config tests for all services (Docker + Serverless modes)
- ✅ Mode-aware configuration loading pattern implemented

### Phase 4: Documentation & Scripts (v0.1.1 - Next)
- [ ] Update all documentation with Phase 3 changes
- [ ] Add deployment guide for both modes
- [ ] Create troubleshooting guide
- [ ] Update README with Phase 3 details
- [ ] Add health check scripts for serverless mode

### Phase 5: Testing & Validation (v0.1.1)
- [ ] Integration tests for both modes
- [ ] CI/CD pipeline updates for dual-mode testing
- [ ] Performance comparison between modes

### Phase 6: Core Features (v0.2.0)
- [ ] Complete SIWE authentication implementation
- [ ] JWT token service
- [ ] Session management
- [ ] User profile CRUD
- [ ] Wallet linking and verification
- [ ] GraphQL schema integration

### Phase 7: Advanced Features (v0.3.0)
- [ ] Redis session caching
- [ ] RabbitMQ event integration
- [ ] Social features (follow system)
- [ ] User statistics aggregation
- [ ] WebSocket subscriptions

### Phase 8: Production Readiness (v0.4.0)
- [ ] Production Kubernetes manifests
- [ ] Monitoring and observability (Prometheus, Jaeger)
- [ ] Rate limiting and throttling
- [ ] API versioning strategy
- [ ] Performance optimization
- [ ] Security hardening

### Phase 9: Enterprise Features (v0.5.0+)
- [ ] Multi-chain support expansion
- [ ] Advanced analytics
- [ ] Admin dashboard
- [ ] Rate limiting by tier
- [ ] Audit logging system

## Success Criteria

### Code Quality
- Minimum 80% test coverage on new code
- Zero critical security issues
- Zero golangci-lint warnings
- All tests passing in CI/CD

### Performance
- API response time < 100ms (p95)
- gRPC latency < 50ms (p95)
- Database query time < 20ms (p95)

### Reliability
- 99.9% uptime in production
- All critical endpoints health-checked
- Graceful degradation on service failures

### Security
- SIWE signature verification mandatory
- JWT token expiration enforced
- SQL injection prevention via prepared statements
- No plaintext secrets in repositories

### Developer Experience
- Local development setup in < 5 minutes
- Hot reload during development
- Comprehensive error messages
- Clear documentation

## Current Gaps & TODOs

### Implementation Gaps
- SIWE test coverage incomplete (placeholders present)
- RabbitMQ integration not implemented
- Redis utilization pending
- Production Kubernetes manifests missing
- Monitoring/observability stack absent
- Serverless health check scripts needed

### Documentation Gaps
- DEVELOPMENT.md referenced but not present
- GraphQL schema documentation incomplete
- Architecture Decision Records (ADRs) missing
- API endpoint examples needed
- Environment setup documentation complete ✅

### Infrastructure Gaps
- No persistent storage volumes
- No ingress controller configuration
- No pod autoscaling policies
- No persistent backup strategy
- No centralized logging

## Unresolved Questions

1. Should refresh token rotation happen at every refresh or only on suspicious activity?
2. How long should session tokens remain valid (current: 15min access, 7d refresh)?
3. Should wallet verification be mandatory or optional?
4. What is the strategy for handling wallet address changes?
5. Should RabbitMQ events be synchronous or fully asynchronous?
6. How should cross-service errors be handled in GraphQL gateway?

---

**Version**: 0.1.0
**Last Updated**: 2025-12-29
**Maintainer**: Zuno Development Team
**Phase 3 Complete**: Application Configuration (2025-12-29)
