# Air Hot-Reload Research for Go Microservices

**Date:** 2026-01-28
**Researcher:** Research Agent
**Context:** Zuno Marketplace API - Docker Compose to Air migration

---

## Executive Summary

Air (cosmtrek/air) is recommended over Docker Compose for local Go microservice development due to:
- **Faster rebuild times** (native Go compilation vs container rebuilds)
- **Simpler port management** (native ports vs port forwarding)
- **Better shared code handling** (direct filesystem watching)
- **Lower resource overhead** (no container overhead)

---

## 1. Air Configuration Best Practices

### Core .air.toml Structure
```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ."
bin = "tmp/main"
include_ext = ["go", "tpl", "tmpl", "html"]
exclude_dir = ["tmp", "vendor", "build"]
delay = 1000
stop_on_root = false
rerun = false
rerun_delay = 500

[log]
time = true
color = true

[misc]
clean_on_exit = false
```

### Microservice-Specific Config
- **exclude_dir**: Always exclude `tmp`, `vendor`, `build`
- **include_dir**: Limit to specific service directories for faster builds
- **delay**: 1000ms prevents rebuild thrashing
- **stop_on_root**: false allows root file changes (go.mod, go.sum)

---

## 2. Multi-Service Setup Strategy

### Per-Service Air Instances
Each service runs Air independently:

```bash
# Terminal 1 - Auth Service
cd services/auth-service && air

# Terminal 2 - User Service
cd services/user-service && air

# Terminal 3 - Wallet Service
cd services/wallet-service && air

# Terminal 4 - GraphQL Gateway
cd services/graphql-gateway && air
```

### Process Management Options
1. **Manual terminals** (simplest, KISS principle)
2. **tmux sessions** (organized, single window)
3. **Makefile targets** (structured, automated)

**Recommendation:** Start with manual terminals, migrate to tmux if needed.

---

## 3. Shared Code Hot-Reload

### Challenge: Multiple Services Watching Same Directory

Air's filesystem watching can conflict when multiple instances watch `shared/`.

### Solutions

#### Option 1: Each Service Watches Its Shared Imports
```toml
# services/auth-service/.air.toml
include_dir = ["cmd", "../shared/auth"]
```

**Pros:** Minimal watching, fast rebuilds
**Cons:** Manual config when adding shared dependencies

#### Option 2: Watch All Shared Code
```toml
# .air.toml in service root
include_dir = ["cmd", "../shared"]
```

**Pros:** Automatic detection
**Cons:** Slower rebuilds (all shared code triggers rebuild)

**Recommendation:** Option 1 for better performance.

---

## 4. Port Management

### Current Zuno Ports
- Auth Service: 50051 (gRPC)
- User Service: 50052 (gRPC)
- Wallet Service: 50053 (gRPC)
- GraphQL Gateway: 8081 (HTTP/WS)

### Strategy
- **Hardcode ports in .air.toml** for consistency
- Use environment variables for port overrides
- No port conflicts with native execution (unlike Docker port mapping)

```toml
# services/auth-service/.air.toml
[build]
bin = "tmp/main"
args_bin = ["--port=50051"]
```

Or via environment:
```bash
export AUTH_SERVICE_PORT=50051
cd services/auth-service && air
```

---

## 5. Build Optimization

### Fast Rebuild Strategies

#### 1. Go Build Cache
```bash
export GOCACHE=$HOME/.cache/go-build
export GOMODCACHE=$HOME/go/pkg/mod
```
- Cached between rebuilds
- Faster incremental compilation

#### 2. Minimal Binary Size
```toml
cmd = "go build -ldflags '-s -w' -o ./tmp/main ."
```
- `-s`: Strip symbol table
- `-w`: Strip DWARF debug info
- 30-40% smaller binaries

#### 3. Build Tags for Conditional Compilation
```bash
# Development builds skip production features
cmd = "go build -tags dev -o ./tmp/main ."
```

#### 4. Parallel Builds (Multiple Services)
Each Air instance builds independently. No bottleneck.

---

## 6. Migration Strategy from Docker Compose

### Phase 1: Add Air Alongside Docker
- Keep Docker Compose for infra services (Postgres, Redis, RabbitMQ)
- Add Air for app services only
- Validate parity

### Phase 2: Create Air-Specific Makefile Targets
```makefile
dev-air: ## Start infra with Air hot-reload
	docker compose up -d postgres redis rabbitmq
	@echo "Starting services with Air..."
	@make -j4 dev-air-auth dev-air-user dev-air-wallet dev-air-gateway

dev-air-auth:
	cd services/auth-service && air

# ... repeat for other services
```

### Phase 3: Sunset Docker Compose for App Services
- Remove app service containers from docker-compose.yml
- Keep only infra services
- Document Air-only workflow

---

## 7. Air vs Tilt (Existing Setup)

**Current State:** Tiltfile.development exists with K8s deployment

**Comparison:**
| Feature | Air | Tilt |
|---------|-----|------|
| Setup Complexity | Low | High (K8s required) |
| Rebuild Speed | Fast (native) | Medium (Docker build) |
| Learning Curve | Minimal | Steep |
| Port Management | Native | Port forwarding |
| Best For | Local dev | K8s parity |

**Recommendation:** Use Air for local dev, keep Tilt for K8s testing.

---

## 8. Implementation Recommendations

### Immediate Actions
1. **Create .air.toml templates** for each service
2. **Add Makefile targets** for Air-based dev workflow
3. **Document tmux session** setup for running 4 services
4. **Test shared code watching** with multiple Air instances

### File Structure
```
services/
  auth-service/
    .air.toml
  user-service/
    .air.toml
  wallet-service/
    .air.toml
  graphql-gateway/
    .air.toml
Makefile (add dev-air targets)
```

### Success Metrics
- Rebuild time < 5 seconds per service
- No shared code watching conflicts
- All services run without port conflicts
- Developer onboarding < 5 minutes

---

## Unresolved Questions

1. **Filesystem watcher limits:** Can Windows handle 4 Air instances watching `shared/` simultaneously? Test required.
2. **gRPC port conflicts:** Verify services don't try binding same port during hot-reload restart.
3. **Environment variable loading:** How to load `.env` files per service with Air?
4. **Integration testing:** How to run e2e tests against Air-spawned services?

---

## Sources

**Web Search Limitation Encountered:** Search tools reached monthly quota (resets 2026-02-01).

**Knowledge Sources:**
- **Air GitHub Repository:** https://github.com/air-verse/air (primary documentation)
- **Go Build Cache:** https://go.dev/ref/mod#build-commands
- **Zuno Codebase Analysis:** Makefile, Tiltfile.development, go.mod

**Manual Verification Required:**
Test multi-instance Air watching shared directories on Windows before full migration.

---

**Next Steps:**
1. Create proof-of-concept with 2 services watching shared code
2. Measure rebuild times vs current Docker Compose
3. Document migration guide for team
