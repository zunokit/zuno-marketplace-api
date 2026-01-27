# Research: Hybrid Docker + Local Hot-Reload Development Workflow

**Date:** 2026-01-28
**Focus:** Environment management, service orchestration, infrastructure dependencies, scripting patterns, Windows/WSL2 compatibility

---

## Executive Summary

Hybrid development workflows combine Docker for production-like infrastructure with local hot-reload for rapid iteration. Key patterns include volume mounting, multi-stage builds, and environment-specific orchestration scripts.

---

## 1. Environment Management

### Current State (Codebase)
- **Serverless mode**: External services (Neon/Upstash/CloudAMQP) via `.env.development.example`
- **Docker mode**: Local containers via `.env.production.example` + `docker-compose.yml`
- **Setup script**: `scripts/setup-env.sh` offers interactive mode selection

### Best Practices
1. **Separate compose files** for dev/prod (`docker-compose.dev.yml` + override files)
2. **Environment variable precedence**: CLI args > `.env.local` > `.env` > defaults
3. **Feature flags** for switching between local/remote services per environment
4. **Validation scripts** to verify required environment variables before startup

### Key Insight
**YAGNI Principle**: Current dual-mode approach is sufficient. No need for complex env management until multiple deployment targets exist.

---

## 2. Service Orchestration

### Current State
- **Docker Compose**: Orchestrates 7 services (postgres, redis, rabbitmq, 3 gRPC services, graphql-gateway)
- **Makefile**: Single `make dev` command handles startup, dependencies, migrations
- **Health checks**: `scripts/health-check.sh` validates service readiness

### Recommended Patterns
1. **Dependency-aware startup**: Use `depends_on` + healthcheck conditions
2. **Hot-reload integration**: Volume mount source code, run `air` in container
3. **Parallel vs sequential start**: Infrastructure services in parallel, apps with proper ordering
4. **Service discovery**: Docker network DNS vs hardcoded localhost ports

### Hybrid Approach
```yaml
# docker-compose.dev.yml (proposed)
services:
  auth-service:
    volumes:
      - ./services/auth-service:/app
    command: air -c .air.toml  # Hot reload inside container
```

**Alternative**: Run some services locally (for hot-reload) + infrastructure in Docker.

---

## 3. Infrastructure Dependencies

### Current Integration
- **Neon PostgreSQL**: Serverless Postgres with branching (free tier)
- **Upstash Redis**: Serverless Redis with edge caching
- **CloudAMQP RabbitMQ**: Managed message queue (free tier available)

### Hybrid Strategy
| Service | Dev Mode | Production |
|---------|----------|------------|
| Database | Local Docker (fast iterations) | Neon (serverless, scalable) |
| Cache | Local Docker | Upstash (global edge) |
| Queue | Local Docker | CloudAMQP (managed) |

### Benefits
- **Fast local feedback**: No network latency, offline development
- **Production parity**: Test migrations/schema changes locally
- **Cost optimization**: Use free managed tiers in production, avoid local resource drain

---

## 4. Scripting Patterns

### Current Makefile Strengths
- ✅ Cross-platform Windows/Unix support (lines 143-154)
- ✅ Single-command startup (`make dev`)
- ✅ Automated migrations on startup
- ✅ Clear help target

### Recommended Additions
```makefile
# Hot reload targets
dev-hot:
	@echo "Starting with hot reload..."
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up

# Local-only mode (no Docker)
dev-local:
	@echo "Starting services locally with Air..."
	cd services/auth-service/cmd && air & \
	cd services/user-service/cmd && air & \
	cd services/graphql-gateway/cmd && air

# Environment validation
check-env:
	@scripts/validate-env.sh
```

### Script Organization
```
scripts/
├── setup-env.sh          # ✅ Exists
├── validate-env.sh       # New: check required vars
├── start-dev.sh          # New: dev mode selector
├── health-check.sh       # ✅ Exists
└── migrate.sh            # New: migration wrapper
```

---

## 5. Windows/WSL2 Compatibility

### Current State
- **Makefile**: Has Windows-specific commands (`if exist`, `rmdir /s /q`)
- **Docker**: Works with Docker Desktop WSL2 backend
- **Scripts**: Bash scripts (require WSL/Git Bash on Windows)

### Recommendations
1. **Dual script support**: Provide `.bat`/`.ps1` equivalents for core scripts
2. **Path handling**: Use `/mnt/e/...` for WSL2 access to Windows filesystem
3. **Docker context**: Ensure `docker context` points to WSL2 backend
4. **Air configuration**: Windows binary compatibility in `.air.toml`

### WSL2-Specific Considerations
- **Filesystem performance**: Keep code in WSL2 filesystem (`/home/...`), not `/mnt/c/`
- **Line endings**: `git autocrlf` handling for shell scripts
- **Permissions**: Executable bits lost on Windows filesystem

---

## Key Findings

### What Works Well
1. **Existing Makefile** provides excellent cross-platform build support
2. **Setup script** demonstrates good env management UX
3. **Docker Compose** properly handles service dependencies

### Gaps Identified
1. **No hot-reload support**: Currently rebuilds containers on every change
2. **No local-only mode**: All services must run in Docker
3. **Windows scripts missing**: Bash-only limits Windows adoption
4. **No dev/prod compose separation**: Single file for all environments

### Recommended Migration Path
1. **Phase 1**: Add Air hot-reload to existing Docker setup
2. **Phase 2**: Create `docker-compose.dev.yml` with volume mounts
3. **Phase 3**: Add `make dev-hot` target for hot-reload mode
4. **Phase 4**: Create Windows PowerShell equivalents for scripts

---

## Sources

**Note**: Web search unavailable due to rate limit. Findings based on:
- Current codebase analysis (Makefile, docker-compose.yml, Dockerfiles, scripts)
- Docker Compose documentation patterns
- Go Air hot-reload best practices
- WSL2 + Docker Desktop integration patterns

**Unresolved Questions**:
- Exact Air + Docker volume mounting pattern for Go microservices?
- WSL2 filesystem performance benchmarks for code vs /mnt?
- Optimal strategy for switching between local/remote Neon DB?
- Should Makefile or bash scripts handle service startup orchestration?

---

**Report End**
