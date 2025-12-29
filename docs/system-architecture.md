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

### Development (Serverless - Phase 2 Complete)

**Infrastructure Providers**:
- Supabase (PostgreSQL) - Free tier
- Upstash (Redis) - Free tier
- CloudAMQP (RabbitMQ) - Little Lemur free tier

**Configuration**: `.env` with serverless connection strings
**Setup**: Run `./scripts/setup-env.sh` and select "serverless" mode
**Startup**: No infrastructure startup required - just run services
**Benefits**:
- Zero local resource usage
- No Docker overhead
- Shared dev environment
- Quick setup (10 minutes)

**Environment Variables** (see `.env.development.example`):
```bash
# Infrastructure mode
INFRA_MODE=serverless

# Supabase (PostgreSQL)
DATABASE_URL=postgresql://postgres:[PASSWORD]@db.xxx.supabase.co:5432/postgres

# Upstash (Redis)
REDIS_URL=redis://default:[PASSWORD]@xxx.upstash.io:6379

# CloudAMQP (RabbitMQ)
CLOUDAMQP_URL=amqp://user:password@xxx.rmq.cloudamqp.com/vhost
```

**Quick Setup Guides**:
- **Supabase**: https://supabase.com/docs/guides/getting-started
- **Upstash**: https://upstash.com/docs/redis/quickstart/redis
- **CloudAMQP**: https://www.cloudamqp.com/docs/how-to-connection-url.html

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

### Infrastructure Mode Switching (Phase 2 Complete)

The application supports switching between Docker and Serverless modes via environment configuration:

**Interactive Setup**:
```bash
./scripts/setup-env.sh  # Guides you through mode selection
```

**Manual Mode Selection**:

**Docker Mode** (default):
```bash
# Copy template
cp .env.production.example .env

# Uses local infrastructure
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
REDIS_HOST=localhost
REDIS_PORT=6379
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672

# Requires: docker compose up -d
```

**Serverless Mode**:
```bash
# Copy template
cp .env.development.example .env

# Uses cloud infrastructure
DATABASE_URL=postgresql://...
REDIS_URL=redis://...
CLOUDAMQP_URL=amqp://...

# No Docker required
```

## Monitoring & Observability (Future)

### Health Checks

**Implemented**:
- `/health` endpoint on Gateway (returns service statuses)
- Readiness checks on all deployments

**Missing**:
- Detailed metrics (Prometheus)
- Distributed tracing (Jaeger)
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

### Alert Scenarios

- Auth service down
- Database connection pool exhausted
- High login failure rate
- Token validation errors
- gRPC service timeouts

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

**Version**: 1.2
**Last Updated**: 2025-12-29
**Diagram Format**: ASCII (future: Mermaid diagrams)
**Phase 2 Complete**: Environment Configuration (Serverless + Docker modes)
