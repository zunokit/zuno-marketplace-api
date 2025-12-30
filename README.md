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
- Go 1.21+
- Make (optional but recommended)
- Tilt (optional, for hot reload)

### Option 1: Docker Compose (Simple)

```bash
# Start infrastructure services
make dev
# or
docker compose up -d

# Check services status
docker compose ps

# View logs
make dev-logs
```

### Option 2: Tilt (Hot Reload - Recommended)

```bash
# Start Tilt (requires Kubernetes enabled in Docker Desktop)
make tilt-up
# or
tilt up

# Open Tilt UI: http://localhost:10350

# Stop Tilt
make tilt-down
```

### Common Commands

```bash
# Show all available commands
make help

# Run tests
make test

# Build services
make build

# Generate protobuf code
make proto

# Run linter
make lint

# Format code
make format

# Run CI pipeline locally
make ci
```

**📖 For detailed development guide, see [DEVELOPMENT.md](DEVELOPMENT.md)**

---

## 🌩️ Serverless Development (NEW - Recommended)

**No Docker required!** Use free-tier cloud services for development.

### Benefits
- ✅ Zero local resource usage
- ✅ $0/month (free tiers)
- ✅ Quick 10-minute setup

### Quick Setup

```bash
# 1. Create free accounts
#    - Supabase (PostgreSQL): https://supabase.com
#    - Upstash (Redis): https://upstash.com
#    - CloudAMQP (RabbitMQ): https://www.cloudamqp.com

# 2. Configure environment
./scripts/setup-env.sh  # Select option 1 (serverless)

# 3. Add connection strings to .env
#    DATABASE_URL=postgresql://...
#    REDIS_URL=redis://...
#    CLOUDAMQP_URL=amqp://...

# 4. Verify setup
./scripts/health-check.sh

# 5. Run services (no Docker needed!)
go run ./services/auth-service/cmd/main.go
```

**Full guide:** [Serverless Development Guide](./docs/serverless-development.md)

### Switching Modes

```bash
# Docker mode (current setup)
INFRA_MODE=docker
docker compose up -d

# Serverless mode (new)
INFRA_MODE=serverless
# No Docker needed!
```

---

## 🐳 Docker Development (Current)

_Existing Docker setup remains fully functional_

See "Quick Start" section above for Docker instructions.

### Environment Variables

The project supports two infrastructure modes:

**Option 1: Docker (default)**
- Uses local Docker containers for PostgreSQL, Redis, and RabbitMQ
- Requires Docker Desktop running
- More resource intensive

**Option 2: Serverless**
- Uses external cloud services (Supabase, Upstash, CloudAMQP)
- No Docker required for infrastructure services
- Free tier accounts available

#### Quick Setup

Run the interactive setup script:

```bash
./scripts/setup-env.sh
```

This will guide you through selecting your infrastructure mode and create the appropriate `.env` file.

#### Manual Setup

For Docker mode, copy `.env.production.example`:
```bash
cp .env.production.example .env
docker compose up -d
```

For Serverless mode, copy `.env.development.example`:
```bash
cp .env.development.example .env
# Then edit .env with your cloud service connection strings
```

#### Environment Variables Reference

**Docker Mode (.env.production.example):**
```env
INFRA_MODE=docker
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DATABASE=nft_marketplace
REDIS_HOST=redis
REDIS_PORT=6379
RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
```

**Serverless Mode (.env.development.example):**
```env
INFRA_MODE=serverless
DATABASE_URL=postgresql://postgres:[PASSWORD]@db.supabase.co:5432/postgres
REDIS_URL=redis://default:[PASSWORD]@xxx.upstash.io:6379
CLOUDAMQP_URL=amqp://user:password@host/vhost
```

**Common Variables:**
```env
JWT_SECRET=your-256-bit-secret-key-here
REFRESH_SECRET=your-256-bit-refresh-secret-key-here
AUTH_GRPC_PORT=50051
USER_GRPC_PORT=50052
WALLET_GRPC_PORT=50053
GATEWAY_HTTP_ADDR=8080
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

**Makefile** - Comprehensive development commands
```bash
make help          # Show all commands
make dev           # Start development environment
make test          # Run tests
make build         # Build all services
make lint          # Run linter
make ci            # Run CI pipeline locally
```

**Tiltfile** - Hot reload development with Kubernetes
- 🔥 Automatic rebuilds on code changes
- 📊 Real-time logs and resource monitoring
- 🎯 Interactive UI at http://localhost:10350

**GitHub Actions** - CI/CD Pipeline
- ✅ Automated linting and testing
- 🐳 Docker image builds
- 📊 Test coverage reports

See [DEVELOPMENT.md](DEVELOPMENT.md) for complete guide.

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

## 📚 Documentation

Complete documentation available in `/docs` directory:

- **[project-overview-pdr.md](docs/project-overview-pdr.md)** - Project vision, PDR, roadmap, and success criteria
- **[codebase-summary.md](docs/codebase-summary.md)** - Repository structure, file inventory, and service breakdown
- **[code-standards.md](docs/code-standards.md)** - Go conventions, patterns, testing, and TDD workflow
- **[system-architecture.md](docs/system-architecture.md)** - High-level architecture, database schema, auth flows, and deployment

## 📄 License

MIT

---

**Status**: 🟢 Clean skeleton ready for development
**Last Updated**: 2025-12-04
**Maintainer**: Zuno Team
