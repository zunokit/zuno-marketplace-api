---
title: "Phase 02: Air Configuration"
description: "Setup Air hot-reload configuration for all 4 microservices"
status: pending
priority: P1
effort: 1.5h
branch: feature/hybird-serverless-and-servers-infrastructure
tags: [air, hot-reload, configuration]
created: 2026-01-28
---

# Phase 02: Air Configuration

## Context Links

- [Main Plan](plan.md)
- [Air Research](research/researcher-01-air-hot-reload.md)
- [Phase 01: Cleanup](phase-01-cleanup-artifacts.md)

## Overview

**Priority:** P1 (core dependency)
**Current Status:** Pending
**Description:** Create `.air.toml` configuration files for all 4 services to enable hot-reload during development.

**Services to Configure:**
1. `auth-service` - Port 4001 (gRPC)
2. `user-service` - Port 4002 (gRPC)
3. `wallet-service` - Port 4003 (gRPC)
4. `graphql-gateway` - Port 4080 (HTTP)

---

## Key Insights

From [Air Research Report](research/researcher-01-air-hot-reload.md):
- **Per-service Air instances** work better than monolithic setup
- **Build delay 1000ms** prevents rebuild thrashing
- **Include specific shared directories** for faster rebuilds
- **Go build cache** enabled by default for incremental compilation
- **Windows compatibility:** Air binary works on Windows

**Shared Code Strategy:**
- Option 1: Each service watches specific shared imports (RECOMMENDED)
- Option 2: Watch all shared code (simpler but slower)
- **Decision:** Start with Option 1 (specific imports)

---

## Requirements

### Functional Requirements
1. Create `.air.toml` in each service directory
2. Configure service-specific ports (4xxx range)
3. Enable shared code watching where applicable
4. Set appropriate build delays and exclusions

### Non-Functional Requirements
- Build time < 5 seconds per service
- No file watcher conflicts
- Works on Windows and Unix systems
- Logs output to console (not files)

---

## Architecture

### Directory Structure (After)
```
services/
├── auth-service/
│   ├── .air.toml              ← NEW
│   ├── cmd/
│   └── internal/
├── user-service/
│   ├── .air.toml              ← NEW
│   ├── cmd/
│   └── internal/
├── wallet-service/
│   ├── .air.toml              ← NEW
│   ├── cmd/
│   └── internal/
└── graphql-gateway/
    ├── .air.toml              ← NEW
    ├── cmd/
    └── internal/
```

### Air Process Model
```
Terminal 1:  cd services/auth-service && air
Terminal 2:  cd services/user-service && air
Terminal 3:  cd services/wallet-service && air
Terminal 4:  cd services/graphql-gateway && air
```

Each Air instance:
- Watches its service code
- Builds to `./tmp/main` binary
- Restarts process on file changes
- Logs to stdout/stderr

---

## Related Code Files

### Files to CREATE
- `services/auth-service/.air.toml`
- `services/user-service/.air.toml`
- `services/wallet-service/.air.toml`
- `services/graphql-gateway/.air.toml`

### Files to READ (for context)
- `services/*/cmd/main.go` - Entry point
- `services/*/go.mod` - Dependencies
- `shared/` - Shared code structure

---

## Implementation Steps

### Step 1: Install Air Tool (5 min)
```bash
# Install Air globally
go install github.com/air-verse/air@latest

# Verify installation
air version

# Add to Makefile install-tools target (in Phase 05)
```

**Success:** `air version` shows version number

**Rollback:** No rollback needed (tool installation)

---

### Step 2: Create Auth Service Air Config (15 min)

Create `services/auth-service/.air.toml`:

```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/main.go"
bin = "tmp/main"
include_ext = ["go", "tpl", "tmpl", "html"]
exclude_dir = ["tmp", "vendor", "build", "logs"]
include_dir = ["cmd", "internal", "pkg", "../../shared/auth"]
exclude_file = []
exclude_unchanged = false
follow_symlink = false
delay = 1000
stop_on_root = false
rerun = false
rerun_delay = 500

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

**Key Configuration:**
- `include_dir`: Watch `cmd`, `internal`, `pkg`, and specific shared modules
- `delay`: 1000ms debounce prevents rebuild thrashing
- `tmp_dir`: Build artifacts in `tmp/` (gitignored)

**Success:** File created without syntax errors

**Rollback:** `rm services/auth-service/.air.toml`

---

### Step 3: Create User Service Air Config (10 min)

Create `services/user-service/.air.toml`:

```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/main.go"
bin = "tmp/main"
include_ext = ["go", "tpl", "tmpl", "html"]
exclude_dir = ["tmp", "vendor", "build", "logs"]
include_dir = ["cmd", "internal", "pkg", "../../shared/user", "../../shared/proto"]
exclude_file = []
exclude_unchanged = false
follow_symlink = false
delay = 1000
stop_on_root = false
rerun = false
rerun_delay = 500

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

**Differences from auth:**
- Includes `../../shared/user` and `../../shared/proto`

**Success:** File created

**Rollback:** `rm services/user-service/.air.toml`

---

### Step 4: Create Wallet Service Air Config (10 min)

Create `services/wallet-service/.air.toml`:

```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/main.go"
bin = "tmp/main"
include_ext = ["go", "tpl", "tmpl", "html"]
exclude_dir = ["tmp", "vendor", "build", "logs"]
include_dir = ["cmd", "internal", "pkg", "../../shared/wallet", "../../shared/proto"]
exclude_file = []
exclude_unchanged = false
follow_symlink = false
delay = 1000
stop_on_root = false
rerun = false
rerun_delay = 500

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

**Success:** File created

**Rollback:** `rm services/wallet-service/.air.toml`

---

### Step 5: Create GraphQL Gateway Air Config (10 min)

Create `services/graphql-gateway/.air.toml`:

```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/main.go"
bin = "tmp/main"
include_ext = ["go", "tpl", "tmpl", "html"]
exclude_dir = ["tmp", "vendor", "build", "logs"]
include_dir = ["cmd", "internal", "pkg", "../../shared/proto", "../../shared/graphql"]
exclude_file = []
exclude_unchanged = false
follow_symlink = false
delay = 1000
stop_on_root = false
rerun = false
rerun_delay = 500

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

**Success:** File created

**Rollback:** `rm services/graphql-gateway/.air.toml`

---

### Step 6: Test Single Service Air (20 min)

```bash
# Navigate to auth service
cd services/auth-service

# Start Air
air

# In another terminal, make a code change
echo "// Test change" >> cmd/main.go

# Verify: Air detects change and rebuilds
# Expected output:
#   building...
#   running...
#   main.go changed
#   building...
#   running...

# Stop Air with Ctrl+C
```

**Success:** Air detects changes and rebuilds successfully

**Rollback:** Debug config errors, rebuild service

---

### Step 7: Verify All Service Configs (10 min)

```bash
# Test each service Air config syntax
for service in auth-service user-service wallet-service graphql-gateway; do
  echo "Testing $service..."
  cd "services/$service"
  air --help 2>&1 | head -1  # Should show Air help
  cd ../..
done
```

**Success:** All configs validate without errors

**Rollback:** Fix syntax errors in `.air.toml` files

---

### Step 8: Add tmp/ to .gitignore (5 min)

```bash
# Add tmp/ to .gitignore if not present
grep -q "^tmp/$" .gitignore || echo "tmp/" >> .gitignore
grep -q "services/*/tmp/" .gitignore || echo "services/*/tmp/" >> .gitignore

# Verify
cat .gitignore | grep tmp
```

**Success:** Build artifacts won't be committed

**Rollback:** Remove from .gitignore if needed

---

## Todo List

- [ ] Install Air tool (`go install github.com/air-verse/air@latest`)
- [ ] Create `services/auth-service/.air.toml`
- [ ] Create `services/user-service/.air.toml`
- [ ] Create `services/wallet-service/.air.toml`
- [ ] Create `services/graphql-gateway/.air.toml`
- [ ] Test auth-service Air with code change
- [ ] Verify all service configs syntax
- [ ] Add `tmp/` to .gitignore
- [ ] Document Air usage (Phase 07)

---

## Success Criteria

✅ Air tool installed globally
✅ All 4 services have `.air.toml` configuration
✅ Configs use 4xxx port range (via env vars)
✅ Shared code watching configured per service
✅ Single service hot-reload tested successfully
✅ Build time < 5 seconds
✅ `tmp/` in .gitignore

**Validation Commands:**
```bash
# Verify configs exist
ls -1 services/*/.air.toml

# Test build (no errors)
cd services/auth-service && air --help

# Verify gitignore
grep tmp .gitignore
```

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Air not installed | Low | Medium | Add to Makefile install-tools |
| Port conflicts on 4xxx | Low | Medium | Use env vars, document in .env |
| Shared code rebuilds slow | Medium | Low | Use specific include_dir paths |
| Windows file watcher issues | Medium | Medium | Test on Windows, use WSL2 if needed |
| tmp/ not gitignored | Low | Low | Add to .gitignore explicitly |

---

## Security Considerations

None (development-only configuration)

---

## Next Steps

**After Phase 02:**
- Phase 03: Environment Setup (configure ports in .env)
- Phase 04: Scripts Creation (automate Air startup)

**Dependencies:**
- Phase 01 should be completed first (clean workspace)

**Follow-up Tasks:**
- Create scripts to run all Air services (Phase 04)
- Set environment variables for ports (Phase 03)

---

## Rollback Plan

**Full Rollback:**
```bash
# Remove all .air.toml files
rm services/auth-service/.air.toml
rm services/user-service/.air.toml
rm services/wallet-service/.air.toml
rm services/graphql-gateway/.air.toml

# Uninstall Air (optional)
go install -a github.com/air-verse/air@latest  # Won't uninstall, just overwrite
```

**Partial Rollback:**
```bash
# Remove specific service config
rm services/auth-service/.air.toml
```
