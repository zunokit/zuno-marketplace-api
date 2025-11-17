# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a clean, production-ready NFT Marketplace backend built with Go microservices. The architecture uses gRPC for inter-service communication and GraphQL as the BFF (Backend for Frontend) API. The primary authentication mechanism is SIWE (Sign-In With Ethereum) using CAIP-10 account identifiers.

**Current Status**: Clean skeleton (v0.1.0) ready for TDD feature development. All database schemas tested and working. Infrastructure services operational.

## Development Commands

### Essential Commands

```bash
# First-time setup
make db-reset                    # Setup fresh database with all migrations
make dev                         # Start all infrastructure services

# Daily development
make dev                         # Start services (preserves data)
make dev-stop                    # Stop services
make dev-logs                    # View logs
make dev-clean                   # Stop + remove volumes (nuclear option)

# Testing
make test                        # Run all tests
make test-service SERVICE=auth   # Test specific service (auth, user, wallet)
make test-coverage               # Generate coverage report (coverage.html)
make test-verbose                # Run tests with verbose output

# Building
make build                       # Build all services to ./build/
make build-service SERVICE=auth  # Build specific service

# Code quality
make lint                        # Run golangci-lint
make format                      # Format code with gofmt + goimports
make ci                          # Full CI pipeline (lint + test + build)

# Database migrations
make migrate-up                  # Apply pending migrations
make migrate-down                # Rollback last migration
make migrate-status              # Show current migration version
make migrate-create NAME=add_feature  # Create new migration files

# Protobuf
make proto                       # Generate Go code from .proto files
```

### Service Ports

- **GraphQL Gateway**: http://localhost:8081/graphql (playground at /playground)
- **Auth Service**: gRPC :50051
- **User Service**: gRPC :50052
- **Wallet Service**: gRPC :50053
- **PostgreSQL**: 5432
- **Redis**: 6379
- **RabbitMQ**: 5672 (management UI: http://localhost:15672)

## Architecture & Key Patterns

### Service Communication Flow

```
Frontend → GraphQL Gateway (HTTP/WS :8081)
              ↓
          gRPC Services (50051-50053)
              ↓
          PostgreSQL + Redis
              ↓
          RabbitMQ (events)
```

### SIWE Authentication Flow

This codebase implements wallet-based authentication using SIWE with CAIP-10 identifiers:

1. **GetNonce**: Client requests nonce for `accountId` (CAIP-10: `eip155:1:0xabc...`)
   - Auth service generates unique nonce stored in `auth_nonces` table
   - Nonce expires in 10 minutes

2. **VerifySiwe**: Client signs SIWE message and sends signature
   - Auth service verifies signature using `SIWEService` (services/auth-service/internal/service/siwe_service.go)
   - Creates/retrieves user via User Service gRPC call
   - Links wallet via Wallet Service gRPC call
   - Generates JWT access + refresh tokens
   - Creates session in `sessions` table with token rotation support

3. **RefreshSession**: Client exchanges refresh token for new tokens
   - Validates refresh token hash
   - Implements token family rotation for security
   - Detects token reuse attacks

### CAIP-10 Account Identifiers

The system uses CAIP-10 format (`eip155:chainId:0xaddress`) throughout:
- Database constraints validate format in `auth_nonces`, `login_events`
- Legacy `0xaddress` format also supported for backwards compatibility
- Chain ID format: `eip155:1` for Ethereum mainnet, `eip155:137` for Polygon, etc.
- See db/migrations/000002_support_caip10_account_id.* for migration

### Database Schema Organization

All schemas are in a single monolithic migration file: `db/migrations/000001_init_schema.up.sql`

**Auth Service Tables**:
- `auth_nonces`: One-time nonces for SIWE (10-minute expiration)
- `sessions`: JWT session management with token rotation
- `login_events`: Audit log of authentication attempts

**User Service Tables**:
- `users`: Core user accounts (UUID-based)
- `profiles`: User profiles with username, bio, social links
- `user_preferences`: Settings (theme, notifications, privacy)
- `user_stats`: Aggregated counts (followers, items, volume)
- `user_follows`: Social graph relationships

**Wallet Service Tables**:
- `wallet_links`: User-wallet associations (supports multiple wallets per user)
- `wallet_activity`: Audit log of wallet operations
- `wallet_verifications`: Wallet ownership verification records

**Key Database Features**:
- Automatic user record creation via `create_user_defaults` trigger
- Primary wallet enforcement via `ensure_single_primary` trigger
- Follow count updates via `update_follow_stats` trigger
- Helper functions: `cleanup_expired_nonces()`, `cleanup_expired_sessions()`, `try_use_nonce()`

### gRPC Service Dependencies

**Auth Service depends on**:
- User Service (CreateUser, GetUser)
- Wallet Service (LinkWallet)

**User Service**: Standalone, no dependencies

**Wallet Service**: Standalone, no dependencies

**GraphQL Gateway depends on**: All three services via gRPC clients

When modifying proto definitions:
1. Edit proto files in `proto/` directory
2. Run `make proto` to regenerate Go code
3. Update service implementations in `services/*/internal/server/`
4. Update GraphQL resolvers in `services/graphql-gateway/graph/schema.resolvers.go`

### JWT Token Management

Location: `services/auth-service/internal/service/jwt_service.go`

- Access tokens: 1 hour expiration (configurable)
- Refresh tokens: 7 days expiration (configurable)
- Token rotation: Each refresh creates new token family generation
- Claims include: `user_id`, `iat`, `exp`
- Secrets loaded from environment: `JWT_SECRET`, `REFRESH_SECRET`

### GraphQL Gateway Middleware

Location: `services/graphql-gateway/internal/middleware/auth.go`

- Extracts JWT from `Authorization: Bearer <token>` header
- Validates token using `JWT_ACCESS_SECRET` env var
- Injects `user_id` into request context
- Public endpoints (getNonce, verifySiwe) bypass auth check
- Protected endpoints automatically get `user_id` from context

### Testing Patterns

The codebase follows TDD:

1. **Unit Tests**: Test individual functions with mocks
   - Example: `services/auth-service/internal/service/jwt_service_test.go`
   - Example: `services/auth-service/internal/service/siwe_service_test.go`

2. **Integration Tests**: Test with real database using testcontainers
   - Example: `tests/e2e/auth/auth_flow_test.go`

3. **Test Database Setup**: Use `make db-reset` to get clean state

When writing tests:
- Use table-driven tests for multiple scenarios
- Mock gRPC clients for unit tests
- Use real Postgres for integration tests
- Minimum 80% coverage for new code

## Project Structure

```
services/
├── auth-service/          # SIWE authentication & JWT management
│   ├── cmd/main.go       # Service entrypoint
│   ├── internal/
│   │   ├── client/       # gRPC clients (user, wallet)
│   │   ├── models/       # Database models
│   │   ├── repository/   # Database access layer
│   │   ├── server/       # gRPC server implementation
│   │   └── service/      # Business logic (SIWE, JWT)
│
├── user-service/         # User profiles & management
│   └── internal/
│       ├── models/
│       ├── repository/
│       └── server/
│
├── wallet-service/       # Wallet linking & verification
│   └── internal/
│       ├── models/
│       ├── repository/
│       └── server/
│
└── graphql-gateway/      # GraphQL BFF API
    ├── cmd/main.go       # Gateway entrypoint
    ├── graph/            # GraphQL schema & resolvers
    ├── internal/
    │   ├── context/      # Request context helpers
    │   ├── cookie/       # Cookie utilities
    │   └── middleware/   # Auth middleware
    └── schema.graphqls   # GraphQL type definitions

shared/
└── proto/pb/            # Generated protobuf code

proto/                   # Protobuf definitions
├── auth.proto          # Auth service interface
├── user.proto          # User service interface
└── wallet.proto        # Wallet service interface

db/migrations/          # Database migrations (golang-migrate format)
```

## Common Development Patterns

### Adding a New gRPC Method

1. Define method in `proto/*.proto`
2. Run `make proto` to regenerate code
3. Implement in `services/*/internal/server/*_server.go`
4. Add GraphQL mutation/query in `services/graphql-gateway/schema.graphqls`
5. Generate GraphQL code: `cd services/graphql-gateway && go run github.com/99designs/gqlgen generate`
6. Implement resolver in `services/graphql-gateway/graph/schema.resolvers.go`
7. Write tests

### Adding a Database Migration

```bash
make migrate-create NAME=add_new_feature
# Edit generated files in db/migrations/
make migrate-up
make migrate-status  # Verify
```

### Environment Variables

Copy `.env.example` to `.env` and modify:
- **Required for auth**: `JWT_SECRET`, `REFRESH_SECRET` (min 32 chars)
- **Database**: `POSTGRES_*` variables
- **Services**: `AUTH_GRPC_PORT`, `USER_GRPC_PORT`, `WALLET_GRPC_PORT`, `GATEWAY_HTTP_ADDR`

### Code Generation Tools

Install required tools:
```bash
make install-tools  # Installs protoc-gen-go, golangci-lint, goimports, migrate
```

## Troubleshooting

**Port conflicts**: Check with `make docker-ps`, stop with `make dev-stop`

**Migration errors**: Use `make db-reset` for fresh start, or `make migrate-force VERSION=N` to fix broken state

**gRPC connection issues**: Ensure services started in correct order (user/wallet before auth). Check `make dev-logs`

**Test failures**: Ensure clean database with `make db-reset` before running integration tests

**Build errors**: Run `go mod tidy` and ensure Go 1.21+ installed

## Related Repositories

- `zuno-marketplace-ui`: Next.js frontend
- `zuno-marketplace-contracts`: Solidity smart contracts (Foundry)
- `zuno-marketplace-sdk`: TypeScript SDK for ABIs
- `zuno-marketplace-abis`: ABI provider service
- `zuno-marketplace-metadata`: NFT metadata storage
- `zuno-marketplace-mini`: Contract testing mini-app

## Branching Strategy

- `main`: Production-ready code
- `feature/*`: Feature branches (merge to main when complete)
- Follow TDD: 🔴 RED (failing test) → 🟢 GREEN (minimal code) → 🔵 REFACTOR

## Commit Message Format

Use conventional commits:
```
<type>(<scope>): <description>

<body>

<footer>
```

**Types**: feat, fix, docs, style, refactor, perf, test, chore
**Scopes**: auth, user, wallet, gateway, database, infra
