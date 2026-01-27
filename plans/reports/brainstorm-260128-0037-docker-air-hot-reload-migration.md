# Brainstorm: Docker + Air Hot-Reload Migration

## Context & Requirements

**Current State:**
- ✅ Full Docker Compose (postgres/redis/rabbitmq in containers)
- ✅ Multi-stage Dockerfiles per service
- ❌ Old Tilt setup (Tiltfile.development, Tiltfile.production) - needs cleanup
- ❌ No hot-reload for local dev

**Target State:**
- **Production:** Docker (current docker-compose.yml)
- **Development:** Local Air + serverless infra (Neon/Upstash/CloudAMQP)
- **Hot-reload:** Critical for shared/ code changes
- **Ports:** 4000-4099 range for local services

**User Requirements:**
1. Comprehensive setup (not partial)
2. Port range 4xxx for local dev
3. Integrate with existing `make dev` command
4. Cleanup old Tilt/serverless artifacts

---

## Cleanup: Remove Old Serverless Artifacts

**Files to DELETE:**
```
Tiltfile.development          # Old Tift config
Tiltfile.production          # Old Tift config
infra/development/k8s/       # K8s manifests (if exists)
.git/refs/heads/feature/tilt-serverless-configs  # Old branch
```

**Files to KEEP:**
```
docker-compose.yml           # Production
.env.development.example     # Update for Air
.env.production.example      # Keep as-is
```

---

## Architecture: Local Air + Serverless Infra

### Infrastructure (Cloud/External)
```
┌─────────────────────────────────────────────────────┐
│  External Services (Free Tier)                      │
├─────────────────────────────────────────────────────┤
│  PostgreSQL:  Neon (cloud)                          │
│  Redis:       Upstash (cloud)                       │
│  RabbitMQ:    CloudAMQP (cloud)                     │
└─────────────────────────────────────────────────────┘
```

### Local Services (Air Hot-Reload)
```
┌─────────────────────────────────────────────────────┐
│  Local Development (Port 4000-4099)                 │
├─────────────────────────────────────────────────────┤
│  auth-service:       localhost:4001 (gRPC)          │
│  user-service:       localhost:4002 (gRPC)          │
│  wallet-service:     localhost:4003 (gRPC)          │
│  graphql-gateway:    localhost:4080 (HTTP)          │
└─────────────────────────────────────────────────────┘
```

### Directory Structure
```
zuno-marketplace-api/
├── services/
│   ├── auth-service/
│   │   ├── .air.toml              # Air config
│   │   └── cmd/main.go
│   ├── user-service/.air.toml
│   ├── wallet-service/.air.toml
│   └── graphql-gateway/.air.toml
├── scripts/
│   ├── dev-air.sh                 # Run Air services
│   ├── watch-shared.sh            # Watch shared/ code
│   └── setup-env.sh               # Setup .env from example
├── .env.development               # Local dev config
├── Makefile (enhanced)
└── docker-compose.yml             # Production only
```

---

## Air Configuration (.air.toml)

### Template (per service)
```toml
# services/auth-service/.air.toml
root = "."
tmp_dir = "tmp"
build_delay = 1000  # ms

[build]
  cmd = "go build -o ./tmp/main ./cmd/main.go"
  bin = "tmp/main"
  include_ext = ["go"]
  exclude_dir = ["tmp", "vendor", "build"]
  include_dir = ["cmd", "internal", "shared"]
  exclude_file = []
  exclude_unchanged = false
  follow_symlink = false
  full_delay = 2000
  stop_on_error = true

[log]
  time = true
  main_only = false

[color]
  main = "magenta"
  builder = "yellow"
  runner = "green"

[misc]
  clean_on_exit = false
```

---

## Environment Configuration

### .env.development (Local Air)
```bash
# Infrastructure Mode
INFRA_MODE=serverless

ENVIRONMENT=development
PORT=4080

# Auth
JWT_SECRET=dev-secret-key-32-chars-min
REFRESH_SECRET=dev-refresh-secret-key-32-chars

# Database (Neon)
DATABASE_URL=postgres://postgres:[PASSWORD]@[PROJECT-ID].aws.neon.tech/neondb?sslmode=require

# Fallback: Docker infra
POSTGRES_HOST=localhost
POSTGRES_PORT=5433
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DATABASE=nft_marketplace

# Redis (Upstash)
REDIS_URL=redis://default:[PASSWORD]@xxx.upstash.io:6379

# Fallback: Docker infra
REDIS_HOST=localhost
REDIS_PORT=6379

# RabbitMQ (CloudAMQP)
CLOUDAMQP_URL=amqp://xxx:xxx@xxx.rmq.cloudamqp.com/xxx

# Fallback: Docker infra
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest

# Local Ports (4xxx range)
AUTH_GRPC_PORT=:4001
USER_GRPC_PORT=:4002
WALLET_GRPC_PORT=:4003
GATEWAY_HTTP_ADDR=:4080

# Service URLs (localhost for local dev)
USER_SERVICE_URL=localhost:4002
WALLET_SERVICE_URL=localhost:4003
```

---

## Enhanced Makefile

```makefile
.PHONY: help dev dev-stop dev-air dev-air-all

# ============================================================
# Development Modes
# ============================================================

help: ## Show help
	@echo ============================================================
	@echo   Zuno NFT Marketplace - Development
	@echo ============================================================
	@echo
	@echo Production Mode (Docker):
	@echo   make dev           - Full Docker with infra containers
	@echo   make dev-stop      - Stop Docker
	@echo
	@echo Development Mode (Air + Serverless):
	@echo   make dev-air       - Start infra + all services with Air
	@echo   make dev-air-auth  - Start only auth-service with Air
	@echo   make dev-air-stop  - Stop all Air services
	@echo   make infra-up      - Start only infra (postgres/redis/rabbitmq)
	@echo   make infra-down    - Stop infra
	@echo
	@echo Common:
	@echo   make test          - Run tests
	@echo   make build         - Build binaries
	@echo ============================================================

# ============================================================
# Docker Mode (Production-like)
# ============================================================

dev: ## Start Docker Compose (production mode)
	docker compose up -d
	@sleep 8
	$(MAKE) migrate
	@echo Ready! GraphQL: http://localhost:8081/graphql

dev-stop: ## Stop Docker Compose
	docker compose down

# ============================================================
# Air Mode (Development with hot-reload)
# ============================================================

dev-air: ## Start infra + all services with Air
	@echo ============================================================
	@echo   Starting Air Development Environment...
	@echo ============================================================
	@$(MAKE) infra-up
	@sleep 3
	@./scripts/dev-air.sh all

dev-air-auth: dev-air ## Alias for dev-air

dev-air-stop: ## Stop Air services
	@echo Stopping Air services...
	@pkill -f "air" || true
	@echo Stopped!

dev-air-logs: ## View Air logs
	@tail -f services/*/tmp/stdout.log

# ============================================================
# Infrastructure (can be Docker or cloud)
# ============================================================

infra-up: ## Start infrastructure services
	docker compose up -d postgres redis rabbitmq
	@echo Waiting for infra...
	@sleep 5
	@echo Infra ready!

infra-down: ## Stop infrastructure
	docker compose down

# ============================================================
# Existing targets
# ============================================================
test:
	go test ./...

build:
	go build ./...
```

---

## Scripts

### scripts/dev-air.sh
```bash
#!/bin/bash
# Run services with Air

SERVICES_DIR="services"
LOGS_DIR="logs"

mkdir -p "$LOGS_DIR"

# Function to run service with Air
run_service() {
  local service=$1
  local port=$2

  echo "Starting $service on port $port..."

  cd "$SERVICES_DIR/$service"
  air > "$LOGS_DIR/$service.log" 2>&1 &
  echo $! > "$LOGS_DIR/$service.pid"
  cd ../..

  sleep 1
}

# Parse arguments
SERVICES="${1:-all}"

case $SERVICES in
  all)
    run_service "auth-service" 4001
    run_service "user-service" 4002
    run_service "wallet-service" 4003
    run_service "graphql-gateway" 4080
    ;;
  auth)
    run_service "auth-service" 4001
    ;;
  user)
    run_service "user-service" 4002
    ;;
  wallet)
    run_service "wallet-service" 4003
    ;;
  gateway)
    run_service "graphql-gateway" 4080
    ;;
esac

echo ============================================================
echo   Services running with Air hot-reload!
echo ============================================================
echo "  auth-service:    localhost:4001"
echo "  user-service:    localhost:4002"
echo "  wallet-service:  localhost:4003"
echo "  graphql-gateway: localhost:4080"
echo ============================================================
echo "Logs: ./logs/"
echo "Stop: make dev-air-stop"
echo ============================================================
```

### scripts/watch-shared.sh
```bash
#!/bin/bash
# Watch shared/ and proto/ for changes, restart dependent services

echo "Watching shared/ and proto/ for changes..."

while true; do
  # Watch for changes in shared/ or proto/
  changed=$(find shared proto -name "*.go" -type f 2>/dev/null | entr -dp -n echo "changed")

  if [ "$changed" == "changed" ]; then
    echo "$(date): Shared code changed, restarting services..."
    pkill -f "air" || true
    sleep 2
    ./scripts/dev-air.sh all
  fi

  sleep 2
done
```

---

## Implementation Phases

### Phase 1: Cleanup (30 min)
- [ ] Delete Tiltfile.development, Tiltfile.production
- [ ] Delete infra/development/k8s/ if exists
- [ ] Clean up git refs to old tilt branch
- [ ] Update README to remove Tilt references

### Phase 2: Air Configuration (1 hour)
- [ ] Install Air: `go install github.com/air-verse/air@latest`
- [ ] Create .air.toml for auth-service
- [ ] Create .air.toml for user-service
- [ ] Create .air.toml for wallet-service
- [ ] Create .air.toml for graphql-gateway
- [ ] Test hot-reload on single service

### Phase 3: Environment Setup (1 hour)
- [ ] Update .env.development.example with 4xxx ports
- [ ] Add INFRA_MODE=serverless flag
- [ ] Document Neon/Upstash/CloudAMQP setup
- [ ] Create setup-env.sh script

### Phase 4: Scripts & Makefile (2 hours)
- [ ] Create scripts/dev-air.sh
- [ ] Create scripts/watch-shared.sh
- [ ] Update Makefile with dev-air targets
- [ ] Add infra-up/infra-down targets
- [ ] Test `make dev-air` command

### Phase 5: Integration & Testing (2 hours)
- [ ] Test multi-service hot-reload
- [ ] Test shared/ code changes
- [ ] Test inter-service communication
- [ ] Verify infra fallback (Docker vs cloud)
- [ ] Update README with new workflow

### Phase 6: Documentation (1 hour)
- [ ] Update README.md
- [ ] Document Neon/Upstash/CloudAMQP setup
- [ ] Add troubleshooting section
- [ ] Create quick-start guide

---

## Success Criteria

✅ **Hot-reload works:** Code changes trigger rebuild in <2s
✅ **Shared code reloads:** Changes in shared/ restart all services
✅ **Single command:** `make dev-air` starts everything
✅ **Clean workflow:** No leftover Tilt/K8s files
✅ **Port consistency:** All services use 4xxx range
✅ **Production unchanged:** docker-compose.yml still works

---

## Risks & Mitigation

| Risk | Mitigation |
|------|------------|
| Air not installed | Add to install-tools in Makefile |
| Port conflicts on 4xxx | Document in .env, make configurable |
| Cloud service downtime | Provide Docker fallback in docker-compose |
| Shared code race conditions | Use debounce in watch script |
| Windows compatibility | Test Air on Windows, use WSL2 if needed |

---

## Next Steps

1. **Confirm this approach** - Air local + serverless infra
2. **Set up cloud accounts** - Neon/Upstash/CloudAMQP free tiers
3. **Create implementation plan** - Detailed step-by-step

**Estimated effort:** 1-2 days for full setup + testing

---

## Unresolved Questions

1. **Preferred cloud services:** Already have Neon/Upstash/CloudAMQP accounts?
2. **Fallback strategy:** If cloud services down, use Docker infra or skip service?
3. **Team size:** Multiple devs need this, or solo project?
4. **CI/CD:** Need Air mode in CI, or only Docker?

Ready to create implementation plan?
