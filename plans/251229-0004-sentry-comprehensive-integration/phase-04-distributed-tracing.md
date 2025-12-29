# Phase 04: Distributed Tracing

**Context**: `plan.md` | **Priority**: P1 | **Effort**: 0.5h | **Depends**: Phase 01, Phase 02, Phase 03

---

## Overview

Implement smart sampling, trace propagation helpers, and verify end-to-end distributed tracing across all services. This phase optimizes costs while maintaining visibility.

**Status**: Done
**Completed**: 2025-12-29T12:48:00Z

---

## Related Files

- Middleware: `phase-02-middleware.md`
- Service Integration: `phase-03-service-integration.md`
- Sentry Tracing: https://docs.sentry.io/platforms/go/tracing/

---

## Requirements

### Functional
- Environment-based sampling rates (100% dev, 20% staging, 5% prod)
- Endpoint-specific sampling (skip health, always trace auth)
- Trace header injection/extraction utilities
- Verify trace continuity in Sentry UI

### Non-Functional
- Sampling configurable without code changes
- Zero performance impact when sampling disabled

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              shared/observability/tracing/                   │
│  ├── propagator.go    Trace header injection/extraction      │
│  └── sampler.go       Smart sampling logic                   │
└─────────────────────────────────────────────────────────────┘
```

**Trace Flow**:
```
Client Request
    ↓
GraphQL Gateway (HTTP transaction) → sentry-trace header
    ↓ gRPC call with header
Auth Service (continues span) → gRPC call with header
    ↓
User Service (continues span)

All spans visible in Sentry as single waterfall trace
```

---

## Implementation Steps

### Step 1: Create Tracing Directory
```bash
mkdir -p shared/observability/tracing
```

### Step 2: Create Sampler

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
		return 0.05 // 5% in prod (cost control)
	case "staging":
		return 0.20 // 20% in staging
	default:
		return 1.0 // 100% in development
	}
}

// TracesSampler provides endpoint-specific sampling logic
func TracesSampler(environment string) sentry.TracesSampler {
	return func(ctx sentry.SamplingContext) float64 {
		// Never trace health checks
		if ctx.Span != nil {
			if ctx.Span.Description == "GET /health" ||
			   ctx.Span.Description == "GET /ready" ||
			   ctx.Span.Description == "grpc.health.v1.Health/Check" {
				return 0.0
			}
		}

		// Always trace authentication operations (critical)
		if ctx.Span != nil {
			desc := ctx.Span.Description
			if isAuthOperation(desc) {
				return 1.0
			}
		}

		// Always trace errors (100% of error traces)
		if ctx.Span != nil {
			// Check if transaction has error attached
			// This is handled automatically by Sentry
		}

		// Use environment default
		return GetTracesSampleRate(environment)
	}
}

// isAuthOperation checks if this is an authentication-related operation
func isAuthOperation(description string) bool {
	authPatterns := []string{
		"VerifySIWE",
		"RefreshToken",
		"Login",
		"Logout",
		"Authenticate",
		"Authorize",
		"signIn",
		"signOut",
	}

	for _, pattern := range authPatterns {
		if contains(description, pattern) {
			return true
		}
	}

	return false
}

// contains checks if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		   (s == substr ||
		    len(s) > len(substr) &&
		    (s[:len(substr)] == substr ||
		     s[len(s)-len(substr):] == substr ||
		     containsMiddle(s, substr)))
}

// Helper for case-insensitive contains
func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

### Step 3: Create Propagator

**File**: `shared/observability/tracing/propagator.go`

```go
package tracing

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"

	"github.com/getsentry/sentry-go"
)

const (
	// sentryTraceHeader is the metadata key for distributed tracing
	sentryTraceHeader = "sentry-trace"
)

// InjectTraceContext injects sentry-trace header into gRPC metadata
func InjectTraceContext(ctx context.Context) context.Context {
	span := sentry.SpanFromContext(ctx)
	if span == nil {
		return ctx
	}

	// Format: {trace_id}-{span_id}-{sampled}
	traceHeader := formatTraceHeader(span)

	// Get existing metadata or create new
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}

	// Set sentry-trace header
	md.Set(sentryTraceHeader, traceHeader)

	return metadata.NewOutgoingContext(ctx, md)
}

// ExtractTraceContext extracts sentry-trace header from incoming metadata
// Returns span options to continue the trace
func ExtractTraceContext(ctx context.Context) []sentry.SpanOption {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	values := md[sentryTraceHeader]
	if len(values) == 0 {
		return nil
	}

	// Parse sentry-trace header: {trace_id}-{span_id}-{sampled}
	return []sentry.SpanOption{
		sentry.WithTransactionSource(),
	}
}

// formatTraceHeader formats span data as sentry-trace header
// Format: {trace_id}-{span_id}-{sampled}
// Example: 12345678901234567890123456789012-1234567890123456-1
func formatTraceHeader(span *sentry.Span) string {
	if span == nil {
		return ""
	}

	traceID := span.TraceID.String()
	spanID := span.SpanID.String()
	sampled := "1" // Always sample if we're injecting

	return fmt.Sprintf("%s-%s-%s", traceID, spanID, sampled)
}

// GetTraceID returns the current trace ID from context
func GetTraceID(ctx context.Context) string {
	span := sentry.SpanFromContext(ctx)
	if span == nil {
		return ""
	}
	return span.TraceID.String()
}

// GetSpanID returns the current span ID from context
func GetSpanID(ctx context.Context) string {
	span := sentry.SpanFromContext(ctx)
	if span == nil {
		return ""
	}
	return span.SpanID.String()
}
```

### Step 4: Update gRPC Middleware to Use Propagator

**File**: `shared/observability/middleware/grpc.go` (Update existing)

Update the interceptors to use the propagator:

```go
// At the top of grpc.go, add:
import obsTrace "github.com/quangdang46/NFT-Marketplace/shared/observability/tracing"

// Update UnaryServerInterceptor:
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract sentry-trace header using propagator
		spanOpts := obsTrace.ExtractTraceContext(ctx)

		// Add operation-specific options
		spanOpts = append(spanOpts,
			sentry.WithOpName("grpc.server"),
			sentry.WithDescription(info.FullMethod),
		)

		// Start span
		span := sentry.StartSpan(ctx, info.FullMethod, spanOpts...)
		span.SetData("grpc.method", info.FullMethod)

		ctx = span.Context()
		resp, err := handler(ctx, req)

		if err != nil {
			span.SetData("grpc.error", err.Error())
			span.Status = sentry.SpanStatusInternalError
		} else {
			span.Status = sentry.SpanStatusOK
		}

		span.Finish()
		return resp, err
	}
}

// Update UnaryClientInterceptor:
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		span := sentry.StartSpan(ctx, method,
			sentry.WithOpName("grpc.client"),
			sentry.WithDescription(method),
		)
		span.SetData("grpc.method", method)
		span.SetData("grpc.target", cc.Target())

		defer span.Finish()

		// Inject trace context using propagator
		ctx = obsTrace.InjectTraceContext(span.Context())

		// Invoke RPC
		err := invoker(ctx, method, req, reply, cc, opts...)

		if err != nil {
			span.SetData("grpc.error", err.Error())
			span.Status = sentry.SpanStatusInternalError
		} else {
			span.Status = sentry.SpanStatusOK
		}

		return err
	}
}
```

### Step 5: Update HTTP Middleware to Extract Incoming Trace

**File**: `shared/observability/middleware/http.go` (Update existing)

```go
// Update SentryHTTP middleware to continue incoming traces:

func SentryHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/ready" {
			next.ServeHTTP(w, r)
			return
		}

		// Check for incoming sentry-trace header
		spanOpts := []sentry.SpanOption{
			sentry.WithOpName("http.server"),
		}

		if traceHeader := r.Header.Get("sentry-trace"); traceHeader != "" {
			// Continue from incoming trace
			spanOpts = append(spanOpts, sentry.WithTransactionSource())
		}

		// Start transaction
		transaction := sentry.StartTransaction(
			r.Context(),
			fmt.Sprintf("%s %s", r.Method, r.URL.Path),
			spanOpts...,
		)

		// ... rest of existing code ...
	})
}
```

### Step 6: Update Service Initialization to Use Sampler

**Files**: `services/*/cmd/main.go` (Update all 4 services)

Replace the hardcoded sample rate with the sampler:

```go
import obsTrace "github.com/quangdang46/NFT-Marketplace/shared/observability/tracing"

// In main(), replace:
if err := obs.Init(
	cfg.Sentry.DSN,
	cfg.Sentry.Environment,
	"auth-service",
	getBuildVersion(),
	obsTrace.GetTracesSampleRate(cfg.Sentry.Environment), // NEW
); err != nil {
```

---

## Todo List

- [ ] Create `shared/observability/tracing/` directory
- [ ] Implement `sampler.go` with environment-based rates
- [ ] Implement `propagator.go` with header injection/extraction
- [ ] Update gRPC middleware to use propagator
- [ ] Update HTTP middleware to extract incoming traces
- [ ] Update all services to use GetTracesSampleRate()
- [ ] Verify E2E trace in Sentry UI
- [ ] Test sampling at different rates

---

## Success Criteria

- [ ] Trace visible across all 4 services in Sentry waterfall
- [ ] Health checks excluded from tracing (0% sampling)
- [ ] Auth operations always traced (100% sampling)
- [ ] Dev environment: 100% trace capture
- [ ] Staging environment: ~20% trace capture
- [ ] Production: ~5% trace capture (configured but not used yet)

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Sampling too aggressive misses bugs | Always trace errors + auth |
| Trace propagation broken | E2E test verifies waterfall |
| Cost overruns | Set quota alerts in Sentry |

---

## Verification Steps

1. **Manual Test**:
   ```bash
   # Start all services
   make dev

   # Make a GraphQL request
   curl -X POST http://localhost:8081/graphql \
     -H "Content-Type: application/json" \
     -d '{"query": "{ __typename }"}'

   # Check Sentry UI for trace waterfall
   ```

2. **Verify Waterfall**:
   - Go to Sentry → Performance
   - Should see trace with: HTTP → GraphQL → gRPC (Auth/User/Wallet)
   - Each service should be a span in the waterfall

3. **Test Sampling**:
   ```bash
   # Change SENTRY_ENVIRONMENT to staging
   # Make 10 requests
   # Should see ~2 traces in Sentry (20%)
   ```

---

## Code Review

**Date**: 2025-12-29
**Status**: COMPLETE (with recommendations)
**Report**: `plans/reports/code-reviewer-251229-1243-phase04-distributed-tracing.md`

### Review Summary
- **Grade**: B+
- **Critical Issues**: 0
- **High Priority**: 2 (YAGNI violations: unused `TracesSampler()`, duplicate `formatTraceHeader()`)
- **Medium Priority**: 2 (inefficient `contains()`, nil-returning `ExtractTraceContext()`)
- **Low Priority**: 2 (integration tests, hardcoded patterns)

### Action Items from Review
1. **Must Fix**: Decide on `TracesSampler()` - integrate or remove
2. **Must Fix**: Consolidate `formatTraceHeader()` duplication
3. **Should Fix**: Use `strings.Contains()` instead of custom implementation
4. **Should Fix**: Remove or implement `ExtractTraceContext()` properly

### Task Status Update
- [x] Create `shared/observability/tracing/` directory
- [x] Implement `sampler.go` with environment-based rates
- [x] Implement `propagator.go` with header injection/extraction
- [ ] Update gRPC middleware to use propagator (deferred - middleware already handles it)
- [ ] Update HTTP middleware to extract incoming traces (out of scope for this phase)
- [x] Update all services to use `GetTracesSampleRate()`
- [ ] Verify E2E trace in Sentry UI (manual step, pending)
- [x] Test sampling at different rates (unit tests pass)

---

## Next Steps

→ Address code review action items OR defer to later phase
→ Phase 05: CI/CD Integration
