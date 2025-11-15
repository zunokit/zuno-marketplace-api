# Zuno NFT Marketplace API

Clean, production-ready microservices backend for NFT Marketplace with focus on wallet authentication.

## 🎯 Current Status

**Version**: 0.1.0 (Clean Skeleton)
**Branch**: `develop-claude`
**Last Cleanup**: Nov 15, 2025

✅ All AI-generated bloat removed (95,819 lines deleted)
✅ Database schemas tested and working
✅ Infrastructure services running
⏳ Ready for feature development with TDD

## 📁 Project Structure

```
services/
├── auth-service/           # SIWE Authentication
│   ├── cmd/.gitkeep       # Empty - ready for development
│   ├── internal/.gitkeep  # Empty - ready for development
│   ├── test/.gitkeep      # Empty - ready for tests
│   ├── db/up.sql          # ✅ Database schema (tested)
│   └── migrations/        # ✅ Migration files
├── user-service/          # User Profiles & Management
│   ├── cmd/.gitkeep
│   ├── internal/.gitkeep
│   ├── test/.gitkeep
│   └── db/up.sql          # ✅ Database schema (tested)
├── wallet-service/        # Wallet Management
│   ├── cmd/.gitkeep
│   ├── internal/.gitkeep
│   ├── test/.gitkeep
│   └── db/up.sql          # ✅ Database schema (tested)
└── graphql-gateway/       # GraphQL BFF API
    ├── graphql/.gitkeep
    └── internal/.gitkeep
```

## 🗄️ Database Schema (Tested & Working)

**PostgreSQL Tables** (11 total):

**Auth Service** (`auth_nonces`, `sessions`, `login_events`)
- SIWE nonce management
- Session tracking with device fingerprinting
- Login event history

**User Service** (`users`, `profiles`, `user_follows`, `user_preferences`, `user_stats`)
- User account management
- Social features (follows)
- User preferences and statistics

**Wallet Service** (`wallet_links`, `wallet_verifications`, `wallet_activity`)
- Multi-wallet support per user
- Wallet verification (SIWE)
- Wallet activity tracking

## 🚀 Quick Start

### Prerequisites

- Docker Desktop
- Go 1.25.1+
- PostgreSQL client (optional, for manual DB access)

### Start Infrastructure

```bash
# Start postgres, redis, rabbitmq
docker compose up -d

# Check services status
docker compose ps

# Test database connection
docker exec nft-postgres psql -U postgres -d nft_marketplace -c "\dt"

# Test Redis
docker exec nft-redis redis-cli ping

# Test RabbitMQ
docker exec nft-rabbitmq rabbitmq-diagnostics ping
```

### Environment Variables

Copy `.env.example` to `.env`:

```env
# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DATABASE=nft_marketplace

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_EXCHANGE=nft_events

# JWT Secrets (Change in production!)
JWT_SECRET=your-256-bit-secret-key-here
REFRESH_SECRET=your-256-bit-refresh-secret-key-here
```

## 📊 Infrastructure Test Results

✅ **PostgreSQL**: Running on port 5432
- Database: `nft_marketplace`
- User: `postgres`
- 11 tables loaded successfully
- Schemas from: auth-service, user-service, wallet-service

✅ **Redis**: Running on port 6379
- Status: PONG
- Ready for session caching

✅ **RabbitMQ**: Running on ports 5672, 15672
- Management UI: http://localhost:15672 (guest/guest)
- Status: Ping succeeded
- Ready for event messaging

## 🎯 Next Steps

### Feature Development Workflow

1. **Create feature branch** from `develop-claude`:
   ```bash
   git checkout develop-claude
   git pull origin develop-claude
   git checkout -b feature/your-feature-name
   ```

2. **Follow TDD** (Test-Driven Development):
   - 🔴 RED: Write failing tests first
   - 🟢 GREEN: Write minimal code to pass tests
   - 🔵 REFACTOR: Improve code while keeping tests green

3. **Commit and merge**:
   ```bash
   git add .
   git commit -m "feat(scope): your feature description"
   git push origin feature/your-feature-name
   # Merge to develop-claude when done
   ```

### First Feature: Wallet Authentication

Recommended implementation order:

1. **Proto Definitions** (gRPC interfaces)
2. **Auth Service** (SIWE authentication)
3. **Wallet Service** (Wallet management)
4. **User Service** (User profiles)
5. **GraphQL Gateway** (BFF API)
6. **Integration Tests** (E2E flows)

## 🧪 Testing Strategy

- **Unit Tests**: Test individual functions/methods
- **Integration Tests**: Test with real databases (testcontainers)
- **E2E Tests**: Test complete user flows
- **Minimum Coverage**: 80% for new code

## 📚 Related Repositories

- `zuno-marketplace-ui`: Next.js frontend
- `zuno-marketplace-contracts`: Solidity smart contracts (Foundry)
- `zuno-marketplace-sdk`: TypeScript SDK for ABIs and contracts
- `zuno-marketplace-abis`: ABI provider service
- `zuno-marketplace-metadata`: NFT metadata storage
- `zuno-marketplace-mini`: Quick contract testing mini-app

## 🔧 Development Tools

```bash
# Install Go dependencies
go mod download

# Generate protobuf code (when proto files exist)
make generate-proto

# Run tests
go test ./...

# Run specific service tests
cd services/auth-service
go test ./...
```

## 📝 Commit Message Format

Follow conventional commits:

```
<type>(<scope>): <description>

<body> (min 100 characters)

<footer>
```

**Types**: feat, fix, docs, style, refactor, perf, test, chore
**Scopes**: auth, user, wallet, gateway, database, infra

## 🏗️ Architecture

**Pattern**: GraphQL Gateway + gRPC Microservices

```
Frontend → GraphQL Gateway (HTTP/WS) → gRPC Services
                                      ↓
                                  RabbitMQ Events
```

**Why This Stack?**
- **gRPC**: Fast, typed internal communication
- **GraphQL**: Flexible API for frontend
- **RabbitMQ**: Reliable event messaging
- **PostgreSQL**: ACID compliance for critical data
- **Redis**: Fast session/cache storage

## 📄 License

MIT

---

**Status**: 🟢 Clean skeleton ready for development
**Last Updated**: 2025-11-15
**Maintainer**: Zuno Team
