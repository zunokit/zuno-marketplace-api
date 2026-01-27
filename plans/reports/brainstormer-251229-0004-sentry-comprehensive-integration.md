# Brainstorming Report: Comprehensive Sentry Integration

**Date**: 2025-12-29
**Type**: Solution Brainstorming
**Issue**: 251229-0004
**Status**: ✅ APPROVED - Ready for Implementation

---

## Problem Statement

Integrate **comprehensive Sentry monitoring** (error tracking + performance monitoring + distributed tracing) into **zuno-marketplace-api** - a Go microservices NFT Marketplace backend.

### Requirements Collected

- **Scope**: Full Sentry (errors + performance + distributed tracing) across all microservices
- **Architecture**: 4 services (GraphQL Gateway + Auth/User/Wallet via gRPC)
- **Environment**: Development/Staging (cost-conscious, <1000 req/s)
- **Security**: Auto-scrub wallet addresses, JWTs, and PII
- **CI/CD**: Full GitHub Actions integration with deploy tracking

### Current State Assessment

- **Clean skeleton** (v0.1.0) - ideal time to add observability
- **No existing monitoring** - starting fresh
- **gRPC internal comms** - needs trace propagation
- **GraphQL BFF layer** - needs resolver instrumentation
- **Shared config pattern** - using `shared/env` package

---

## Approaches Evaluated

| Approach                         | Pros                                            | Cons                                                   | Verdict         |
| -------------------------------- | ----------------------------------------------- | ------------------------------------------------------ | --------------- |
| **1. Native Sentry per service** | Simple start, zero abstraction                  | Code duplication, DRY violation, inconsistent tracing  | ❌ Rejected     |
| **2. OpenTelemetry + Sentry**    | Vendor-agnostic, standard instrumentation       | Over-engineered, 3x dependencies, steep learning curve | ❌ Rejected     |
| **3. Hybrid KISS (Recommended)** | Single SDK, shared package, cost-optimized, DRY | Locked to Sentry (acceptable)                          | ✅ **SELECTED** |

### Approach 3 Rationale

**YAGNI Principle**: OpenTelemetry adds complexity for "future flexibility" when committing to Sentry for dev/staging.

**KISS Principle**: Single `sentry-go` dependency with minimal abstraction layers.

**DRY Principle**: Shared `shared/observability` package reused across all services.

**Cost Optimization**: Smart sampling (100% errors, 5-20% traces based on endpoint importance).

---

## Final Recommended Solution: Approach 3 - Hybrid KISS

### Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│                      Sentry Cloud                                │
│  ┌────────────┐  ┌──────────────┐  ┌─────────────────────────┐  │
│  │   Errors   │  │ Performance  │  │ Distributed Tracing     │  │
│  └────────────┘  └──────────────┘  └─────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼ sentry-trace header
┌──────────────────────────────────────────────────────────────────┐
│              shared/observability/ (NEW PACKAGE)                 │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  sentry.go                                                 │  │
│  │  - Init(dsn, env, service, sampleRate)                     │  │
│  │  - Flush()                                                 │  │
│  │  - scrubSensitiveData() - Wallet/JWT/Email filtering       │  │
│  └────────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  middleware/                                               │  │
│  │  ├── http.go      - Chi middleware for GraphQL Gateway     │  │
│  │  ├── grpc.go      - Server + Client interceptors          │  │
│  │  └── graphql.go   - gqlgen field middleware                │  │
│  └────────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  tracing/                                                  │  │
│  │  ├── propagator.go - sentry-trace header injection        │  │
│  │  └── sampler.go     - Environment/endpoint-based sampling  │  │
│  └────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
         │                    │                    │
         ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│ GraphQL Gateway │  │  Auth Service   │  │ User/Wallet     │
│  (HTTP + GraphQL│  │   (gRPC)        │  │   (gRPC)        │
│   + Chi Router) │  │                 │  │                 │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

### File Structure

```
shared/observability/
├── sentry/
│   ├── sentry.go           # Init, Flush, BeforeSend hook
│   └── scrubber.go         # Regex patterns for sensitive data
├── middleware/
│   ├── http.go             # Chi middleware
│   ├── grpc.go             # gRPC interceptors (unary + stream)
│   └── graphql.go          # gqlgen field middleware
├── tracing/
│   ├── propagator.go       # Trace ID injection/extraction
│   └── sampler.go          # Smart sampling logic
└── README.md               # Usage documentation

services/*/cmd/main.go              # Add sentry.Init() call
services/*/internal/config/config.go # Add Sentry config fields
```

---

## Implementation Specification

### Phase 1: Core Sentry Package

**File**: `shared/observability/sentry/sentry.go`

```go
package sentry

import (
    "time"
    "github.com/getsentry/sentry-go"
)

// Init initializes Sentry with production-ready defaults
func Init(dsn, environment, service, release string, tracesSampleRate float64) error {
    return sentry.Init(sentry.ClientOptions{
        Dsn:              dsn,
        Environment:      environment,
        Release:          release,
        TracesSampleRate: tracesSampleRate,
        EnableTracing:    true,

        // Security: Auto-scrub sensitive data
        SendDefaultPII:      false,
        MaxRequestBodyBytes: 10_000,
        MaxBreadcrumbs:      100,

        // Privacy scrubbing hook
        BeforeSend: scrubEvent,
        BeforeSendTransaction: scrubTransaction,

        // Attach stacktraces
        AttachStacktrace: true,
        StacktraceConfig: sentry.StacktraceConfig{
            ContextLines: 5,
        },
    })
}

// Flush ensures all events are sent before shutdown
func Flush(timeout time.Duration) bool {
    return sentry.Flush(timeout)
}
```

**File**: `shared/observability/sentry/scrubber.go`

```go
package sentry

import (
    "regexp"
    "github.com/getsentry/sentry-go"
)

var (
    ethereumAddrPattern = regexp.MustCompile(`0x[a-fA-F0-9]{40}`)
    jwtPattern          = regexp.MustCompile(`eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]*`)
    emailPattern        = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
)

func scrubEvent(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
    if event.Request != nil {
        scrubHeaders(event.Request.Headers)
        if event.Request.Data != nil {
            event.Request.Data = scrubMap(event.Request.Data)
        }
    }
    if event.Breadcrumbs != nil {
        for _, b := range event.Breadcrumbs {
            if b.Data != nil {
                b.Data = scrubMap(b.Data)
            }
        }
    }
    if event.Extra != nil {
        event.Extra = scrubMap(event.Extra)
    }
    return event
}

func scrubTransaction(tx *sentry.Event, hint *sentry.EventHint) *sentry.Event {
    // Same scrubbing for transactions
    return scrubEvent(tx, hint)
}

func scrubMap(data map[string]interface{}) map[string]interface{} {
    // Apply regex patterns to all string values
    return data // Implementation details...
}
```

### Phase 2: HTTP Middleware

**File**: `shared/observability/middleware/http.go`

```go
package middleware

import (
    "net/http"
    "fmt"
    "github.com/go-chi/chi/v5"
    "github.com/getsentry/sentry-go"
)

func SentryHTTP(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip health check tracing (optional)
        if r.URL.Path == "/health" {
            next.ServeHTTP(w, r)
            return
        }

        // Start transaction
        transaction := sentry.StartTransaction(r.Context(),
            fmt.Sprintf("%s %s", r.Method, r.URL.Path),
            sentry.TransactionName(r.URL.Path),
            sentry.Op("http.server"),
        )
        defer transaction.Finish()

        // Add HTTP context
        transaction.SetData("http.method", r.Method)
        transaction.SetData("http.url", r.URL.String())
        transaction.SetData("http.scheme", r.URL.Scheme)
        transaction.SetData("http.host", r.Host)

        // Continue with trace context
        r = r.WithContext(transaction.Context())
        next.ServeHTTP(w, r)
    })
}
```

### Phase 3: gRPC Interceptors

**File**: `shared/observability/middleware/grpc.go`

```go
package middleware

import (
    "context"
    "fmt"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
    "github.com/getsentry/sentry-go"
)

// UnaryServerInterceptor traces incoming gRPC requests
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        // Extract sentry-trace from metadata
        var opts []sentry.SpanOption
        if md, ok := metadata.FromIncomingContext(ctx); ok {
            if traceHeader := md["sentry-trace"]; len(traceHeader) > 0 {
                // Continue parent span
                opts = append(opts, sentry.WithTransactionSource())
            }
        }

        // Start span
        span := sentry.StartSpan(ctx, info.FullMethod, opts...)
        span.SetData("grpc.method", info.FullMethod)
        span.Op = "grpc.server")

        ctx = span.Context()
        defer span.Finish()

        // Execute handler
        resp, err := handler(ctx, req)

        // Record error if any
        if err != nil {
            span.SetData("grpc.error", err.Error())
        }

        return resp, err
    }
}

// UnaryClientInterceptor injects trace into outbound gRPC calls
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
    return func(
        ctx context.Context,
        method string,
        req, reply interface{},
        cc *grpc.ClientConn,
        invoker grpc.UnaryInvoker,
        opts ...grpc.CallOption,
    ) error {
        span := sentry.StartSpan(ctx, method, sentry.Op("grpc.client"))
        defer span.Finish()

        // Inject sentry-trace header
        md, ok := metadata.FromOutgoingContext(ctx)
        if !ok {
            md = metadata.New(nil)
        }
        if traceID := span.TraceID.String(); traceID != "" {
            md = md.Copy()
            // Add sentry-trace header
        }
        ctx = metadata.NewOutgoingContext(ctx, md)

        return invoker(span.Context(), method, req, reply, cc, opts...)
    }
}
```

### Phase 4: Smart Sampling

**File**: `shared/observability/tracing/sampler.go`

```go
package tracing

import (
    "github.com/getsentry/sentry-go"
)

// GetTracesSampleRate returns sampling rate based on environment
func GetTracesSampleRate(environment string) float64 {
    switch environment {
    case "production":
        return 0.05  // 5% in prod (cost control)
    case "staging":
        return 0.20  // 20% in staging
    default:
        return 1.0   // 100% in development
    }
}

// TracesSampler provides endpoint-specific sampling
func TracesSampler() sentry.TracesSampler {
    return func(ctx sentry.SamplingContext) float64 {
        // Always trace critical operations
        if ctx.Span.Description == "POST /graphql" {
            // Could check for mutations
            return 1.0
        }

        // Never trace health checks
        if ctx.Span.Name == "GET /health" {
            return 0.0
        }

        // Use default rate
        return GetTracesSampleRate(ctx.ClientOptions.Environment)
    }
}
```

### Phase 5: Service Integration Example

**File**: `services/auth-service/cmd/main.go`

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/getsentry/sentry-go"
    obs "github.com/zunokit/zuno-marketplace-api/shared/observability/sentry"
    obsgrpc "github.com/zunokit/zuno-marketplace-api/shared/observability/middleware"
    // ... other imports
)

func main() {
    // Load config
    cfg := config.Load()

    // Initialize Sentry
    if err := obs.Init(
        cfg.Sentry.DSN,
        cfg.Sentry.Environment,
        "auth-service",
        buildVersion(),
        tracing.GetTracesSampleRate(cfg.Sentry.Environment),
    ); err != nil {
        log.Printf("Sentry init failed: %v", err)
    }
    defer obs.Flush(2 * time.Second)

    // ... database, repository setup ...

    // Create gRPC server with Sentry interceptor
    grpcServer := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            obsgrpc.UnaryServerInterceptor(),
        ),
    )

    // ... rest of setup ...

    // Graceful shutdown
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
        <-sigChan
        log.Println("Shutting down...")
        obs.Flush(2 * time.Second)
        grpcServer.GracefulStop()
    }()

    // Start server
    if err := grpcServer.Serve(listener); err != nil {
        sentry.CaptureException(err)
        log.Fatal(err)
    }
}
```

**File**: `services/graphql-gateway/cmd/main.go`

```go
package main

import (
    obs "github.com/zunokit/zuno-marketplace-api/shared/observability/sentry"
    obshttp "github.com/zunokit/zuno-marketplace-api/shared/observability/middleware"
    obsgrpc "github.com/zunokit/zuno-marketplace-api/shared/observability/middleware"
)

func main() {
    cfg := config.Load()

    // Init Sentry
    obs.Init(cfg.Sentry.DSN, cfg.Sentry.Environment, "graphql-gateway", ...)

    // ...

    router := chi.NewRouter()
    router.Use(obshttp.SentryHTTP)  // Add Sentry middleware
    router.Use(middleware.Logger)
    router.Use(middleware.Recoverer)

    // Connect to gRPC services with tracing
    authConn, err := grpc.Dial(
        cfg.Services.AuthServiceURL,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithChainUnaryInterceptor(obsgrpc.UnaryClientInterceptor()),
    )

    // ...
}
```

### Phase 6: Configuration

**File**: `services/*/internal/config/config.go`

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Sentry   SentryConfig  // NEW
    // ... other fields
}

type SentryConfig struct {
    DSN         string
    Environment string
}
```

**File**: `.env.example`

```env
# Sentry Monitoring
SENTRY_DSN=https://xxxxxxxxxxxx@xxxxx.ingest.sentry.io/xxxxxxx
SENTRY_ENVIRONMENT=development
```

---

## CI/CD Integration

### GitHub Actions Workflow

**File**: `.github/workflows/deploy.yml`

```yaml
name: Deploy to Staging

on:
  push:
    branches: [main, develop]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0 # Full history for Sentry

      - name: Create Sentry Release
        run: |
          npx @sentry/wizard@latest -i github -p bash
        env:
          SENTRY_AUTH_TOKEN: ${{ secrets.SENTRY_AUTH_TOKEN }}
          SENTRY_ORG: zuno
          SENTRY_PROJECT: zuno-marketplace-api

      - name: Build & Deploy
        run: |
          docker compose -f docker-compose.staging.yml up -d

      - name: Notify Sentry of Deploy
        run: |
          VERSION=$(git describe --tags --always)
          curl -sL "{{ SENTRY_DEPLOY_WEBHOOK }}" \
            -d version="$VERSION" \
            -d environment="staging" \
            -d url="${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}"
```

---

## Risk Assessment & Mitigation

| Risk                              | Impact | Probability | Mitigation                                 |
| --------------------------------- | ------ | ----------- | ------------------------------------------ |
| **Sentry DSN leaked**             | High   | Medium      | Use secrets manager, never commit to git   |
| **Quota exceeded**                | Medium | Low         | Smart sampling, set alerts                 |
| **gRPC trace propagation breaks** | Medium | Low         | Comprehensive E2E tests                    |
| **Performance overhead**          | Low    | Very Low    | Async sending, sampling                    |
| **Sensitive data leakage**        | High   | Low         | Multi-layer scrubbing, test with real data |

---

## Success Metrics

| Metric                 | Target                    | Measurement                    |
| ---------------------- | ------------------------- | ------------------------------ |
| **Error capture rate** | >95%                      | Test by triggering panics      |
| **Trace continuity**   | 100% across services      | Verify waterfalls in Sentry UI |
| **PII scrubbing**      | 0 wallet addresses leaked | Manual review of test events   |
| **Performance impact** | <5ms p99 latency          | Load tests before/after        |
| **Setup time**         | <4 hours                  | Time from DSN to first trace   |

---

## Implementation Checklist

### Core Infrastructure

- [ ] Create `shared/observability/` package structure
- [ ] Implement `sentry/sentry.go` with Init/Flush
- [ ] Implement `sentry/scrubber.go` with regex patterns
- [ ] Add unit tests for scrubbing logic

### Middleware Layer

- [ ] Implement `middleware/http.go` for Chi
- [ ] Implement `middleware/grpc.go` with interceptors
- [ ] Implement `middleware/graphql.go` for gqlgen
- [ ] Add middleware tests

### Service Integration

- [ ] Update `auth-service/cmd/main.go`
- [ ] Update `graphql-gateway/cmd/main.go`
- [ ] Update `user-service/cmd/main.go`
- [ ] Update `wallet-service/cmd/main.go`
- [ ] Add Sentry config to all services

### Distributed Tracing

- [ ] Implement `tracing/propagator.go`
- [ ] Implement `tracing/sampler.go`
- [ ] Test trace propagation Gateway → gRPC
- [ ] Verify in Sentry UI

### CI/CD

- [ ] Add SENTRY_AUTH_TOKEN to GitHub Secrets
- [ ] Create deploy workflow with release tracking
- [ ] Test deploy notification webhook

### Documentation

- [ ] Add observability README
- [ ] Update DEVELOPMENT.md with Sentry setup
- [ ] Document scrubbing patterns

---

## Next Steps

1. **Create detailed implementation plan** using `/plan` command
2. **Implement in phases** (Core → Middleware → Services → CI/CD)
3. **Test thoroughly** in dev environment before staging
4. **Document** team onboarding procedures

---

## Unresolved Questions

1. **Sentry Project Structure**: Single project for all services or separate per service?

   - _Recommendation_: Single project with `service` tag for simpler billing

2. **Sampling Strategy**: Should authentication mutations be sampled at 100%?

   - _Recommendation_: Yes, auth failures are critical

3. **Trace Retention**: How long to keep traces for debugging?
   - _Recommendation_: 30 days for staging, 7 days for dev

---

## Sources

- [Sentry for Go - Official Documentation](https://docs.sentry.io/platforms/go/)
- [Set Up Tracing in Go](https://docs.sentry.io/platforms/go/tracing/)
- [Debugging Microservices with Sentry](https://sentry.io/resources/debugging-microservices-and-distributed-systems/)
- [Distributed Tracing Fundamentals](https://sentry.io/resources/what-is-distributed-tracing/)
- [Golang, OpenTelemetry, and Sentry](https://levelup.gitconnected.com/golang-opentelemetry-and-sentry-the-underrated-distributed-tracing-stack-69dcda886ffe)

---

**Status**: ✅ Ready for implementation planning
**Estimated Effort**: 4-6 hours for full integration
**Risk Level**: Low (greenfield project, no legacy dependencies)
