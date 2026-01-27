---
title: "Phase 04: Scripts Creation"
description: "Create development scripts for Air hot-reload workflow"
status: pending
priority: P1
effort: 1.5h
branch: feature/hybird-serverless-and-servers-infrastructure
tags: [scripts, bash, air, hot-reload, automation]
created: 2026-01-28
---

# Phase 04: Scripts Creation

## Context Links

- [Main Plan](plan.md)
- [Hybrid Workflow Research](research/researcher-02-hybrid-workflow.md)
- [Phase 02: Air Configuration](phase-02-air-configuration.md)
- [Phase 03: Environment Setup](phase-03-environment-setup.md)

## Overview

**Priority:** P1 (enables single-command dev workflow)
**Current Status:** Pending
**Description:** Create shell scripts to automate Air service startup, shared code watching, and environment setup.

**Scripts to Create:**
1. `scripts/dev-air.sh` - Start all/specific services with Air
2. `scripts/watch-shared.sh` - Monitor shared/ code changes
3. `scripts/stop-air.sh` - Stop all Air services
4. `scripts/validate-env.sh` - Verify environment configuration

---

## Key Insights

From [Hybrid Workflow Research](research/researcher-02-hybrid-workflow.md):
- **Process management:** Use PID files for service tracking
- **Logging:** Redirect to `logs/` directory for debugging
- **Graceful shutdown:** Capture signals for clean shutdown
- **Windows compatibility:** Bash scripts (require WSL/Git Bash)

**Shared Code Challenge:**
- Multiple Air instances watching `shared/` can conflict
- Solution: Separate watcher script that restarts services on shared changes
- Alternative: Each service watches specific shared imports (Phase 02 config)

---

## Requirements

### Functional Requirements
1. `dev-air.sh` starts any combination of services
2. `watch-shared.sh` detects changes in `shared/` and `proto/`
3. `stop-air.sh` cleanly stops all Air processes
4. `validate-env.sh` checks required environment variables

### Non-Functional Requirements
- Cross-platform (Linux/macOS/Windows-WSL)
- Idempotent (can run multiple times safely)
- Clear output formatting
- Graceful error handling

---

## Architecture

### Script Workflow

```
make dev-air
  ↓
scripts/dev-air.sh all
  ↓
┌─────────────────────────────────────┐
│  Start 4 Air instances (parallel)   │
├─────────────────────────────────────┤
│  auth-service    → logs/auth.log    │
│  user-service    → logs/user.log    │
│  wallet-service  → logs/wallet.log  │
│  graphql-gateway → logs/gateway.log │
└─────────────────────────────────────┘
  ↓
[Optional] scripts/watch-shared.sh
  ↓
Detect shared/ changes → Restart all services
```

### Directory Structure (After)
```
scripts/
├── dev-air.sh          ← NEW (start services)
├── watch-shared.sh     ← NEW (watch shared code)
├── stop-air.sh         ← NEW (stop services)
├── validate-env.sh     ← NEW (validate env)
├── setup-env.sh        ← UPDATED (from Phase 03)
└── health-check.sh     ← EXISTING (keep)
logs/                   ← NEW (service logs)
├── auth-service.log
├── user-service.log
├── wallet-service.log
└── graphql-gateway.log
```

---

## Related Code Files

### Files to CREATE
- `scripts/dev-air.sh`
- `scripts/watch-shared.sh`
- `scripts/stop-air.sh`
- `scripts/validate-env.sh`

### Files to MODIFY
- `scripts/setup-env.sh` (from Phase 03)

### Files to READ (for context)
- `services/*/.air.toml` (from Phase 02)
- `.env.development.example` (from Phase 03)

---

## Implementation Steps

### Step 1: Create logs/ Directory (2 min)

```bash
# Create logs directory
mkdir -p logs

# Add to .gitignore
grep -q "^logs/$" .gitignore || echo "logs/" >> .gitignore
grep -q "^logs/*.log" .gitignore || echo "logs/*.log" >> .gitignore

# Verify
ls -la logs/
```

**Success:** `logs/` directory created and gitignored

**Rollback:** `rmdir logs && git checkout .gitignore`

---

### Step 2: Create dev-air.sh Script (25 min)

Create `scripts/dev-air.sh`:

```bash
#!/bin/bash
# Start services with Air hot-reload
# Usage: ./scripts/dev-air.sh [all|auth|user|wallet|gateway]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVICES_DIR="$PROJECT_ROOT/services"
LOGS_DIR="$PROJECT_ROOT/logs"
PID_DIR="$PROJECT_ROOT/logs"

# Ensure logs directory exists
mkdir -p "$LOGS_DIR"

# Service configurations (name:port)
declare -A SERVICES=(
  ["auth-service"]="4001"
  ["user-service"]="4002"
  ["wallet-service"]="4003"
  ["graphql-gateway"]="4080"
)

# Function to print colored output
print_info() {
  echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
  echo -e "${GREEN}✓${NC} $1"
}

print_error() {
  echo -e "${RED}✗${NC} $1"
}

print_header() {
  echo
  echo -e "${YELLOW}============================================================${NC}"
  echo -e "${YELLOW}  $1${NC}"
  echo -e "${YELLOW}============================================================${NC}"
  echo
}

# Function to start a single service
start_service() {
  local service=$1
  local port=$2
  local log_file="$LOGS_DIR/$service.log"
  local pid_file="$PID_DIR/$service.pid"

  # Check if already running
  if [ -f "$pid_file" ]; then
    local pid=$(cat "$pid_file")
    if ps -p "$pid" > /dev/null 2>&1; then
      print_info "$service already running (PID: $pid)"
      return 0
    else
      rm -f "$pid_file"
    fi
  fi

  print_info "Starting $service on port $port..."

  # Navigate to service directory
  cd "$SERVICES_DIR/$service"

  # Start Air in background
  nohup air > "$log_file" 2>&1 &
  local pid=$!

  # Save PID
  echo $pid > "$pid_file"

  cd "$PROJECT_ROOT"

  # Wait briefly for startup
  sleep 2

  # Verify process started
  if ps -p $pid > /dev/null 2>&1; then
    print_success "$service started (PID: $pid)"
    return 0
  else
    print_error "$service failed to start. Check $log_file"
    return 1
  fi
}

# Parse arguments
SERVICE_ARG="${1:-all}"

print_header "Starting Air Development Environment"

# Validate Air is installed
if ! command -v air &> /dev/null; then
  print_error "Air is not installed"
  echo
  echo "Install with: go install github.com/air-verse/air@latest"
  exit 1
fi

# Validate .env file exists
if [ ! -f "$PROJECT_ROOT/.env.development" ]; then
  print_error ".env.development not found"
  echo
  echo "Create from example: cp .env.development.example .env.development"
  echo "Then edit with your credentials"
  exit 1
fi

# Start services based on argument
case "$SERVICE_ARG" in
  all)
    print_info "Starting all services..."
    for service in "${!SERVICES[@]}"; do
      start_service "$service" "${SERVICES[$service]}"
    done
    ;;
  auth|auth-service)
    start_service "auth-service" "${SERVICES[auth-service]}"
    ;;
  user|user-service)
    start_service "user-service" "${SERVICES[user-service]}"
    ;;
  wallet|wallet-service)
    start_service "wallet-service" "${SERVICES[wallet-service]}"
    ;;
  gateway|graphql-gateway)
    start_service "graphql-gateway" "${SERVICES[graphql-gateway]}"
    ;;
  *)
    print_error "Unknown service: $SERVICE_ARG"
    echo
    echo "Usage: $0 [all|auth|user|wallet|gateway]"
    exit 1
    ;;
esac

# Print summary
print_header "Services Running"

for service in "${!SERVICES[@]}"; do
  local pid_file="$PID_DIR/$service.pid"
  if [ -f "$pid_file" ]; then
    local pid=$(cat "$pid_file")
    if ps -p "$pid" > /dev/null 2>&1; then
      echo -e "  ${GREEN}●${NC} $service - Port ${SERVICES[$service]} (PID: $pid)"
    else
      echo -e "  ${RED}○${NC} $service - Stopped"
    fi
  else
    echo -e "  ${YELLOW}○${NC} $service - Not started"
  fi
done

echo
echo -e "GraphQL Playground: ${GREEN}http://localhost:4080/graphql${NC}"
echo
echo -e "Logs: ${BLUE}$LOGS_DIR${NC}"
echo -e "Stop: ${YELLOW}make dev-air-stop${NC} or ${YELLOW}./scripts/stop-air.sh${NC}"
echo
```

Make executable:
```bash
chmod +x scripts/dev-air.sh
```

**Success:** Script created and executable

**Rollback:** `rm scripts/dev-air.sh`

---

### Step 3: Create stop-air.sh Script (10 min)

Create `scripts/stop-air.sh`:

```bash
#!/bin/bash
# Stop all Air services

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_DIR="$PROJECT_ROOT/logs"

echo -e "${YELLOW}============================================================${NC}"
echo -e "${YELLOW}  Stopping Air Services${NC}"
echo -e "${YELLOW}============================================================${NC}"
echo

# Stop services from PID files
if [ -d "$PID_DIR" ]; then
  for pid_file in "$PID_DIR"/*.pid; do
    if [ -f "$pid_file" ]; then
      service=$(basename "$pid_file" .pid)
      pid=$(cat "$pid_file")

      if ps -p "$pid" > /dev/null 2>&1; then
        echo -e "  Stopping ${GREEN}$service${NC} (PID: $pid)..."
        kill "$pid"
        rm -f "$pid_file"
      else
        echo -e "  ${YELLOW}$service${NC} not running (stale PID file)"
        rm -f "$pid_file"
      fi
    fi
  done
fi

# Also kill any remaining Air processes
pkill -f "air" || true

echo
echo -e "${GREEN}✓ All services stopped${NC}"
```

Make executable:
```bash
chmod +x scripts/stop-air.sh
```

**Success:** Script created

**Rollback:** `rm scripts/stop-air.sh`

---

### Step 4: Create watch-shared.sh Script (20 min)

Create `scripts/watch-shared.sh`:

```bash
#!/bin/bash
# Watch shared/ and proto/ for changes, restart dependent services
# Requires: entr (install with: brew install entr / apt install entr)

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

print_info() {
  echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
  echo -e "${GREEN}✓${NC} $1"
}

echo -e "${YELLOW}============================================================${NC}"
echo -e "${YELLOW}  Watching Shared Code${NC}"
echo -e "${YELLOW}============================================================${NC}"
echo

# Check if entr is installed
if ! command -v entr &> /dev/null; then
  print_info "entr not found. Installing..."
  if [[ "$OSTYPE" == "darwin"* ]]; then
    brew install entr
  else
    sudo apt-get install -y entr
  fi
fi

print_info "Monitoring shared/ and proto/ for changes..."
echo
print_info "Press Ctrl+C to stop"
echo

# Watch for changes
while true; do
  # Find Go files in shared/ and proto/
  find shared proto -name "*.go" -type f 2>/dev/null | \
    entr -dp -n echo "changed" 2>/dev/null

  # If changes detected, restart services
  if [ $? -eq 0 ]; then
    echo
    print_success "Shared code changed at $(date)"
    print_info "Restarting all services..."
    echo

    # Stop all services
    "$PROJECT_ROOT/scripts/stop-air.sh"

    # Wait for cleanup
    sleep 2

    # Start all services
    "$PROJECT_ROOT/scripts/dev-air.sh" all

    echo
    print_info "Waiting for next change..."
    echo
  fi

  sleep 1
done
```

Make executable:
```bash
chmod +x scripts/watch-shared.sh
```

**Success:** Script created

**Rollback:** `rm scripts/watch-shared.sh`

---

### Step 5: Create validate-env.sh Script (15 min)

Create `scripts/validate-env.sh`:

```bash
#!/bin/bash
# Validate .env.development configuration

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$PROJECT_ROOT/.env.development"

echo -e "${YELLOW}============================================================${NC}"
echo -e "${YELLOW}  Environment Validation${NC}"
echo -e "${YELLOW}============================================================${NC}"
echo

# Check if .env exists
if [ ! -f "$ENV_FILE" ]; then
  echo -e "${RED}✗${NC} .env.development not found"
  echo
  echo "Create with: cp .env.development.example .env.development"
  exit 1
fi

# Load environment
set -a
source "$ENV_FILE"
set +a

# Required variables
errors=0

check_var() {
  local var_name=$1
  local var_value=${!var_name}
  local is_required=${2:-true}

  if [ -z "$var_value" ]; then
    if [ "$is_required" = "true" ]; then
      echo -e "${RED}✗${NC} $var_name is not set"
      ((errors++))
    else
      echo -e "${YELLOW}○${NC} $var_name is optional (not set)"
    fi
  elif [[ "$var_value" == *"["* ]]; then
    echo -e "${YELLOW}○${NC} $var_name contains placeholder (needs real value)"
    ((errors++))
  else
    echo -e "${GREEN}✓${NC} $var_name"
  fi
}

echo "Required Variables:"
check_var "INFRA_MODE"
check_var "ENVIRONMENT"
check_var "AUTH_GRPC_PORT"
check_var "USER_GRPC_PORT"
check_var "WALLET_GRPC_PORT"
check_var "GATEWAY_HTTP_ADDR"
check_var "JWT_SECRET"
check_var "REFRESH_SECRET"

echo
echo "Infrastructure (depends on INFRA_MODE):"

if [ "$INFRA_MODE" = "serverless" ]; then
  check_var "DATABASE_URL" true
  check_var "REDIS_URL" true
  check_var "CLOUDAMQP_URL" true
else
  echo -e "${YELLOW}○${NC} Docker mode - checking fallback config"
  check_var "POSTGRES_HOST" true
  check_var "POSTGRES_PORT" true
  check_var "REDIS_HOST" true
  check_var "REDIS_PORT" true
  check_var "RABBITMQ_HOST" true
  check_var "RABBITMQ_PORT" true
fi

echo
echo "Optional Variables:"
check_var "ETHEREUM_RPC_URL" false
check_var "POLYGON_RPC_URL" false
check_var "IPFS_API_URL" false
check_var "IPFS_API_KEY" false

echo
if [ $errors -gt 0 ]; then
  echo -e "${RED}✗${NC} Validation failed with $errors error(s)"
  echo
  echo "Please fix the issues above in $ENV_FILE"
  exit 1
else
  echo -e "${GREEN}✓${NC} All required variables are set"
  echo
  exit 0
fi
```

Make executable:
```bash
chmod +x scripts/validate-env.sh
```

**Success:** Script created

**Rollback:** `rm scripts/validate-env.sh`

---

### Step 6: Test All Scripts (15 min)

```bash
# Test environment validation
./scripts/validate-env.sh

# Test starting single service
./scripts/dev-air.sh auth
sleep 3
./scripts/stop-air.sh

# Test starting all services
./scripts/dev-air.sh all
sleep 5

# Check logs
ls -la logs/
tail -n 20 logs/auth-service.log

# Stop all services
./scripts/stop-air.sh
```

**Success:** All scripts work without errors

**Rollback:** Debug script issues individually

---

### Step 7: Add scripts/*.sh to .gitignore (5 min)

```bash
# Don't ignore scripts, but ensure they're executable
git update-index --chmod=+x scripts/*.sh

# Verify
git ls-files --stage scripts/*.sh
```

**Success:** Scripts are executable in git

**Rollback:** None (just metadata)

---

## Todo List

- [ ] Create `logs/` directory
- [ ] Create `scripts/dev-air.sh`
- [ ] Create `scripts/stop-air.sh`
- [ ] Create `scripts/watch-shared.sh`
- [ ] Create `scripts/validate-env.sh`
- [ ] Make all scripts executable
- [ ] Test dev-air.sh with single service
- [ ] Test dev-air.sh with all services
- [ ] Test stop-air.sh
- [ ] Test validate-env.sh
- [ ] Verify logs are created

---

## Success Criteria

✅ All scripts created and executable
✅ `dev-air.sh` starts any service combination
✅ `stop-air.sh` cleanly stops all services
✅ `watch-shared.sh` monitors shared/ code
✅ `validate-env.sh` checks required variables
✅ Logs written to `logs/` directory
✅ Scripts work on Linux/macOS/WSL

**Validation Commands:**
```bash
# Check scripts exist
ls -l scripts/*.sh

# Test env validation
./scripts/validate-env.sh

# Test start/stop
./scripts/dev-air.sh auth
./scripts/stop-air.sh
```

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Scripts not executable | Low | Medium | chmod +x in Step 2-5 |
| entr not installed | Medium | Low | Auto-install in watch-shared.sh |
| Port conflicts | Low | Medium | Validate env before starting |
| Zombie processes | Low | Low | stop-air.sh uses PID files |
| Windows incompatibility | Medium | Medium | Document WSL requirement |

---

## Security Considerations

- Scripts execute Air with current user permissions
- No privilege escalation required
- `.env` files never logged (only validation messages)
- PID files may expose PIDs (not sensitive)

---

## Next Steps

**After Phase 04:**
- Phase 05: Makefile Enhancement (integrate scripts)
- Phase 06: Testing & Validation (end-to-end test)

**Dependencies:**
- Phase 02 required (Air configs)
- Phase 03 required (env vars)

**Follow-up Tasks:**
- Test scripts on Windows/WSL
- Document script usage in Phase 07

---

## Rollback Plan

**Full Rollback:**
```bash
# Remove all scripts
rm scripts/dev-air.sh
rm scripts/stop-air.sh
rm scripts/watch-shared.sh
rm scripts/validate-env.sh
rm -rf logs/

# Restore .gitignore
git checkout .gitignore
```

**Partial Rollback:**
```bash
# Remove specific script
rm scripts/watch-shared.sh
```
