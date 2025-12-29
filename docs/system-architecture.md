# System Architecture

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend (Web3 DApp)                      │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTP/WebSocket
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│                    GraphQL Gateway (Port 8081)                   │
│  - Query Router                                                  │
│  - Mutation Coordinator                                          │
│  - Subscription Handler                                          │
│  - Auth Middleware                                               │
│  - Rate Limiting                                                 │
└────┬────────────────────────┬────────────────────────────┬──────┘
     │ gRPC                   │ gRPC                       │ gRPC
     ↓                        ↓                            ↓
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│  Auth Service    │  │  User Service    │  │ Wallet Service   │
│  (Port 50051)    │  │  (Port 50052)    │  │ (Port 50053)     │
├──────────────────┤  ├──────────────────┤  ├──────────────────┤
│ - SIWE Verify    │  │ - Profile Mgmt   │  │ - Wallet Link    │
│ - JWT Generation │  │ - User CRUD      │  │ - CAIP-10 Addr   │
│ - Session Mgmt   │  │ - Preferences    │  │ - Verification   │
│ - Login Events   │  │ - Statistics     │  │ - Activity Log   │
└────┬─────────────┘  └────┬─────────────┘  └────┬─────────────┘
     │ SQL                  │ SQL                  │ SQL
     └─────────────────────┬┴──────────────────────┘
                           ↓
                    ┌──────────────────┐
                    │  PostgreSQL 15   │
                    │ (Port 5432)      │
                    ├──────────────────┤
                    │ - 11 tables      │
                    │ - ACID schemas   │
                    │ - Replication    │
                    └──────────────────┘

                    Cache Layer        Event Queue
                    ┌──────────────┐   ┌──────────────┐
                    │ Redis 7      │   │ RabbitMQ 3   │
                    │ (Port 6379)  │   │ (Port 5672)  │
                    └──────────────┘   └──────────────┘
```

## Communication Patterns

### Client → Gateway (HTTP/GraphQL)

**Protocol**: HTTP/1.1 (WebSocket for subscriptions future)
**Format**: GraphQL queries, mutations, subscriptions
**Authentication**: JWT Bearer token in Authorization header
**Port**: 8081

**Request Flow**:

1. Client sends GraphQL query with JWT token
2. Gateway middleware validates token
3. Resolvers fan out to gRPC services
4. Results aggregated and returned

### Service → Service (gRPC)

**Protocol**: gRPC (HTTP/2)
**Format**: Protocol Buffers
**Discovery**: Kubernetes DNS (service.namespace.svc.cluster.local)
**Ports**: 50051 (Auth), 50052 (User), 50053 (Wallet)

**Service Dependencies**:

- Gateway depends on: Auth, User, Wallet
- Auth depends on: User, Wallet (internal calls)
- Wallet depends on: User

**Retry Strategy**: Implemented at resolver level (future)

## Database Architecture

### Schema Organization

Three logical schemas, one physical database:

```
Database: nft_marketplace

├── public/auth_nonces
├── public/sessions
├── public/login_events
│
├── public/users
├── public/profiles
├── public/user_follows
├── public/user_preferences
├── public/user_stats
│
├── public/wallet_links
├── public/wallet_verifications
└── public/wallet_activity
```

### Table Specifications

#### Auth Service Tables

**auth_nonces** - SIWE nonce management

```
id              BIGSERIAL PRIMARY KEY
address         VARCHAR(42) NOT NULL  -- Ethereum address
nonce           VARCHAR(255) NOT NULL -- Random nonce
created_at      TIMESTAMP DEFAULT NOW()
expires_at      TIMESTAMP NOT NULL    -- Nonce validity window
used_at         TIMESTAMP NULL        -- Consumption timestamp
used_by_session UUID NULL             -- Session that consumed it

UNIQUE (address, nonce)
INDEX idx_auth_nonces_address_expires (address, expires_at)
INDEX idx_auth_nonces_used (used_at) -- For cleanup queries
```

**sessions** - User session tracking

```
id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
user_id         UUID NOT NULL REFERENCES users(id)
device_hash     VARCHAR(64)                  -- Device fingerprint
issued_at       TIMESTAMP DEFAULT NOW()
expires_at      TIMESTAMP NOT NULL           -- Token expiry
last_activity   TIMESTAMP DEFAULT NOW()
ip_address      INET                         -- Session origin
user_agent      VARCHAR(512)                 -- Browser info

INDEX idx_sessions_user_id (user_id)
INDEX idx_sessions_expires_at (expires_at)
```

**login_events** - Audit trail

```
id              BIGSERIAL PRIMARY KEY
user_id         UUID NOT NULL REFERENCES users(id)
address         VARCHAR(42) NOT NULL         -- Signing address
event_type      VARCHAR(50) NOT NULL         -- 'login', 'refresh', 'logout'
success         BOOLEAN DEFAULT TRUE
timestamp       TIMESTAMP DEFAULT NOW()
ip_address      INET
error_message   VARCHAR(500) NULL

INDEX idx_login_events_user_id (user_id)
INDEX idx_login_events_timestamp (timestamp)
```

#### User Service Tables

**users** - Core user accounts

```
id              UUID PRIMARY KEY              -- UUID v5(namespace, address)
account_id      VARCHAR(255) UNIQUE           -- CAIP-10: chain:network:address
email           VARCHAR(255) UNIQUE NULL
verified_email  BOOLEAN DEFAULT FALSE
created_at      TIMESTAMP DEFAULT NOW()
updated_at      TIMESTAMP DEFAULT NOW()
deleted_at      TIMESTAMP NULL               -- Soft delete support

INDEX idx_users_account_id (account_id)
INDEX idx_users_email (email)
INDEX idx_users_created_at (created_at)
```

**profiles** - User profile information

```
id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
user_id         UUID NOT NULL UNIQUE REFERENCES users(id)
username        VARCHAR(255) UNIQUE NULL
avatar_url      VARCHAR(1024) NULL
bio             TEXT NULL
website         VARCHAR(1024) NULL
location        VARCHAR(255) NULL
updated_at      TIMESTAMP DEFAULT NOW()

FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
```

**user_follows** - Social relationships

```
follower_id     UUID NOT NULL REFERENCES users(id)
following_id    UUID NOT NULL REFERENCES users(id)
created_at      TIMESTAMP DEFAULT NOW()

PRIMARY KEY (follower_id, following_id)
CONSTRAINT followers_different CHECK (follower_id != following_id)
INDEX idx_user_follows_following (following_id)
```

**user_preferences** - User settings

```
id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
user_id         UUID NOT NULL UNIQUE REFERENCES users(id)
email_notify    BOOLEAN DEFAULT TRUE
push_notify     BOOLEAN DEFAULT TRUE
language        VARCHAR(10) DEFAULT 'en'
theme           VARCHAR(20) DEFAULT 'light'
updated_at      TIMESTAMP DEFAULT NOW()

FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
```

**user_stats** - Activity statistics

```
id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
user_id         UUID NOT NULL UNIQUE REFERENCES users(id)
total_logins    BIGINT DEFAULT 0
followers_count BIGINT DEFAULT 0
following_count BIGINT DEFAULT 0
last_login      TIMESTAMP NULL
updated_at      TIMESTAMP DEFAULT NOW()

FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
```

#### Wallet Service Tables

**wallet_links** - Multi-wallet associations

```
id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
user_id         UUID NOT NULL REFERENCES users(id)
address         VARCHAR(42) NOT NULL           -- Ethereum address
chain_id        BIGINT NOT NULL               -- EVM chain ID
caip10_account  VARCHAR(255) NOT NULL         -- Normalized CAIP-10
is_primary      BOOLEAN DEFAULT FALSE          -- Primary wallet
linked_at       TIMESTAMP DEFAULT NOW()
verified_at     TIMESTAMP NULL                -- SIWE verification timestamp
deleted_at      TIMESTAMP NULL                -- Soft delete

UNIQUE (user_id, address, chain_id)
UNIQUE (user_id, is_primary) WHERE is_primary = TRUE  -- One primary max
INDEX idx_wallet_links_address (address)
INDEX idx_wallet_links_caip10 (caip10_account)
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
```

**wallet_verifications** - Verification status tracking

```
id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
wallet_link_id  UUID NOT NULL REFERENCES wallet_links(id)
message_hash    VARCHAR(66) NOT NULL          -- keccak256(message)
signature       VARCHAR(132) NOT NULL          -- Raw signature
verified_at     TIMESTAMP NOT NULL             -- When verified
expires_at      TIMESTAMP                      -- Verification validity
status          VARCHAR(50) DEFAULT 'verified' -- verified, expired, revoked

UNIQUE (wallet_link_id, message_hash)
INDEX idx_wallet_verifications_verified_at (verified_at)
FOREIGN KEY (wallet_link_id) REFERENCES wallet_links(id)
```

**wallet_activity** - Transaction history

```
id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
wallet_link_id  UUID NOT NULL REFERENCES wallet_links(id)
activity_type   VARCHAR(50) NOT NULL          -- 'transfer', 'swap', 'nft_mint'
tx_hash         VARCHAR(66) NULL               -- Transaction hash
amount          NUMERIC(40, 18) NULL          -- Token amount
token_address   VARCHAR(42) NULL               -- Contract address
timestamp       TIMESTAMP DEFAULT NOW()
metadata        JSONB NULL                    -- Extra data

INDEX idx_wallet_activity_wallet_link_id (wallet_link_id)
INDEX idx_wallet_activity_timestamp (timestamp)
FOREIGN KEY (wallet_link_id) REFERENCES wallet_links(id)
```

### Query Patterns

**Atomic Nonce Consumption**:

```sql
BEGIN;
  UPDATE auth_nonces
  SET used_at = NOW(), used_by_session = $1
  WHERE address = $2 AND nonce = $3 AND used_at IS NULL
  RETURNING id;
COMMIT;
```

**Ensure User Idempotency**:

```sql
INSERT INTO users (id, account_id, email, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
ON CONFLICT (id) DO UPDATE
SET email = COALESCE(EXCLUDED.email, users.email),
    updated_at = NOW()
RETURNING *;
```

**User-Wallet Association**:

```sql
SELECT u.id, u.account_id, wl.address, wl.chain_id
FROM users u
LEFT JOIN wallet_links wl ON u.id = wl.user_id
WHERE u.id = $1 AND wl.deleted_at IS NULL;
```

## Authentication Flow

### SIWE Login Flow

```
┌─────────────────┐
│  1. Get Nonce   │
│  Client → API   │
└────────┬────────┘
         │ Response: random nonce
         ↓
┌─────────────────────────────┐
│ 2. Sign Message with Wallet │
│ Client signs SIWE message   │
│ with private key            │
└────────┬────────────────────┘
         │ Signed message + signature
         ↓
┌────────────────────────────┐
│ 3. Verify Signature        │
│ Auth Service checks:       │
│ - Signature valid          │
│ - Nonce not used           │
│ - Message recent           │
└────────┬───────────────────┘
         │ ✓ Valid
         ↓
┌────────────────────────────┐
│ 4. Consume Nonce Atomically│
│ Mark nonce as used         │
│ Prevent replay attacks     │
└────────┬───────────────────┘
         │
         ↓
┌────────────────────────────┐
│ 5. Ensure User Exists      │
│ Call User Service          │
│ Create if needed (idempotent)
└────────┬───────────────────┘
         │ user_id
         ↓
┌────────────────────────────┐
│ 6. Generate Tokens         │
│ - Access token (15 min)    │
│ - Refresh token (7 days)   │
└────────┬───────────────────┘
         │ JWT tokens
         ↓
┌────────────────────────────┐
│ 7. Create Session          │
│ Store device fingerprint   │
│ Track last activity        │
└────────┬───────────────────┘
         │
         ↓
┌────────────────────────────┐
│ 8. Log Event               │
│ Record login for audit     │
│ (fire-and-forget)          │
└────────┬───────────────────┘
         │ tokens
         ↓
     ┌───────┐
     │Client │
     └───────┘
```

### Token Structure

**Access Token** (JWT, short-lived)

```
Header:
{
  "alg": "HS256",
  "typ": "JWT",
  "kid": "access-token"
}

Payload:
{
  "sub": "user-uuid",
  "aud": "graphql-gateway",
  "iss": "auth-service",
  "iat": 1701705600,
  "exp": 1701706500,        // 15 min expiry
  "session_id": "session-uuid",
  "address": "0x1234...",
  "email": "user@example.com"
}

Signature: HMAC-SHA256(secret)
```

**Refresh Token** (JWT, long-lived, rotates)

```
Header:
{
  "alg": "HS256",
  "typ": "JWT",
  "kid": "refresh-token"
}

Payload:
{
  "sub": "user-uuid",
  "aud": "auth-service",
  "iss": "auth-service",
  "iat": 1701705600,
  "exp": 1702310400,         // 7 day expiry
  "session_id": "session-uuid",
  "token_family": "family-uuid",  // For rotation tracking
  "generation": 1,                // For replay detection
  "session_fingerprint": "hash"
}

Signature: HMAC-SHA256(refresh-secret)
```

### Security Features

#### Nonce Management

- Random nonce generated per auth request
- Single-use (atomic database constraint)
- 10-minute expiration window
- Prevents replay attacks

#### Refresh Token Rotation

- New token issued on each refresh
- Token family tracks related tokens
- Reuse detection on invalid family
- Suspicious activity triggers session revocation

#### Session Fingerprinting

- Device hash = MD5(User-Agent + IP)
- Mismatch triggers re-authentication
- Prevents session token theft

#### Token Expiry

- Access: 15 minutes (low risk window)
- Refresh: 7 days (long maintenance)
- Nonce: 10 minutes (auth window)

## Service Communication

### gRPC Service Interfaces

**Auth Service** (AuthServiceServer)

```protobuf
service Auth {
  rpc GetNonce(GetNonceRequest) returns (GetNonceResponse)
  rpc VerifySIWE(VerifySIWERequest) returns (VerifySIWEResponse)
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse)
  rpc GetSession(GetSessionRequest) returns (GetSessionResponse)
  rpc LogEvent(LogEventRequest) returns (LogEventResponse)
}
```

**User Service** (UserServiceServer)

```protobuf
service User {
  rpc GetUser(GetUserRequest) returns (GetUserResponse)
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse)
  rpc UpdateProfile(UpdateProfileRequest) returns (UpdateProfileResponse)
  rpc GetFollows(GetFollowsRequest) returns (GetFollowsResponse)
  rpc Follow(FollowRequest) returns (FollowResponse)
}
```

**Wallet Service** (WalletServiceServer)

```protobuf
service Wallet {
  rpc LinkWallet(LinkWalletRequest) returns (LinkWalletResponse)
  rpc GetWallets(GetWalletsRequest) returns (GetWalletsResponse)
  rpc SetPrimaryWallet(SetPrimaryWalletRequest) returns (SetPrimaryWalletResponse)
}
```

### Service-to-Service Calls

**Initialization** (in GraphQL Gateway):

```go
// Create gRPC connections at startup
authConn, _ := grpc.Dial("auth-service:50051")
userConn, _ := grpc.Dial("user-service:50052")
walletConn, _ := grpc.Dial("wallet-service:50053")

authClient := pb.NewAuthClient(authConn)
userClient := pb.NewUserClient(userConn)
walletClient := pb.NewWalletClient(walletConn)
```

**Within Services** (Auth calling User):

```go
// Auth service depends on User service
type authServer struct {
  userClient pb.UserClient
}

// Call User service when logging in
user, err := s.userClient.CreateUser(ctx, &pb.CreateUserRequest{
  AccountID: accountID,
})
```

## Infrastructure Components

### PostgreSQL Database

**Configuration**:

- Version: PostgreSQL 15
- Container: `postgres:15-alpine`
- Port: 5432
- Volume: `postgres-data` (persisted)
- User: postgres
- Database: nft_marketplace

**Connection Pool** (from each service):

```
Max Open Connections: 25
Max Idle Connections: 5
Connection Max Lifetime: 5 minutes
```

**Replication** (future):

- Primary-standby replication
- Automatic failover
- Backup recovery procedures

### Redis Cache

**Configuration**:

- Version: Redis 7
- Container: `redis:7-alpine`
- Port: 6379
- Volume: `redis-data` (persisted)
- Maxmemory: 256MB
- Eviction: allkeys-lru

**Current Usage**: None (infrastructure ready)

**Planned Usage**:

- Session storage (30-day TTL)
- User preferences cache (24-hour TTL)
- Rate limit counters (per-minute)

### RabbitMQ Message Queue

**Configuration**:

- Version: RabbitMQ 3 (latest)
- Container: `rabbitmq:3-management`
- AMQP Port: 5672
- Management UI: 15672 (guest/guest)
- Volume: `rabbitmq-data` (persisted)

**Current Usage**: None (infrastructure ready)

**Planned Exchanges**:

- `nft_events` - Direct exchange for service events
- `audit_logs` - Topic exchange for audit trail

**Planned Queues**:

- `login_events.queue` - Login audit logging
- `user_updates.queue` - User state changes
- `wallet_updates.queue` - Wallet activity

## Deployment Architectures

### Development (Docker Compose)

**Services**:

- postgres (database)
- redis (cache)
- rabbitmq (message queue)
- auth-service
- user-service
- wallet-service
- graphql-gateway

**Configuration**: `docker-compose.yml`
**Startup**: `docker compose up -d`
**Logs**: `docker compose logs -f`

### Development (Kubernetes + Tilt)

**Components**:

- PostgreSQL StatefulSet (persistence via PVC)
- Redis Deployment (in-memory cache)
- RabbitMQ StatefulSet (message broker)
- Auth Service Deployment (replicas: 1)
- User Service Deployment (replicas: 1)
- Wallet Service Deployment (replicas: 1)
- GraphQL Gateway Deployment (replicas: 2)

**Features**:

- Hot reload via Tilt
- Service DNS discovery
- ConfigMap for config
- Secrets for sensitive data
- Health checks enabled

**Startup**: `tilt up`
**UI**: http://localhost:10350

### Production (Kubernetes - TODO)

**Missing Configurations**:

- Ingress controller
- TLS/SSL certificates
- Pod autoscaling
- Resource limits
- Persistent volumes
- Backup strategy
- Monitoring stack
- Logging aggregation

## Monitoring & Observability

### Error Tracking (Sentry)

**Status**: Phase 04 Complete (Core + Middleware + Service Integration + Distributed Tracing)

**Implementation**: `shared/observability/sentry/` + `shared/observability/middleware/` + `shared/observability/tracing/`

**Integrated Services**:

- Auth Service (Port 50051)
- User Service (Port 50052)
- Wallet Service (Port 50053)
- GraphQL Gateway (Port 8081)

**Features**:

- **Automatic Error Capture**: Captures unhandled panics and exceptions
- **Performance Monitoring**: Distributed tracing with configurable sampling rates
- **Privacy-First**: Automatic scrubbing of sensitive data before transmission
- **Breadcrumbs**: Context tracking for user actions leading to errors
- **Stack Traces**: Full stack trace attachment for debugging

**Privacy Scrubbing Patterns**:

```
- Ethereum addresses: 0x[a-fA-F0-9]{40} → [FILTERED:ETH_ADDRESS]
- CAIP-10 IDs: eip155:1:0x... → [FILTERED:CAIP10]
- JWT tokens: eyJ... → [FILTERED:JWT]
- Email: user@domain.com → [FILTERED:EMAIL]
- Private keys: 64-char hex → [FILTERED:PRIVATE_KEY]
- Sensitive headers: Authorization, Cookie, X-API-Key → [FILTERED]
```

**Integration Pattern** (per service):

```go
import obs "github.com/zunokit/zuno-marketplace-api/shared/observability/sentry"
import obsTrace "github.com/zunokit/zuno-marketplace-api/shared/observability/tracing"

func main() {
    // Initialize Sentry with environment-based sampling
    obs.Init(
        cfg.Sentry.DSN,
        cfg.Sentry.Environment,
        "auth-service",     // service name
        "v1.0.0",          // release version
        obsTrace.GetTracesSampleRate(cfg.Sentry.Environment), // 100% dev, 20% staging, 5% prod
    )
    defer obs.Flush(2 * time.Second)

    // Capture errors
    if err != nil {
        obs.CaptureException(err)
    }

    // Get trace ID for logging
    traceID := obsTrace.GetTraceID(ctx)
    log.Printf("Processing trace: %s", traceID)
}
```

**Configuration Requirements** (per service):

```bash
SENTRY_DSN=https://...@sentry.io/...
SENTRY_ENVIRONMENT=production|staging|development  # Determines sampling rate automatically
SENTRY_RELEASE=v1.0.0
# Sampling rate is now auto-calculated based on environment:
# - development: 100%
# - staging: 20%
# - production: 5%
```

**Configuration Decisions**:

- **Sentry Project Strategy**: Share one Sentry project for all services (current), with future scaling to allow per-service projects
- **Trace Sampling Rate**: Environment-based (100% dev, 20% staging, 5% prod) - balances debugging needs with cost control
- **Smart Sampling**: Health checks excluded from tracing (0%), auth operations always traced (100%)
- **Custom Tags**: Add `user_id` and `wallet_hash` tags for user context (wallet address SHA256 hashed)

**Recommended Tag Usage**:

```go
// Add user context with hashed wallet
import "crypto/sha256"

walletHash := sha256.Sum256([]byte(walletAddress))
sentry.ConfigureScope(func(scope *sentry.Scope) {
    scope.SetTag("user_id", userID)
    scope.SetTag("wallet_hash", hex.EncodeToString(walletHash[:])[:16]) // First 16 chars
})
```

### Middleware Architecture (Phase 02 - 03)

**Implementation**: `shared/observability/middleware/` + Service Integration

The middleware layer provides automatic distributed tracing across all service communication protocols. All services now have Sentry middleware integrated.

#### HTTP Middleware (Chi)

**File**: `middleware/http.go`

**Purpose**: Capture HTTP requests as Sentry transactions with distributed tracing support.

**Features**:

- Health/ready endpoint skip (`/health`, `/ready`) for clean trace data
- Transaction name from HTTP method + path (e.g., `GET /api/v1/users`)
- HTTP context capture: method, URL, scheme, host, path, query, remote_addr
- Distributed tracing via `sentry-trace` header extraction
- Custom `responseWriter` wrapper for status code capture
- Status code to span status mapping:
  - 4xx → `InvalidArgument`
  - 5xx → `InternalError`
  - Others → `OK`

**Usage**:

```go
import obshttp "github.com/zunokit/zuno-marketplace-api/shared/observability/middleware"

router := chi.NewRouter()
router.Use(obshttp.SentryHTTP)  // Add BEFORE other middleware
router.Use(middleware.Logger)
router.Use(middleware.Recoverer)
```

#### gRPC Interceptors

**File**: `middleware/grpc.go`

**Purpose**: Distributed tracing for gRPC service-to-service communication.

**UnaryServerInterceptor**:

- Captures incoming gRPC calls as Sentry spans
- Extracts `sentry-trace` from incoming metadata for distributed tracing
- Captures method name and service info
- Error mapping to span status

**UnaryClientInterceptor**:

- Injects `sentry-trace` header into outbound gRPC calls
- Captures target service and method info
- Formats trace header: `{trace_id}-{span_id}-{sampled}`

**StreamServerInterceptor**:

- Support for streaming gRPC RPCs
- Context propagation via `streamWithContext` wrapper

**Usage**:

```go
import obsgrpc "github.com/zunokit/zuno-marketplace-api/shared/observability/middleware"

// Server
grpcServer := grpc.NewServer(
    grpc.ChainUnaryInterceptor(obsgrpc.UnaryServerInterceptor()),
)

// Client
conn, err := grpc.Dial(
    serviceURL,
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithChainUnaryInterceptor(obsgrpc.UnaryClientInterceptor()),
)
```

#### GraphQL Middleware

**File**: `middleware/graphql.go`

**Purpose**: Field-level and operation-level tracing for GraphQL resolvers.

**GraphQLFieldMiddleware**:

- Creates span for each GraphQL field resolution
- Captures field name, type, parent type
- Operation context (query/mutation/subscription)
- Variables count (sanitized, no raw values)
- Panic recovery with proper span cleanup

**GraphQLResponseMiddleware**:

- Operation-level metrics
- Error count tracking
- Operation type classification

**Usage**:

```go
import obsgraphql "github.com/zunokit/zuno-marketplace-api/shared/observability/middleware"

srv := handler.NewDefaultServer(schema)
srv.Use(obsgraphql.GraphQLFieldMiddleware())
srv.Use(obsgraphql.GraphQLResponseMiddleware())
```

#### Distributed Tracing Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Distributed Tracing Flow                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Client Request                                                         │
│       │                                                                 │
│       ▼                                                                 │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ HTTP Middleware (Chi)                                             │  │
│  │  - Start transaction: "GET /api/v1/users"                        │  │
│  │  - Extract sentry-trace header (if present)                      │  │
│  │  - Set trace context on request                                  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│       │                                                                 │
│       ▼                                                                 │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ GraphQL Gateway Resolver                                          │  │
│  │  - GraphQLFieldMiddleware creates span for each field            │  │
│  │  - gRPC client interceptor injects sentry-trace                  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│       │                                                                 │
│       ▼ (sentry-trace header in metadata)                               │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ gRPC Service (Auth/User/Wallet)                                   │  │
│  │  - UnaryServerInterceptor extracts sentry-trace                  │  │
│  │  - Continues parent span from trace header                       │  │
│  │  - Creates child span for gRPC method                            │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│       │                                                                 │
│       ▼                                                                 │
│  Sentry receives distributed trace with linked spans                    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Test Coverage

**File**: `middleware/middleware_test.go`

Tests: 6/6 passing

- `TestSentryHTTP`: Health/ready endpoint skip, regular tracing, POST requests
- `TestResponseWriter`: Status code wrapper functionality
- `TestUnaryServerInterceptor`: gRPC server request handling
- `TestUnaryClientInterceptor`: gRPC client call handling
- `TestFormatTraceHeader`: Trace header format validation
- `TestStreamServerInterceptor`: gRPC streaming support

#### Service Integration (Phase 03)

**Implementation**: Service-level Sentry initialization and middleware integration

All 4 services now have integrated Sentry observability:

**Auth Service** (`services/auth-service/cmd/main.go`):

- Sentry initialization with non-blocking approach
- gRPC server interceptor for incoming request tracing
- Graceful shutdown with Sentry flush
- Config: `SentryConfig` struct with DSN and Environment

**User Service** (`services/user-service/cmd/main.go`):

- Sentry initialization with non-blocking approach
- gRPC server interceptor for incoming request tracing
- Graceful shutdown with Sentry flush
- Config: `SentryConfig` struct with DSN and Environment

**Wallet Service** (`services/wallet-service/cmd/main.go`):

- Sentry initialization with non-blocking approach
- gRPC server interceptor for incoming request tracing
- Graceful shutdown with Sentry flush
- Config: `SentryConfig` struct with DSN and Environment

**GraphQL Gateway** (`services/graphql-gateway/cmd/main.go`):

- Sentry initialization with non-blocking approach
- HTTP middleware (`obs.SentryHTTP`) for request transaction tracking
- gRPC client interceptors for all service connections (distributed tracing)
- Graceful shutdown with Sentry flush
- Config: `SentryConfig` struct with DSN and Environment

**Configuration Pattern** (per service):

```go
// Load configuration
cfg := config.Load()

// Initialize Sentry (non-blocking on failure)
if cfg.Sentry.DSN != "" {
    if err := obs.Init(
        cfg.Sentry.DSN,
        cfg.Sentry.Environment,
        "service-name",      // service-specific
        getBuildVersion(),   // version
        0.2,                 // 20% sampling
    ); err != nil {
        log.Printf("Sentry init failed (continuing): %v", err)
    } else {
        log.Println("Sentry initialized")
        defer obs.Flush(2 * time.Second)
    }
} else {
    log.Println("Sentry DSN not configured, skipping")
}
```

**Environment Variables**:

```bash
SENTRY_DSN=https://...@sentry.io/...      # Required for Sentry
SENTRY_ENVIRONMENT=development              # development|staging|production
```

**Distributed Tracing End-to-End Flow**:

```
Client Request
    │
    ├──► GraphQL Gateway (HTTP)
    │     │
    │     ├──► obs.SentryHTTP (transaction start)
    │     │
    │     └──► gRPC Client (auth/user/wallet)
    │           │
    │           └──► obs.UnaryClientInterceptor (inject sentry-trace)
    │                 │
    │                 └──► gRPC Service
    │                       │
    │                       └──► obs.UnaryServerInterceptor (extract sentry-trace, continue span)
    │
    └──► Sentry receives complete distributed trace
```

#### Smart Sampling & Trace Helpers (Phase 04)

**Implementation**: `shared/observability/tracing/` + Service Integration

**Environment-Based Sampling**:

- Development: 100% (full traces for debugging)
- Staging: 20% (balanced visibility)
- Production: 5% (cost control)

**Smart Sampling Logic** (via `TracesSampler`):

- Health checks always excluded (0% sampling):
  - `GET /health`
  - `GET /ready`
  - `grpc.health.v1.Health/Check`
- Auth operations always traced (100% sampling):
  - `VerifySIWE`
  - `RefreshToken`
  - `Login`
  - `Logout`
  - `Authenticate`
  - `Authorize`
  - `signIn`/`signOut`

**Trace Helper Functions**:

```go
import obsTrace "github.com/zunokit/zuno-marketplace-api/shared/observability/tracing"

// Get sampling rate by environment
rate := obsTrace.GetTracesSampleRate("production") // 0.05

// Get smart sampler with endpoint-specific logic
sampler := obsTrace.TracesSampler("staging")

// Inject trace context into outbound gRPC calls
ctx = obsTrace.InjectTraceContext(ctx)

// Get current trace ID for logging/correlation
traceID := obsTrace.GetTraceID(ctx)
spanID := obsTrace.GetSpanID(ctx)
```

**Service Integration Updates**:
All 4 services now use `obsTrace.GetTracesSampleRate(cfg.Sentry.Environment)`:

```go
// services/auth-service/cmd/main.go
obs.Init(
    cfg.Sentry.DSN,
    cfg.Sentry.Environment,
    "auth-service",
    getBuildVersion(),
    obsTrace.GetTracesSampleRate(cfg.Sentry.Environment), // Dynamic by env
)

// services/user-service/cmd/main.go
// services/wallet-service/cmd/main.go
// services/graphql-gateway/cmd/main.go
// (same pattern applied)
```

### Health Checks

**Implemented**:

- `/health` endpoint on Gateway (returns service statuses)
- Readiness checks on all deployments

**Missing**:

- Detailed metrics (Prometheus)
- Log aggregation (ELK/Loki)
- Alert rules (AlertManager)

### Metrics to Track

- API response time (p50, p95, p99)
- gRPC call latency
- Database query performance
- Token refresh rate
- Login success/failure ratio
- Cache hit rate
- Queue depth
- Error rate by service

### Alert Scenarios

- Auth service down
- Database connection pool exhausted
- High login failure rate
- Token validation errors
- gRPC service timeouts
- Spike in error rate

## Security Architecture

### Token Security

**JWT Validation**:

- Signature verification (HMAC-SHA256)
- Expiry check (reject expired tokens)
- Algorithm check (only HS256 accepted)
- Audience validation

**Token Storage** (client):

- Access token: Memory (cleared on tab close)
- Refresh token: Secure HTTP-only cookie

**Token Transmission**:

- Access: Authorization header (Bearer)
- Refresh: HTTP-only cookie (automatic)

### Network Security

**In-Transit**:

- gRPC over TLS (future, development uses plain)
- HTTP/HTTPS for clients (future: HTTPS enforced)
- Service-to-service mTLS (future)

**At-Rest**:

- Database: TDE (transparent data encryption, future)
- Secrets: Kubernetes Secrets (encrypted etcd backend)
- Tokens: Signed, not encrypted (trust platform)

### Access Control

**Unauthenticated Endpoints**:

- `POST /graphql` query (GetNonce, VerifySIWE)
- `GET /health`

**Authenticated Endpoints**:

- All other queries and mutations
- JWT required in Authorization header
- Device fingerprint validation

## Unresolved Architecture Questions

1. Should gRPC services implement circuit breaker pattern?
2. What retry strategy for transient service failures?
3. Should we use service mesh (Istio) for production?
4. How to handle database failover in Kubernetes?
5. What's the backup recovery time objective (RTO)?
6. Should we implement API versioning (v1, v2)?
7. How to handle cross-service transaction rollback?
8. Should Redis be primary or purely cache layer?

---

**Version**: 1.1
**Last Updated**: 2025-12-29
**Diagram Format**: ASCII (future: Mermaid diagrams)
