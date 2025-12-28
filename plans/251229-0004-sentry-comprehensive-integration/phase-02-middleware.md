# Phase 02: Middleware Layer

**Context**: `plan.md` | **Priority**: P1 | **Effort**: 1.5h | **Depends**: Phase 01

---

## Overview

Create middleware for HTTP (Chi), gRPC (server/client), and GraphQL (gqlgen) to automatically capture transactions, propagate traces, and instrument request handling.

**Status**: Pending

---

## Related Files

- Core Package: `phase-01-core-package.md`
- Chi Router: https://github.com/go-chi/chi
- gRPC Interceptors: https://grpc.io/docs/labs/go-interceptors/
- gqlgen Middleware: https://gqlgen.com/reference/datamodel

---

## Requirements

### Functional
- HTTP middleware for Chi router (GraphQL Gateway)
- gRPC server interceptor for unary calls
- gRPC client interceptor for outbound calls
- GraphQL field middleware for resolver instrumentation
- Extract and inject `sentry-trace` header for distributed tracing
- Skip tracing for health endpoints

### Non-Functional
- Minimal overhead (<1ms)
- No breaking changes to existing middleware chains

---

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    shared/observability/middleware/           │
│  ├── http.go         Chi HTTP middleware                      │
│  ├── grpc.go         gRPC server + client interceptors        │
│  └── graphql.go      gqlgen field middleware                  │
└──────────────────────────────────────────────────────────────┘
```

**Trace Propagation Flow**:
```
HTTP Request → Chi MW → Transaction → sentry-trace header
                                    ↓
gRPC Client → Inject Header → gRPC Server → Extract → Continue Span
                                    ↓
                             GraphQL Resolver → Child Span
```

---

## Implementation Steps

### Step 1: Create Middleware Directory
```bash
mkdir -p shared/observability/middleware
```

### Step 2: Create HTTP Middleware

**File**: `shared/observability/middleware/http.go`

```go
package middleware

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
)

// SentryHTTP is Chi middleware that captures HTTP requests as Sentry transactions
func SentryHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip health check tracing
		if r.URL.Path == "/health" || r.URL.Path == "/ready" {
			next.ServeHTTP(w, r)
			return
		}

		// Start transaction
		transaction := sentry.StartTransaction(
			r.Context(),
			fmt.Sprintf("%s %s", r.Method, r.URL.Path),
			sentry.WithTransactionName(r.URL.Path),
			sentry.WithOpName("http.server"),
			sentry.WithDescription(fmt.Sprintf("%s request to %s", r.Method, r.URL.Path)),
		)
		defer transaction.Finish()

		// Add HTTP context data
		transaction.SetData("http.method", r.Method)
		transaction.SetData("http.url", r.URL.String())
		transaction.SetData("http.scheme", r.URL.Scheme)
		transaction.SetData("http.host", r.Host)
		transaction.SetData("http.path", r.URL.Path)
		transaction.SetData("http.query", r.URL.RawQuery)
		transaction.SetData("http.remote_addr", r.RemoteAddr)

		// Extract sentry-trace from incoming request for distributed tracing
		if traceHeader := r.Header.Get("sentry-trace"); traceHeader != "" {
			transaction.SetData("sentry.trace", traceHeader)
		}

		// Continue with trace context
		r = r.WithContext(transaction.Context())

		// Wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)

		// Set HTTP status on transaction
		transaction.SetHttpStatus(rw.status)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
```

### Step 3: Create gRPC Interceptors

**File**: `shared/observability/middleware/grpc.go`

```go
package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/getsentry/sentry-go"
)

const (
	// sentryTraceHeader is the header key for distributed tracing
	sentryTraceHeader = "sentry-trace"
)

// UnaryServerInterceptor creates a Sentry span for each incoming gRPC call
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract sentry-trace header from metadata
		spanOpts := []sentry.SpanOption{
			sentry.WithOpName("grpc.server"),
			sentry.WithDescription(info.FullMethod),
		}

		// Check for incoming trace header
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md[sentryTraceHeader]; len(values) > 0 {
				// Continue parent span from trace header
				spanOpts = append(spanOpts, continueFromTraceHeader(values[0]))
			}
		}

		// Start span
		span := sentry.StartSpan(ctx, info.FullMethod, spanOpts...)
		span.SetData("grpc.method", info.FullMethod)
		span.SetData("grpc.service", info.Server)

		// Execute handler with span context
		ctx = span.Context()
		resp, err := handler(ctx, req)

		// Record error if any
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

// UnaryClientInterceptor injects sentry-trace header into outbound gRPC calls
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Start client span
		span := sentry.StartSpan(ctx, method,
			sentry.WithOpName("grpc.client"),
			sentry.WithDescription(method),
		)
		span.SetData("grpc.method", method)
		span.SetData("grpc.target", cc.Target())

		defer func() {
			if span != nil {
				span.Finish()
			}
		}()

		// Inject sentry-trace header into metadata
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		// Copy metadata to avoid modifying original
		md = md.Copy()

		// Add sentry-trace header
		if span.TraceID.String() != "" {
			traceHeader := formatTraceHeader(span)
			md.Set(sentryTraceHeader, traceHeader)
		}

		ctx = metadata.NewOutgoingContext(span.Context(), md)

		// Invoke RPC with updated context
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

// continueFromTraceHeader creates a span option from sentry-trace header
func continueFromTraceHeader(header string) sentry.SpanOption {
	// Parse sentry-trace header format: {trace_id}-{span_id}-{sampled}
	// This is a simplified version - Sentry SDK handles full parsing
	return sentry.WithTransactionSource()
}

// formatTraceHeader formats span data as sentry-trace header
func formatTraceHeader(span *sentry.Span) string {
	// Format: {trace_id}-{span_id}-{sampled}
	// Simplified - production code should use proper Sentry SDK utilities
	return fmt.Sprintf("%s-%s-1",
		span.TraceID.String(),
		span.SpanID.String(),
	)
}

// StreamServerInterceptor creates a Sentry span for streaming gRPC calls
// (Optional - implement if using streaming RPCs)
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		// Start span
		span := sentry.StartSpan(ctx, info.FullMethod,
			sentry.WithOpName("grpc.server.stream"),
			sentry.WithDescription(info.FullMethod),
		)
		span.SetData("grpc.method", info.FullMethod)
		span.SetData("grpc.stream", "true")

		defer span.Finish()

		// Wrap stream with span context
		wrappedStream := &streamWithContext{
			ServerStream: ss,
			ctx:          span.Context(),
		}

		err := handler(srv, wrappedStream)

		if err != nil {
			span.SetData("grpc.error", err.Error())
			span.Status = sentry.SpanStatusInternalError
		}

		return err
	}
}

// streamWithContext wraps ServerStream with custom context
type streamWithContext struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *streamWithContext) Context() context.Context {
	return s.ctx
}
```

### Step 4: Create GraphQL Middleware

**File**: `shared/observability/middleware/graphql.go`

```go
package middleware

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/getsentry/sentry-go"
)

// GraphQLFieldMiddleware creates Sentry spans for each GraphQL field resolution
func GraphQLFieldMiddleware() graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		// Get field context
		fc := graphql.GetFieldContext(ctx)
		if fc == nil {
			return next(ctx)
		}

		// Create span for this field
		span := sentry.StartSpan(ctx,
			fc.Field.Field.Name,
			sentry.WithOpName("graphql.resolve"),
			sentry.WithDescription(fc.Field.ObjectDefinition+"."+fc.Field.Name),
		)

		// Add GraphQL context data
		span.SetData("graphql.field.name", fc.Field.Field.Name)
		span.SetData("graphql.field.type", fc.Field.Definition.Type.String())
		span.SetData("graphql.parent_type", fc.Field.ObjectDefinition)

		// Add operation info if available
		if rc := graphql.GetOperationContext(ctx); rc != nil {
			span.SetData("graphql.operation.name", rc.OperationName)
			span.SetData("graphql.operation.type", rc.Operation.String())

			// Get query variables (sanitized)
			if rc.Variables != nil {
				span.SetData("graphql.variables.count", len(rc.Variables))
			}
		}

		defer func() {
			if r := recover(); r != nil {
				span.SetData("graphql.panic", r)
				span.Status = sentry.SpanStatusInternalError
				span.Finish()
				panic(r)
			}
		}()

		// Resolve field
		result, err := next(span.Context())

		if err != nil {
			span.SetData("graphql.error", err.Error())
			span.Status = sentry.SpanStatusInternalError
		} else {
			span.Status = sentry.SpanStatusOK
		}

		span.Finish()

		return result, err
	}
}

// GraphQLResponseMiddleware captures operation-level metrics
// (Optional - use for query-level insights)
func GraphQLResponseMiddleware() graphql.ResponseMiddleware {
	return func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		// Get operation context
		rc := graphql.GetOperationContext(ctx)
		if rc == nil {
			return next(ctx)
		}

		// Start operation-level span
		span := sentry.StartSpan(ctx,
			rc.OperationName,
			sentry.WithOpName("graphql.operation"),
			sentry.WithDescription(rc.Operation.String()),
		)

		span.SetData("graphql.operation.name", rc.OperationName)
		span.SetData("graphql.operation.type", rc.Operation.String())

		// Execute operation
		response := next(span.Context())

		// Record errors
		if response != nil && len(response.Errors) > 0 {
			span.SetData("graphql.errors.count", len(response.Errors))
			span.Status = sentry.SpanStatusInternalError
		} else {
			span.Status = sentry.SpanStatusOK
		}

		span.Finish()

		return response
	}
}
```

### Step 5: Update shared/observability/README.md

Add middleware usage documentation:

```markdown
### Middleware Usage

#### HTTP Middleware (Chi)

```go
import obshttp "github.com/quangdang46/NFT-Marketplace/shared/observability/middleware"

router := chi.NewRouter()
router.Use(obshttp.SentryHTTP)  // Add BEFORE other middleware
router.Use(middleware.Logger)
router.Use(middleware.Recoverer)
```

#### gRPC Server Interceptor

```go
import obsgrpc "github.com/quangdang46/NFT-Marketplace/shared/observability/middleware"

grpcServer := grpc.NewServer(
	grpc.ChainUnaryInterceptor(
		obsgrpc.UnaryServerInterceptor(),
	),
)
```

#### gRPC Client Interceptor

```go
conn, err := grpc.Dial(
	serviceURL,
	grpc.WithTransportCredentials(insecure.NewCredentials()),
	grpc.WithChainUnaryInterceptor(
		obsgrpc.UnaryClientInterceptor(),
	),
)
```

#### GraphQL Middleware

```go
import obsgraphql "github.com/quangdang46/NFT-Marketplace/shared/observability/middleware"

// In GraphQL server setup
srv := handler.NewDefaultServer(schema)
srv.Use(obsgraphql.GraphQLFieldMiddleware())
srv.Use(obsgraphql.GraphQLResponseMiddleware())
```
```

---

## Todo List

- [ ] Create `shared/observability/middleware/` directory
- [ ] Implement `http.go` with Chi middleware
- [ ] Implement `grpc.go` with server + client interceptors
- [ ] Implement `graphql.go` with field + response middleware
- [ ] Update README.md with middleware usage
- [ ] Write middleware tests
- [ ] Verify compilation with `go build ./...`

---

## Success Criteria

- [ ] HTTP middleware creates transactions for all non-health endpoints
- [ ] gRPC interceptors propagate sentry-trace header correctly
- [ ] GraphQL middleware captures field resolution time
- [ ] All middleware chains work without breaking existing functionality

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Middleware order issues | Document insertion order clearly |
| Trace propagation fails | E2E test across services |
| Performance overhead | Minimal - span creation is lightweight |

---

## Next Steps

→ Phase 03: Service Integration
