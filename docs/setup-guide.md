# Setup Guide

**Version**: 1.0.0 | **Last Updated**: 2025-12-29

---

## Quick Start

### Prerequisites

| Tool | Version | Required For |
|------|---------|--------------|
| Go | 1.21+ | Building services |
| Docker Desktop | Latest | Infrastructure services |
| Make | Any | Build automation |
| Git | Any | Version control |

### Windows Users

Install **Git Bash** or **WSL2** for full Makefile support. The Makefile uses Unix commands (`sleep`, `find`, `mkdir -p`).

---

## 1. Clone Repository

```bash
git clone https://github.com/zunokit/zuno-marketplace-api.git
cd zuno-marketplace-api
git checkout feature/add-sentry  # or main/develop-claude
```

---

## 2. Infrastructure Setup (Docker Compose)

### Start All Services

```bash
make dev
```

This command:
1. Stops and removes old containers
2. Starts PostgreSQL, Redis, RabbitMQ
3. Waits for PostgreSQL to be ready
4. Runs database migrations

### Verify Services

```bash
# Check all services running
docker compose ps

# Expected output:
# NAME           STATUS          PORTS
# postgres       Up              0.0.0.0:5433->5432/tcp
# redis          Up              0.0.0.0:6379->6379/tcp
# rabbitmq       Up              0.0.0.0:5672->5672/tcp, 0.0.0.0:15672->15672/tcp
```

### Service URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| GraphQL Playground | http://localhost:8081/graphql | - |
| PostgreSQL | localhost:5433 | postgres/postgres |
| RabbitMQ Management | http://localhost:15672 | guest/guest |

---

## 3. Environment Configuration

### Create .env File

```bash
cp .env.example .env
```

### Edit .env

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

# Sentry (Optional - for error tracking)
SENTRY_DSN=https://your-dsn@sentry.io/project-id
SENTRY_ENVIRONMENT=development
SENTRY_TRACES_SAMPLE_RATE=0.1
```

---

## 4. Build Services

```bash
# Build all services
make build

# Build individual service
make build-auth
make build-user
make build-wallet
make build-gateway

# Show version info
make build-version
```

Binaries are created in `./build/`:
- `build/auth-service`
- `build/user-service`
- `build/wallet-service`
- `build/graphql-gateway`

### Version Information

Builds inject version info via ldflags:

```bash
$ make build-version
Version: 9f70046-dirty
BuildTime: 2025-12-29T12:00:00Z
LDFLAGS: -ldflags -X main.Version=9f70046-dirty -X main.BuildTime=2025-12-29T12:00:00Z
```

---

## 5. Run Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run with verbose output
make test-verbose
```

---

## 6. CI/CD Setup (Sentry Integration)

### GitHub Secrets Configuration

Go to: **Repository → Settings → Secrets and variables → Actions → New repository secret**

Add the following secrets:

| Secret Name | Value | How to Get |
|------------|-------|------------|
| `SENTRY_AUTH_TOKEN` | Your auth token | https://sentry.io/settings/auth-tokens/ (scopes: `project:releases`, `project:write`) |
| `SENTRY_ORG` | `zuno` | Your Sentry organization slug |
| `SENTRY_PROJECT` | `zuno-marketplace-api` | Your Sentry project slug |

### CI/CD Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | Push to `main`, `develop-claude`, PR | Lint, test, build, Sentry release |
| `deploy-staging.yml` | Push to `main`, `develop-claude`, `feature/add-sentry` | Deploy to staging with Sentry tracking |
| `pr.yml` | Pull request | PR quality checks |

### Test CI/CD Locally

```bash
# Trigger workflow by pushing
git push origin feature/add-sentry

# Check GitHub Actions tab for workflow run
```

---

## 7. Development Workflow

### TDD Cycle

```bash
# 1. Write failing test
# 2. Run test (should fail)
make test

# 3. Write minimal code to pass
# 4. Run test (should pass)
make test

# 5. Refactor
# 6. Run tests again
make test
```

### Code Quality

```bash
# Format code
make format

# Run linter
make lint

# Run full CI locally
make ci
```

---

## 8. Common Commands

```bash
make help              # Show all commands
make dev               # Start infrastructure
make dev-stop          # Stop infrastructure
make dev-logs          # View logs
make dev-clean         # Remove all data
make test              # Run tests
make build             # Build services
make migrate           # Run migrations
make proto             # Generate protobuf
make lint              # Run linter
make format            # Format code
make install-tools     # Install dev tools
```

---

## 9. Troubleshooting

### Port Already in Use

```bash
# Check what's using port 5433
netstat -ano | findstr :5433  # Windows
lsof -i :5433                  # macOS/Linux

# Change port in .env
POSTGRES_PORT=5434
```

### Database Connection Failed

```bash
# Verify PostgreSQL running
docker compose ps postgres

# Check logs
docker compose logs postgres

# Restart infrastructure
make dev-clean
make dev
```

### Migration Errors

```bash
# Check migration status
make migrate-status

# Rollback last migration
make migrate-down

# Re-run migrations
make migrate
```

### Makefile Commands Fail on Windows

Use **Git Bash** or **WSL2** instead of PowerShell/CMD:

```bash
# Open Git Bash and run
cd /e/zuno-marketplace-api
make build
```

---

## 10. Production Deployment

See [Deployment Guide](deployment-guide.md) for:
- Production environment setup
- Sentry release tracking
- Deploy notifications
- Rollback procedures
- Monitoring and alerts

---

## Next Steps

1. **Configure Sentry**: Add `SENTRY_DSN` to `.env` for error tracking
2. **Explore GraphQL**: Visit http://localhost:8081/graphql
3. **Run Integration Tests**: `make test`
4. **Set Up CI/CD**: Add GitHub Secrets for Sentry integration

---

**Related Docs**:
- [Deployment Guide](deployment-guide.md)
- [System Architecture](system-architecture.md)
- [Code Standards](code-standards.md)
- [Project Overview](project-overview-pdr.md)
