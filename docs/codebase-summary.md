# Codebase Summary

## Repository Overview

**Project**: Zuno NFT Marketplace API
**Repository Size**: 138 tracked files
**Primary Language**: Go 1.21+
**Architecture**: Microservices (gRPC) + GraphQL Gateway BFF
**Infrastructure**: Docker Compose + Kubernetes/Tilt

## Directory Structure

```
zuno-marketplace-api/
├── .claude/                    # Claude Code configuration
│   ├── workflows/             # Development workflows
│   ├── skills/                # Agent skills
│   ├── commands/              # Custom slash commands
│   └── scripts/               # Development scripts
├── .factory/                  # Factory templates for agents
│   ├── commands/              # Agent command templates
│   └── droids/                # Agent configurations
├── db/                        # Database migrations
│   └── migrations/            # SQL migration files
├── docs/                      # Project documentation
├── infra/                     # Infrastructure configurations
│   ├── development/
│   │   ├── build/             # Build scripts
│   │   ├── docker/            # Dockerfiles
│   │   └── k8s/               # Kubernetes manifests
│   └── production/            # Production configs (TODO)
├── proto/                     # gRPC protobuf definitions
├── services/                  # Microservices (4 services)
│   ├── auth-service/          # SIWE authentication
│   ├── user-service/          # User management
│   ├── wallet-service/        # Wallet management
│   └── graphql-gateway/       # GraphQL BFF API
├── shared/                    # Shared packages
│   ├── env/                   # Environment loading
│   └── proto/pb/              # Generated protobuf code
├── CLAUDE.md                  # Claude Code instructions
├── docker-compose.yml         # Local development
├── go.mod                     # Go dependencies
├── Makefile                   # Development commands
├── Tiltfile                   # Kubernetes hot reload
└── README.md                  # Project documentation
```

## Services Breakdown

### 1. Auth Service (Port 50051)

**Location**: `services/auth-service/`
**Responsibility**: SIWE authentication, JWT tokens, session management

**Key Components**:
- `cmd/main.go` - Service entry point
- `internal/config/` - Configuration loading
- `internal/models/` - Data models (AuthNonce, Session, LoginEvent)
- `internal/repository/` - Data access layer (interfaces-based)
- `internal/service/` - Business logic (SIWE, JWT)
- `internal/server/` - gRPC server implementation
- `internal/client/` - gRPC client connections to other services

**Key Files**:
- `internal/models/auth_nonce.go` - Nonce generation/validation
- `internal/models/session.go` - Session tracking
- `internal/models/login_event.go` - Login audit trail
- `internal/service/siwe_service.go` - SIWE verification (includes atomic nonce consumption)
- `internal/service/jwt_service.go` - Token generation/validation
- `internal/repository/nonce_repository.go` - Nonce persistence
- `internal/repository/session_repository.go` - Session persistence

**Test Coverage**: 6+ test files with table-driven tests
- `internal/service/jwt_service_test.go` - Token validation
- `internal/service/siwe_service_test.go` - SIWE verification (incomplete)
- `internal/repository/login_event_repository_test.go` - Event logging

**Database Tables**: 3
- `auth_nonces` - SIWE nonce management
- `sessions` - Session tracking with device fingerprinting
- `login_events` - Login history

**Dependencies**: User Service, Wallet Service (via gRPC)

### 2. User Service (Port 50052)

**Location**: `services/user-service/`
**Responsibility**: User profiles, account management, preferences

**Key Components**:
- `cmd/main.go` - Service entry point
- `internal/config/` - Configuration loading
- `internal/models/` - Data models (User, Profile)
- `internal/repository/` - Data access layer
- `internal/server/` - gRPC server implementation

**Key Files**:
- `internal/models/user.go` - User domain model with UUID v5 generation
- `internal/repository/user_repository.go` - User CRUD + deterministic ID generation

**Database Tables**: 5
- `users` - Core user accounts
- `profiles` - User profile information
- `user_follows` - Social relationships
- `user_preferences` - User settings
- `user_stats` - Activity statistics

**Dependencies**: None (depended on by Auth & Wallet services)

### 3. Wallet Service (Port 50053)

**Location**: `services/wallet-service/`
**Responsibility**: Multi-wallet management, CAIP-2/CAIP-10 support

**Key Components**:
- `cmd/main.go` - Service entry point
- `internal/config/` - Configuration loading
- `internal/models/` - Data models (WalletLink)
- `internal/repository/` - Data access layer
- `internal/server/` - gRPC server implementation

**Key Files**:
- `internal/models/wallet_link.go` - Wallet linking with CAIP-10 support
- `internal/repository/wallet_repository.go` - Wallet CRUD operations

**Database Tables**: 3
- `wallet_links` - Multi-wallet associations
- `wallet_verifications` - Verification tracking
- `wallet_activity` - Activity logging

**Dependencies**: User Service (via gRPC)

### 4. GraphQL Gateway (Port 8081)

**Location**: `services/graphql-gateway/`
**Responsibility**: GraphQL API, HTTP endpoint, BFF layer

**Key Components**:
- `cmd/main.go` - Service entry point
- `internal/config/` - Configuration loading
- `internal/middleware/` - HTTP middleware (auth, logging)
- `internal/health/` - Health check implementation
- `internal/context/` - Request context management
- `internal/cookie/` - Cookie handling (sessions)
- `graph/` - GraphQL schema and resolvers

**GraphQL Schemas**:
- `graph/schemas/auth.graphqls` - Auth queries/mutations
- `graph/schemas/user.graphqls` - User queries/mutations
- `graph/schemas/wallet.graphqls` - Wallet queries/mutations
- `graph/schemas/schema.graphqls` - Root schema

**Key Files**:
- `graph/resolver.go` - GraphQL resolver entry point
- `graph/schema.resolvers.go` - Auto-generated resolver implementations
- `graph/generated.go` - gqlgen generated code
- `internal/middleware/auth.go` - JWT verification middleware
- `internal/health/checker.go` - Health endpoint

**Test Coverage**: 3+ test files
- `internal/middleware/auth_test.go` - Auth middleware
- `internal/health/checker_test.go` - Health check

**Dependencies**: All 3 gRPC services (Auth, User, Wallet)

## Shared Packages

### `shared/env/` - Environment Loading

**File**: `shared/env/env.go`
**Purpose**: Type-safe environment variable loading
**Features**:
- `GetString()` - Load string variables
- `GetInt()` - Load integer variables
- `GetBool()` - Load boolean variables
- Required vs optional variables
- Sensible defaults

### `shared/proto/pb/` - Generated Protobuf Code

**Generated Files** (6 files):
- `auth.pb.go` - Auth message definitions
- `auth_grpc.pb.go` - Auth service interface
- `user.pb.go` - User message definitions
- `user_grpc.pb.go` - User service interface
- `wallet.pb.go` - Wallet message definitions
- `wallet_grpc.pb.go` - Wallet service interface

**Source**: `proto/` directory proto definitions

## Protocol Buffers

**Location**: `proto/` directory

### `proto/auth.proto` (5 RPC methods)
```protobuf
service Auth {
  rpc GetNonce(GetNonceRequest) returns (GetNonceResponse)
  rpc VerifySIWE(VerifySIWERequest) returns (VerifySIWEResponse)
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse)
  rpc GetSession(GetSessionRequest) returns (GetSessionResponse)
  rpc LogEvent(LogEventRequest) returns (LogEventResponse)
}
```

### `proto/user.proto` (3 RPC methods)
```protobuf
service User {
  rpc GetUser(GetUserRequest) returns (GetUserResponse)
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse)
  rpc UpdateProfile(UpdateProfileRequest) returns (UpdateProfileResponse)
}
```

### `proto/wallet.proto` (2 RPC methods)
```protobuf
service Wallet {
  rpc LinkWallet(LinkWalletRequest) returns (LinkWalletResponse)
  rpc GetWallets(GetWalletsRequest) returns (GetWalletsResponse)
}
```

## Infrastructure & Deployment

### Development Setup

**Docker Compose** (`docker-compose.yml`)
- PostgreSQL 15 (port 5432)
- Redis 7 (port 6379)
- RabbitMQ 3 (ports 5672, 15672)
- 5 service containers (auth, user, wallet, gateway, postgres)

### Kubernetes Deployment

**Development K8s Manifests** (`infra/development/k8s/`)
- `postgres.yaml` - PostgreSQL StatefulSet
- `redis.yaml` - Redis Deployment
- `rabbitmq.yaml` - RabbitMQ Deployment
- `auth-service-deployment.yaml` - Auth Service
- `user-service-deployment.yaml` - User Service
- `wallet-service-deployment.yaml` - Wallet Service
- `graphql-gateway-deployment.yaml` - GraphQL Gateway
- `app-config.yaml` - ConfigMap
- `secrets.yaml.example` - Secret template

**Features**:
- Tilt integration for hot reload
- Service discovery via DNS
- Resource limits specified
- Health checks configured
- No ingress (TODO for production)

### Dockerfiles

**Development Dockerfiles** (`infra/development/docker/`)
- `auth-service.Dockerfile` - Multi-stage Alpine build
- `user-service.Dockerfile` - Multi-stage Alpine build
- `wallet-service.Dockerfile` - Multi-stage Alpine build
- `graphql-gateway.Dockerfile` - Multi-stage Alpine build
- `postgres.dockerfile` - Custom PostgreSQL image

**Pattern**: Multi-stage builds with Alpine base (< 100MB images)

### Database Migrations

**Location**: `db/migrations/`

**Migration Files**:
- `000001_init_schema.up.sql` - Initial schema with all 11 tables
- `000001_init_schema.down.sql` - Schema teardown
- `000002_support_caip10_account_id.up.sql` - CAIP-10 support
- `000002_support_caip10_account_id.down.sql` - Rollback CAIP-10

**Management**: DB migration tool (likely migrate)

## Code Organization Pattern

### Clean Architecture Layers

Each service follows consistent layering:

```
main.go (Entry point)
  ↓
config/ (Configuration loading)
  ↓
models/ (Domain models)
  ↓
repository/ (Data access - interfaces)
  ↓
service/ (Business logic)
  ↓
server/ (gRPC/HTTP handlers)
  ↓
Database
```

### Repository Pattern

All repositories implement interfaces for testing:

```go
type UserRepository interface {
  GetUser(ctx context.Context, id string) (*User, error)
  CreateUser(ctx context.Context, user *User) error
  UpdateUser(ctx context.Context, user *User) error
}
```

### Dependency Injection

Explicit dependency passing in constructors:

```go
type UserService struct {
  repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
  return &UserService{repo: repo}
}
```

## Testing Strategy

### Test Coverage

**Existing Test Files**: 6+
- JWT service tests (token validation)
- SIWE service tests (signature verification - incomplete)
- Auth middleware tests (token extraction)
- Health checker tests (endpoint validation)
- Login event repository tests
- E2E auth flow tests

**Coverage Requirement**: 80% minimum on new code (enforced in CI/CD)

### Testing Patterns

**Table-Driven Tests**: Standard approach
```go
tests := []struct {
  name    string
  input   interface{}
  want    interface{}
  wantErr bool
}{
  // test cases
}
```

**In-Memory Database**: testcontainers for integration tests

## CI/CD Pipeline

**Location**: `.github/workflows/` (not shown in repomix)

**GitHub Actions Workflows**:
- Lint (golangci-lint)
- Test (go test with coverage)
- Build (Docker image builds)
- Security Scan (gosec)
- PR Validation (coverage threshold 80%)

## Development Tools

### Makefile

**Common Commands** (`Makefile`)
- `make dev` - Start Docker Compose environment
- `make test` - Run tests
- `make build` - Build all services
- `make proto` - Generate protobuf code
- `make lint` - Run linter
- `make format` - Format code
- `make ci` - Run CI pipeline locally

### Tilt

**File**: `Tiltfile`
**Features**:
- Hot reload on code changes
- Live logs and resource monitoring
- Kubernetes integration
- UI at http://localhost:10350

### Build Scripts

**Location**: `infra/development/build/`
- `.bat` files for Windows builds
- Service-specific build logic

## File Statistics

**Total Files**: 138 tracked
**Go Source Files**: 39
- Auth Service: 8 files
- User Service: 5 files
- Wallet Service: 5 files
- GraphQL Gateway: 8 files
- Shared: 2 files
- Generated: 6 files

**Configuration Files**: 20+
- Kubernetes manifests
- Dockerfiles
- Docker Compose

**Documentation**: 10+ files
- README.md
- QUICKSTART.md
- TILT.md
- Service READMEs
- This codebase summary

**Database**: 4 files
- Migration files
- DB README

## Key Dependencies

### Go Packages (from go.mod)
- gRPC/protobuf libraries
- PostgreSQL driver
- JWT library
- GraphQL gqlgen
- Standard library (context, encoding, net, etc.)

### Runtime Dependencies
- PostgreSQL 15
- Redis 7
- RabbitMQ 3
- Go 1.21+

## Code Statistics

**Approximate Lines of Code**:
- Auth Service: ~800 lines
- User Service: ~400 lines
- Wallet Service: ~400 lines
- GraphQL Gateway: ~600 lines
- Shared: ~100 lines
- Generated Code: ~5000 lines (protobuf)
- Tests: ~600 lines
- **Total**: ~8000 lines (excluding generated code)

## Architecture Patterns Used

1. **Repository Pattern** - Data access abstraction
2. **Dependency Injection** - Explicit dependencies
3. **Interface-Based Design** - Testable components
4. **gRPC for IPC** - Service-to-service communication
5. **GraphQL BFF** - Client-facing API aggregation
6. **Table-Driven Tests** - Maintainable test suites
7. **Atomic Operations** - Database-level consistency

## Current Implementation Status

### Implemented
- Service scaffolding and configuration
- Database schema and migrations
- Proto definitions
- Repository interfaces and implementations
- Basic service structure
- GraphQL schema definitions
- Health check endpoints
- Docker and Kubernetes setup

### Partial
- SIWE verification (has TODOs)
- JWT token service
- Session management
- GraphQL resolvers (basic structure)

### TODO
- Complete SIWE implementation
- Full JWT middleware
- RabbitMQ integration
- Redis caching
- Complete E2E tests
- Production Kubernetes manifests
- Monitoring and observability

## Unresolved Items

1. RabbitMQ event handling implementation approach
2. Redis caching strategy and TTL values
3. GraphQL subscription (WebSocket) support needed?
4. Production ingress and TLS configuration
5. Monitoring stack (Prometheus/Jaeger) integration
6. Backup and disaster recovery procedures

---

**Generated**: 2025-12-04
**Last Updated**: 2025-12-04
**Source**: repomix output analysis
