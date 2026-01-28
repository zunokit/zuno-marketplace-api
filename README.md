# Zuno NFT Marketplace API

Go microservices for NFT marketplace with gRPC services and GraphQL gateway.

## Quick Start

### Option 1: Air Hot-Reload (Development) ⚡

Fast development with hot-reload on code changes:

```bash
# 1. Install tools
make install-tools

# 2. Setup environment
cp .env.development.example .env.development
# Edit .env.development with your Neon/Upstash/CloudAMQP credentials

# 3. Start all services with hot-reload
make dev

# GraphQL Playground: http://localhost:4080/graphql
```

**Air Commands:**
| Command | Description |
|---------|-------------|
| `make dev` | Start all services |
| `make dev-auth` | Start auth-service only |
| `make dev-user` | Start user-service only |
| `make dev-wallet` | Start wallet-service only |
| `make dev-gateway` | Start graphql-gateway only |
| `make dev-stop` | Stop all Air services |
| `make dev-logs` | View service logs |


## Services

| Service         | Port (Air) | Port (Docker) | Description                |
| --------------- | ---------- | ------------- | -------------------------- |
| auth-service    | 4001       | 50051         | Authentication (SIWE, JWT) |
| user-service    | 4002       | 50052         | User profile management    |
| wallet-service  | 4003       | 50053         | Wallet operations          |
| graphql-gateway | 4080       | 8081          | GraphQL API gateway        |

## Infrastructure

### Development (Air Mode)

- **PostgreSQL:** [Neon](https://neon.tech) (serverless)
- **Redis:** [Upstash](https://upstash.com) (serverless)
- **RabbitMQ:** [CloudAMQP](https://www.cloudamqp.com) (serverless)

### Production (Docker Mode)

- All infrastructure runs in Docker containers

## Development Workflow

### Hot-Reload Development

```bash
# Terminal 1: Start services with Air
make dev

# Terminal 2: Make code changes
# Air automatically detects changes and rebuilds (< 2s)

# View logs
make dev-logs
# Or: tail -f logs/*.log
```

### Running Tests

```bash
make test              # Run all tests
make test-coverage     # Generate coverage report
make test-verbose      # Verbose output
```

### Code Quality

```bash
make lint              # Run linter
make format            # Format code
make proto             # Generate protobuf code
```

## Environment Configuration

### Development (`.env.development`)

```bash
INFRA_MODE=serverless
ENVIRONMENT=development

# Services (4xxx ports for Air)
AUTH_GRPC_PORT=:4001
USER_GRPC_PORT=:4002
WALLET_GRPC_PORT=:4003
GATEWAY_HTTP_ADDR=:4080

# Serverless infrastructure
DATABASE_URL=postgres://...@neon.tech...
REDIS_URL=redis://...@upstash.io
CLOUDAMQP_URL=amqp://...@cloudamqp.com/...
```

### Production (`infra/development/k8s/secrets.yaml`)


## Database Migrations

### Development (Serverless/Neon)

```bash
# Load environment variables first
source .env.development  # or use `make dev`

# Run migrations on Neon
make migrate-dev              # Apply pending migrations
make migrate-dev-status       # Check current version
make migrate-dev-down         # Rollback last migration
```

### Production (Docker)

```bash
# Uses Docker PostgreSQL (localhost:5433)
make migrate              # Apply pending migrations
make migrate-status       # Check current version
make migrate-down         # Rollback last migration
```

### Create New Migration

```bash
make migrate-create NAME=add_feature
# Creates: db/migrations/000003_add_feature.up.sql
#          db/migrations/000003_add_feature.down.sql
```

**Note:** Neon Serverless adds 2-3s cold start latency for first connection.

## Project Structure

```
zuno-marketplace-api/
├── services/
│   ├── auth-service/          # Authentication microservice
│   ├── user-service/          # User profile microservice
│   ├── wallet-service/        # Wallet microservice
│   └── graphql-gateway/       # GraphQL gateway
├── shared/                    # Shared code
│   ├── proto/                 # Protocol buffers
│   ├── env/                   # Shared config
│   ├── observability/         # Logging, tracing
│   ├── rabbitmq/              # RabbitMQ client
│   └── redis/                 # Redis client
├── db/migrations/             # Database migrations
├── scripts/                   # Development scripts
├── docker-compose.yml         # Production Docker setup
└── Makefile                   # Build automation
```

## Troubleshooting

### Air not working?

```bash
# Reinstall Air
go install github.com/air-verse/air@latest

# Verify installation
air version
```

### Port conflicts?

Air uses 4xxx ports (4001-4099). Docker uses 5xxx ports (50051-50053).

### Migrations failing on Neon?

```bash
# Verify DATABASE_URL format (should use pooler endpoint)
grep DATABASE_URL .env.development

# Check SSL mode is set
# Should have: ?sslmode=require

# Test database connection
make migrate-dev-status

# First migration after database sleep may timeout - retry once
```

### Services not starting?

```bash
# Check logs
make dev-logs
```

## Contributing

1. Create feature branch
2. Make changes with hot-reload (`make dev`)
3. Run tests (`make test`)
4. Submit PR

## License

MIT
