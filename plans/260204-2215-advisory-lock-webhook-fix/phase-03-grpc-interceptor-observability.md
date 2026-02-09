# Phase 3: gRPC Interceptor Observability

**Date:** 2026-02-04
**Priority:** P1 (High)
**Status:** Pending
**Estimated:** 1.5 hours

---

## Context Links

- [Research: gRPC Interceptor](./research/researcher-02-grpc-interceptor.md) - Interceptor patterns
- [Main Plan](./plan.md) - Overview and parallelization

---

## Parallelization Info

**Can run in parallel with:** Phase 1 (Repository Lock Methods)

**Reason:** No file overlap. This phase modifies `main.go` only.

**Blocked by:** None

**Blocks:** Phase 4 (Integration Tests)

---

## Overview

Add Sentry observability interceptor to collection-service gRPC server. Enables distributed tracing for all RPC calls, matching pattern used by wallet-service and auth-service.

**Key Gap:** Collection service currently has NO interceptors (unlike other services).

---

## Key Insights from Research

1. **Collection service lacks observability** - No Sentry tracing currently
2. **Shared middleware exists** - `shared/observability/middleware/grpc.go` is production-ready
3. **Pattern established** - wallet-service, auth-service, user-service all use it
4. **Automatic trace propagation** - Interceptor extracts/continues sentry-trace header

---

## Requirements

### Functional

- Add Sentry initialization to main.go
- Add gRPC interceptor chain with observability
- Continue distributed traces from inbound requests
- Flush Sentry events on shutdown

### Non-Functional

- No performance regression (< 1ms overhead)
- Non-blocking if Sentry init fails
- Graceful degradation if Sentry unavailable

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     main.go                                  │
│  1. Load config                                             │
│  2. Initialize Sentry (non-blocking)                        │
│  3. Create gRPC server with ChainUnaryInterceptor           │
│     ├─ grpcMiddleware.UnaryServerInterceptor()              │
│  4. Register services                                       │
│  5. Serve & Flush Sentry on shutdown                        │
└─────────────────────────────────────────────────────────────┘

Interceptor Flow:
┌─────────────────────────────────────────────────────────────┐
│  Incoming gRPC Request                                      │
│  1. Extract sentry-trace header from metadata               │
│  2. Continue parent span from trace header                  │
│  3. Start child span for this RPC                           │
│  4. Execute handler with span context                       │
│  5. Record error if any                                     │
│  6. Finish span                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Related Code Files

**References:**
- `services/collection-service/cmd/main.go` - Entry point
- `shared/observability/middleware/grpc.go` - Sentry interceptor
- `services/collection-service/internal/config/config.go` - Config structure
- `services/wallet-service/cmd/main.go` - Reference implementation

**Creates:**
- None (modifies existing files only)

---

## File Ownership (EXCLUSIVE)

This phase EXCLUSIVELY modifies:

1. **`services/collection-service/cmd/main.go`**
   - Add Sentry initialization
   - Add gRPC interceptor chain
   - Add Sentry flush on shutdown
   - Add `getBuildVersion()` function

**Conflict Prevention:** No other phases modify main.go.

---

## Implementation Steps

### Step 1: Add Imports

**File:** `services/collection-service/cmd/main.go`

**Add to imports:**
```go
import (
    "time"

    grpcMiddleware "github.com/zunokit/zuno-marketplace-api/shared/observability/middleware"
    obs "github.com/zunokit/zuno-marketplace-api/shared/observability/sentry"
    obsTrace "github.com/zunokit/zuno-marketplace-api/shared/observability/tracing"
    // ... existing imports
)
```

### Step 2: Add Build Version Variables

**After:** `package main`

```go
// Version and BuildTime are injected via ldflags during build
var (
    Version   = "dev"
    BuildTime = "unknown"
)
```

### Step 3: Add Sentry Initialization

**Location:** After config load, before database init (after line 35)

```go
// Initialize Sentry (non-blocking)
if cfg.Sentry.DSN != "" {
    if err := obs.Init(
        cfg.Sentry.DSN,
        cfg.Sentry.Environment,
        "collection-service",
        getBuildVersion(),
        obsTrace.GetTracesSampleRate(cfg.Sentry.Environment),
    ); err != nil {
        log.Infof("Sentry init failed (continuing): %v", err)
    } else {
        log.Info("Sentry initialized")
        defer obs.Flush(2 * time.Second)
    }
} else {
    log.Info("Sentry DSN not configured, skipping")
}
```

**Key Points:**
- Non-blocking - continues if init fails
- Defer flush - ensures events sent before shutdown
- Service name: "collection-service"

### Step 4: Add gRPC Interceptor Chain

**Location:** Replace existing gRPC server creation (lines 67-70)

**Replace:**
```go
grpcServer := grpc.NewServer(
    grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
    grpc.MaxSendMsgSize(10*1024*1024), // 10MB
)
```

**With:**
```go
// Create gRPC server with Sentry interceptor
grpcServer := grpc.NewServer(
    grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
    grpc.MaxSendMsgSize(10*1024*1024), // 10MB
    grpc.ChainUnaryInterceptor(
        grpcMiddleware.UnaryServerInterceptor(),
    ),
)
```

### Step 5: Add Sentry Flush in Shutdown

**Location:** In graceful shutdown goroutine (after line 96)

```go
go func() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan

    log.Info("Shutting down Collection Service...")

    // Flush Sentry before shutdown
    if cfg.Sentry.DSN != "" {
        log.Info("Flushing Sentry events...")
        obs.Flush(2 * time.Second)
    }

    grpcServer.GracefulStop()
    log.Info("Collection Service stopped")
}()
```

### Step 6: Add Build Version Helper

**Location:** End of file (before closing `}`)

```go
// getBuildVersion returns the version injected by build ldflags
func getBuildVersion() string {
    return Version
}
```

### Step 7: Update Config Structure

**File:** `services/collection-service/internal/config/config.go`

**Add Sentry config (confirmed needed from validation):**
```go
type Config struct {
    // ... existing fields ...
    Sentry SentryConfig `mapstructure:"sentry"`
}

type SentryConfig struct {
    DSN         string `mapstructure:"dsn"`
    Environment string `mapstructure:"environment"`
}
```

**Reference:** Pattern from `services/wallet-service/internal/config/config.go`

**Environment variables:**
```env
SENTRY_DSN=https://<key>@sentry.io/<project>
SENTRY_ENVIRONMENT=development
```

---

## Todo List

- [ ] Add build version variables
- [ ] Add observability imports
- [ ] Add Sentry initialization (non-blocking)
- [ ] Add gRPC interceptor chain
- [ ] Add Sentry flush in shutdown handler
- [ ] Add `getBuildVersion()` helper
- [ ] **Add Sentry config to config.go** (confirmed in validation)
- [ ] Run `go build ./services/collection-service/cmd/`
- [ ] Test compilation succeeds
- [ ] Verify service starts without errors

---

## Success Criteria

- [ ] Sentry interceptor added to gRPC server
- [ ] Service compiles without errors
- [ ] Service starts without Sentry (graceful degradation)
- [ ] Distributed tracing spans created for RPCs
- [ ] Sentry flush on shutdown

---

## Conflict Prevention

**Files modified exclusively by this phase:**
- `main.go` - Only Phase 3 touches this

**No overlap with:**
- Phase 1 (repository layer)
- Phase 2 (webhook handler)
- Phase 4 (integration tests)

**Can run parallel with:** Phase 1 (different files)

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Sentry init blocks service startup | Non-blocking - continues on failure |
| Performance regression | Interceptor overhead < 1ms |
| Missing Sentry config | Add config fields or use empty defaults |
| Build flags not set | Default to "dev" version |

---

## Security Considerations

- Sentry DSN from config (not hardcoded)
- No sensitive data in traces by default
- Trace ID propagation via metadata (standard practice)

---

## Next Steps

**After this phase completes:**
1. Phase 4 (Integration Tests) can verify distributed tracing
2. Deploy to staging to verify Sentry integration
3. Verify traces appear in Sentry dashboard

**Can run parallel with:** Phase 1 (no dependency)

---

## Verification Steps

**Manual testing:**
1. Build service: `go build -ldflags "-X main.Version=1.0.0" ./services/collection-service/cmd/`
2. Run service locally
3. Send webhook request
4. Check Sentry dashboard for trace

**Expected trace structure:**
```
Trace: <trace-id>
├─ grpc.server: /pb.CollectionService/ProcessIndexerWebhook
│  ├─ db:postgres.query (acquire lock)
│  ├─ db:postgres.query (check processed)
│  └─ db:postgres.query (update collection)
```
