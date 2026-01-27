---
title: "Phase 05: Makefile Enhancement"
description: "Add Air hot-reload targets to Makefile"
status: pending
priority: P1
effort: 1h
branch: feature/hybird-serverless-and-servers-infrastructure
tags: [makefile, automation, air, docker]
created: 2026-01-28
---

# Phase 05: Makefile Enhancement

## Context Links

- [Main Plan](plan.md)
- [Hybrid Workflow Research](research/researcher-02-hybrid-workflow.md)
- [Phase 04: Scripts Creation](phase-04-scripts-creation.md)
- [Existing Makefile](../../Makefile)

## Overview

**Priority:** P1 (developer experience)
**Current Status:** Pending
**Description:** Enhance Makefile with Air hot-reload targets while preserving existing Docker workflow.

**New Targets:**
- `make dev-air` - Start all services with Air
- `make dev-air-stop` - Stop Air services
- `make infra-up` - Start infra containers (postgres/redis/rabbitmq)
- `make infra-down` - Stop infra containers
- `make validate-env` - Validate environment

**Preserved Targets:**
- `make dev` - Full Docker (unchanged)
- `make dev-stop` - Stop Docker (unchanged)
- All existing build/test targets (unchanged)

---

## Key Insights

From [Hybrid Workflow Research](research/researcher-02-hybrid-workflow.md):
- **Existing Makefile** provides excellent cross-platform support
- **Windows commands** properly handled (if/else for OS detection)
- **Single-command startup** is key to good developer experience
- **Help target** should clearly separate Docker vs Air modes

**Design Principles:**
- Keep existing `make dev` workflow unchanged
- Add Air targets as alternatives
- Clear separation in help output
- Use scripts from Phase 04

---

## Requirements

### Functional Requirements
1. Add `dev-air` target to start all services with Air
2. Add `dev-air-stop` target to stop Air services
3. Add `infra-up`/`infra-down` for infra containers
4. Add `validate-env` for environment validation
5. Update `help` target with new commands
6. Add Air to `install-tools` target

### Non-Functional Requirements
- Backward compatible (existing targets unchanged)
- Cross-platform (Windows/Unix)
- Clear documentation in help output
- No breaking changes to CI/CD

---

## Architecture

### Makefile Structure (After)

```
Makefile
├── Configuration (existing)
├── Help (UPDATED - show both modes)
├── Docker Mode (existing - unchanged)
│   ├── dev (full docker)
│   ├── dev-stop
│   └── dev-logs
├── Air Mode (NEW)
│   ├── dev-air (start with Air)
│   ├── dev-air-stop
│   ├── dev-air-logs
│   └── validate-env
├── Infrastructure (NEW)
│   ├── infra-up (docker infra only)
│   └── infra-down
├── Migrations (existing)
├── Testing (existing)
├── Building (existing)
├── Code Generation (existing)
└── Tools (UPDATED - add Air)
```

### Workflow Comparison

**Docker Mode (Existing):**
```bash
make dev          # One command starts everything
make dev-stop     # Stop everything
```

**Air Mode (New):**
```bash
make validate-env # Check environment first
make infra-up     # Start infra (optional for serverless)
make dev-air      # Start services with Air
make dev-air-stop # Stop Air services
make infra-down   # Stop infra (optional)
```

---

## Related Code Files

### Files to MODIFY
- `Makefile` - Add new targets, update help

### Files to REFERENCE
- `scripts/dev-air.sh` (from Phase 04)
- `scripts/stop-air.sh` (from Phase 04)
- `scripts/validate-env.sh` (from Phase 04)
- `docker-compose.yml` (for infra-up target)

---

## Implementation Steps

### Step 1: Read Current Makefile (5 min)

```bash
# Read existing Makefile
cat Makefile

# Note:
# - Existing targets structure
# - Windows compatibility code
# - Help target format
```

**Success:** Understand current Makefile structure

**Rollback:** None (read-only)

---

### Step 2: Add Air Tool to install-tools (5 min)

Update `install-tools` target:

```makefile
install-tools: ## Install development tools
	@echo Installing development tools...
	go install github.com/air-verse/air@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo Tools installed!
```

**Change:** Add `air` installation as first tool

**Success:** `make install-tools` installs Air

**Rollback:** Remove Air line

---

### Step 3: Update Help Target (15 min)

Replace help target with new version:

```makefile
help: ## Show help
	@echo ============================================================
	@echo   Zuno NFT Marketplace - Development
	@echo ============================================================
	@echo
	@echo MODE 1 - Docker Compose (Production-like):
	@echo   make dev           - Start all services in Docker
	@echo   make dev-stop      - Stop Docker services
	@echo   make dev-logs      - View Docker logs
	@echo   make dev-clean     - Stop and remove all data
	@echo
	@echo MODE 2 - Air Hot-Reload (Development):
	@echo   make validate-env  - Validate .env.development
	@echo   make dev-air       - Start services with Air (all 4)
	@echo   make dev-air-stop  - Stop Air services
	@echo   make dev-air-logs  - View Air logs
	@echo
	@echo Infrastructure (for Air mode):
	@echo   make infra-up      - Start infra containers (postgres/redis/rabbitmq)
	@echo   make infra-down    - Stop infra containers
	@echo
	@echo Common Commands:
	@echo   make test          - Run tests
	@echo   make test-coverage - Run tests with coverage
	@echo   make build         - Build all services
	@echo   make migrate       - Run database migrations
	@echo   make proto         - Generate protobuf
	@echo   make lint          - Run linter
	@echo   make format        - Format code
	@echo
	@echo Tools:
	@echo   make install-tools - Install development tools (includes Air)
	@echo ============================================================
```

**Success:** Help shows both modes clearly

**Rollback:** Restore original help target

---

### Step 4: Add Air Mode Targets (20 min)

Add after Docker Mode section:

```makefile
# ============================================================
# Air Hot-Reload Mode (Development)
# ============================================================

validate-env: ## Validate environment configuration
	@echo Validating environment...
	@./scripts/validate-env.sh

dev-air: ## Start all services with Air hot-reload
	@echo ============================================================
	@echo   Starting Air Development Environment...
	@echo ============================================================
	@./scripts/validate-env.sh
	@./scripts/dev-air.sh all

dev-air-auth: ## Start only auth-service with Air
	@./scripts/dev-air.sh auth

dev-air-user: ## Start only user-service with Air
	@./scripts/dev-air.sh user

dev-air-wallet: ## Start only wallet-service with Air
	@./scripts/dev-air.sh wallet

dev-air-gateway: ## Start only graphql-gateway with Air
	@./scripts/dev-air.sh gateway

dev-air-stop: ## Stop all Air services
	@echo Stopping Air services...
	@./scripts/stop-air.sh

dev-air-logs: ## View Air service logs
	@echo Opening logs (Ctrl+C to exit)...
	@tail -f logs/*.log 2>/dev/null || echo "No logs found. Start services first with 'make dev-air'"

dev-air-status: ## Show status of Air services
	@echo Air Service Status:
	@echo
	@if [ -d "logs" ]; then \
		for pid in logs/*.pid; do \
			if [ -f "$$pid" ]; then \
				service=$$(basename "$$pid" .pid); \
				pid=$$(cat "$$pid"); \
				if ps -p "$$pid" > /dev/null 2>&1; then \
					echo "  ✓ $$service (PID: $$pid)"; \
				else \
					echo "  ✗ $$service (not running)"; \
				fi; \
			fi; \
		done; \
	else \
		echo "  No services started"; \
	fi
```

**Success:** All Air targets work

**Rollback:** Remove this section

---

### Step 5: Add Infrastructure Targets (10 min)

Add after Air Mode section:

```makefile
# ============================================================
# Infrastructure (for Air mode)
# ============================================================

infra-up: ## Start infrastructure services (postgres/redis/rabbitmq)
	@echo Starting infrastructure...
	docker compose up -d postgres redis rabbitmq
	@echo Waiting for services to be ready...
	@sleep 5
	@echo Infrastructure ready!
	@echo PostgreSQL: localhost:5433
	@echo Redis:      localhost:6379
	@echo RabbitMQ:   localhost:5672 (UI: http://localhost:15672)

infra-down: ## Stop infrastructure services
	@echo Stopping infrastructure...
	docker compose down
	@echo Infrastructure stopped!

infra-logs: ## View infrastructure logs
	docker compose logs -f postgres redis rabbitmq

infra-status: ## Show infrastructure status
	@echo Infrastructure Status:
	@echo
	@docker compose ps postgres redis rabbitmq
```

**Success:** Infra targets work

**Rollback:** Remove this section

---

### Step 6: Add Environment Validation (5 min)

Add to Configuration section (after database config):

```makefile
# Environment configuration
ENV_FILE ?= .env.development
INFRA_MODE ?= serverless
```

**Success:** Environment vars configurable

**Rollback:** Remove lines

---

### Step 7: Test All New Targets (15 min)

```bash
# Test help
make help

# Test env validation
make validate-env

# Test infra targets (if using Docker infra)
make infra-up
make infra-status
make infra-logs
make infra-down

# Test Air targets
make dev-air-auth
sleep 5
make dev-air-status
make dev-air-logs  # Ctrl+C to exit
make dev-air-stop

# Test full startup
make dev-air
sleep 5
make dev-air-status
make dev-air-stop
```

**Success:** All targets execute without errors

**Rollback:** Debug individual targets

---

### Step 8: Verify Backward Compatibility (5 min)

```bash
# Ensure existing targets still work
make test
make build
make proto

# Ensure Docker mode unchanged
make dev
sleep 5
make dev-stop
```

**Success:** All existing targets work

**Rollback:** Restore original Makefile

---

## Todo List

- [ ] Read current Makefile
- [ ] Add Air to install-tools target
- [ ] Update help target with both modes
- [ ] Add validate-env target
- [ ] Add dev-air targets
- [ ] Add infra-up/infra-down targets
- [ ] Add dev-air-status target
- [ ] Test all new targets
- [ ] Verify backward compatibility
- [ ] Update README (Phase 07)

---

## Success Criteria

✅ Air tool added to install-tools
✅ Help target shows both Docker and Air modes
✅ All Air targets work (dev-air, dev-air-stop, etc.)
✅ Infra targets work (infra-up, infra-down)
✅ Environment validation works
✅ Existing targets unchanged (make dev, make test, etc.)
✅ Cross-platform compatibility maintained

**Validation Commands:**
```bash
# Check help output
make help | grep -A 5 "MODE 2"

# Test Air mode
make dev-air
make dev-air-status
make dev-air-stop

# Test infra mode
make infra-up
make infra-down

# Verify backward compat
make dev
make dev-stop
```

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Breaking existing targets | Low | High | Test all existing targets |
| Script path issues | Low | Medium | Use relative paths from project root |
| Windows compatibility | Low | Medium | Test on Windows/WSL |
| Confusing help output | Medium | Low | Clear separation of modes |
| make dev-air conflicts with make dev | Low | Low | Different names, separate modes |

---

## Security Considerations

None (Makefile just orchestrates scripts)

---

## Next Steps

**After Phase 05:**
- Phase 06: Testing & Validation (end-to-end testing)
- Phase 07: Documentation (update README)

**Dependencies:**
- Phase 02 required (Air configs)
- Phase 03 required (env vars)
- Phase 04 required (scripts)

**Follow-up Tasks:**
- Update README with new workflow (Phase 07)
- Create quick-start guide

---

## Rollback Plan

**Full Rollback:**
```bash
# Restore original Makefile
git checkout Makefile

# Verify
make help
```

**Partial Rollback:**
```bash
# Remove specific sections manually
# 1. Remove Air tool from install-tools
# 2. Remove Air Mode section
# 3. Remove Infrastructure section
# 4. Restore original help target
```
