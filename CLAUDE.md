# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Role & Responsibilities

Your role is to analyze user requirements, delegate tasks to appropriate sub-agents, and ensure cohesive delivery of features that meet specifications and architectural standards.

## Workflows

<<<<<<< HEAD

## Development Rules

1. **Mandatory Testing**: Always create a test file whenever creating an important file (logic, services, repositories). For a file named `<name>.go`, the test file MUST be named `<name>_test.go`.
2. **TDD Pattern**: Always follow Test-Driven Development patterns (Red -> Green -> Refactor).

### Example: TDD Workflow

When adding a new function `CalculateTotal(items []Item) int`:

1. **🔴 RED**: Create `calculate_test.go` and write a test case that fails (because the function doesn't exist or is empty).
   ```go
   func TestCalculateTotal(t *testing.T) {
       items := []Item{{Price: 10}, {Price: 20}}
       total := CalculateTotal(items)
       if total != 30 {
           t.Errorf("Expected 30, got %d", total)
       }
   }
   ```
2. **🟢 GREEN**: Implement the minimal code in `calculate.go` to make the test pass.
   ```go
   func CalculateTotal(items []Item) int {
       sum := 0
       for _, item := range items {
           sum += item.Price
       }
       return sum
   }
   ```
3. **🔵 REFACTOR**: Optimize code if needed, ensuring tests still pass.
### Service Ports

- **GraphQL Gateway**: http://localhost:8081/graphql (playground at /playground)
- **Auth Service**: gRPC :50051
- **User Service**: gRPC :50052
- **Wallet Service**: gRPC :50053
- **Collection Service**: gRPC :50054
- **Media Service**: gRPC :50055
- **PostgreSQL**: 5432
- **Redis**: 6379
- **RabbitMQ**: 5672 (management UI: http://localhost:15672)

## Architecture & Key Patterns

### Service Communication Flow
=======
- Primary workflow: `./.claude/workflows/primary-workflow.md`
- Development rules: `./.claude/workflows/development-rules.md`
- Orchestration protocols: `./.claude/workflows/orchestration-protocol.md`
- Documentation management: `./.claude/workflows/documentation-management.md`
- And other workflows: `./.claude/workflows/*`

**IMPORTANT:** Analyze the skills catalog and activate the skills that are needed for the task during the process.
**IMPORTANT:** You must follow strictly the development rules in `./.claude/workflows/development-rules.md` file.
**IMPORTANT:** Before you plan or proceed any implementation, always read the `./README.md` file first to get context.
**IMPORTANT:** Sacrifice grammar for the sake of concision when writing reports.
**IMPORTANT:** In reports, list any unresolved questions at the end, if any.
**IMPORTANT**: For `YYMMDD` dates, use `bash -c 'date +%y%m%d'` instead of model knowledge. Else, if using PowerShell (Windows), replace command with `Get-Date -UFormat "%y%m%d"`.

## Documentation Management

We keep all important docs in `./docs` folder and keep updating them, structure like below:
>>>>>>> develop-claude

```
./docs
├── project-overview-pdr.md
├── code-standards.md
├── codebase-summary.md
├── design-guidelines.md
├── deployment-guide.md
├── system-architecture.md
└── project-roadmap.md
```

<<<<<<< HEAD
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

**Migration errors**: Use `make migrate-force VERSION=N` to fix broken state

**gRPC connection issues**: Ensure services started in correct order (user/wallet before auth). Check `make dev-logs`


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
=======
**IMPORTANT:** *MUST READ* and *MUST COMPLY* all *INSTRUCTIONS* in project `./CLAUDE.md`, especially *WORKFLOWS* section is *CRITICALLY IMPORTANT*, this rule is *MANDATORY. NON-NEGOTIABLE. NO EXCEPTIONS. MUST REMEMBER AT ALL TIMES!!!*
>>>>>>> develop-claude
