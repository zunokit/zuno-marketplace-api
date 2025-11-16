# Quick Start Guide

## 🚀 Recommended: Docker Compose

**One command to start everything:**

```bash
make dev
```

That's it! This will:
1. ✅ Start all services (PostgreSQL, Redis, RabbitMQ, API services)
2. ✅ Run database migrations automatically
3. ✅ Show you where to access everything

### Access Services

- **GraphQL Playground**: http://localhost:8081/graphql
- **PostgreSQL**: localhost:5433
- **RabbitMQ UI**: http://localhost:15672 (guest/guest)

### View Logs

```bash
make dev-logs
```

### Stop Everything

```bash
make dev-stop
```

### Start Fresh (Delete All Data)

```bash
make dev-clean
make dev
```

---

## 🔧 Common Commands

### Database

```bash
make migrate              # Run migrations
make migrate-status       # Check migration version
make migrate-create NAME=add_users  # Create new migration
make migrate-down         # Rollback last migration
```

### Testing

```bash
make test                 # Run all tests
make test-coverage        # Run tests with coverage report
```

### Building

```bash
make build                # Build all services
make proto                # Generate protobuf code
```

### Code Quality

```bash
make format               # Format code
make lint                 # Run linter
```

---

## 🆘 Troubleshooting

### Services won't start

```bash
# Clean everything and start fresh
make dev-clean
make dev
```

### Port already in use

```bash
# Stop all Docker containers
docker compose down

# Or check what's using the port
netstat -ano | findstr :8081
netstat -ano | findstr :5433
```

### Migration errors

```bash
# Check migration status
make migrate-status

# Rollback if needed
make migrate-down
```

---

## 📦 Services

| Service | Type | Port | Description |
|---------|------|------|-------------|
| GraphQL Gateway | HTTP/WS | 8081 | API Gateway |
| Auth Service | gRPC | 50051 | Authentication |
| User Service | gRPC | 50052 | User management |
| Wallet Service | gRPC | 50053 | Wallet operations |
| PostgreSQL | Database | 5433 | Main database |
| Redis | Cache | 6379 | Cache & sessions |
| RabbitMQ | Message Queue | 5672 | Event bus |

---

## 🔥 Advanced: Tilt/Kubernetes

For hot reload and production-like environment, see [TILT.md](TILT.md)

**Only use Tilt if:**
- You need hot reload (code changes instantly reflected)
- You want to test Kubernetes deployment
- You're comfortable with Kubernetes

**For most development, stick with `make dev`!**
