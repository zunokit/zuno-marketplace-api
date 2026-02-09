# gRPC Observability Interceptor Research

**Date:** 2026-02-04
**Focus:** Interceptor patterns, timeout management, and Sentry integration for collection-service

---

## Current State Analysis

### Collection Service (cmd/main.go)
```go
// NO interceptors configured - bare gRPC server
grpcServer := grpc.NewServer(
    grpc.MaxRecvMsgSize(10*1024*1024),
    grpc.MaxSendMsgSize(10*1024*1024),
)
```

### Other Services Pattern (auth/wallet)
```go
// Use ChainUnaryInterceptor with shared middleware
grpcServer := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        grpcMiddleware.UnaryServerInterceptor(),
    ),
)
```

**Key Gap:** Collection service lacks observability interceptors present in other services.

---

## 1. Interceptor Architecture

### UnaryServerInterceptor Pattern
```go
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{},
        info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

        // Extract sentry-trace header from metadata
        spanOpts := []sentry.SpanOption{
            sentry.WithOpName("grpc.server"),
            sentry.WithDescription(info.FullMethod),
        }

        // Continue parent span from trace header
        if md, ok := metadata.FromIncomingContext(ctx); ok {
            if values := md["sentry-trace"]; len(values) > 0 {
                spanOpts = append(spanOpts, obsTrace.ContinueFromTraceHeader(values[0]))
            }
        }

        span := sentry.StartSpan(ctx, info.FullMethod, spanOpts...)
        defer span.Finish()

        ctx = span.Context()
        resp, err := handler(ctx, req)

        if err != nil {
            span.SetData("grpc.error", err.Error())
            span.Status = sentry.SpanStatusInternalError
        }

        return resp, err
    }
}
```

### StreamServerInterceptor Pattern
```go
func StreamServerInterceptor() grpc.StreamServerInterceptor {
    return func(srv interface{}, ss grpc.ServerStream,
        info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

        ctx := ss.Context()
        span := sentry.StartSpan(ctx, info.FullMethod,
            sentry.WithOpName("grpc.server.stream"))
        defer span.Finish()

        // Wrap stream with span context
        wrappedStream := &streamWithContext{
            ServerStream: ss,
            ctx:          span.Context(),
        }

        return handler(srv, wrappedStream)
    }
}
```

### ChainUnaryInterceptor Pattern
```go
grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        TimeoutInterceptor(30*time.Second),       // Per-method timeout
        grpcMiddleware.UnaryServerInterceptor(),  // Sentry tracing
        LoggingInterceptor(),                     // Request logging
    ),
)
```

---

## 2. Sentry Integration

### Distributed Tracing Setup
```go
// Trace header format: {trace_id}-{span_id}-{sampled}
// Example: 12345678901234567890123456789012-1234567890123456-1

func ContinueFromTraceHeader(header string) sentry.SpanOption {
    return func(span *sentry.Span) {
        parts := strings.Split(header, "-")
        if len(parts) != 3 {
            return
        }

        // Parse trace ID (32 hex chars) and parent span ID (16 hex chars)
        traceIDBytes, _ := hex.DecodeString(parts[0])
        parentSpanIDBytes, _ := hex.DecodeString(parts[1])

        var traceID sentry.TraceID
        copy(traceID[:], traceIDBytes)

        var parentSpanID sentry.SpanID
        copy(parentSpanID[:], parentSpanIDBytes)

        span.TraceID = traceID
        span.ParentSpanID = parentSpanID
    }
}
```

### Init Pattern (from auth/wallet services)
```go
obs.Init(
    cfg.Sentry.DSN,
    cfg.Sentry.Environment,
    "collection-service",
    getBuildVersion(),
    obsTrace.GetTracesSampleRate(cfg.Sentry.Environment),
)
defer obs.Flush(2 * time.Second)
```

---

## 3. Timeout Management

### Context Propagation Pattern
```go
func (s *CollectionServer) ProcessIndexerWebhook(
    ctx context.Context, req *pb.ProcessIndexerWebhookRequest) (*pb.ProcessIndexerWebhookResponse, error) {

    // Check incoming deadline
    if deadline, ok := ctx.Deadline(); ok {
        remaining := time.Until(deadline)
        s.logger.Debug("Webhook deadline", zap.Duration("remaining", remaining))
    }

    // Database operations automatically inherit deadline
    processed, err := s.processedEventRepo.IsEventProcessed(ctx, eventID)

    return response, nil
}
```

### Server Timeout Interceptor
```go
func ServerTimeoutInterceptor(maxDuration time.Duration) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{},
        info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

        // Check client deadline
        clientDeadline, hasClientDeadline := ctx.Deadline()
        serverDeadline := time.Now().Add(maxDuration)

        // Use earlier deadline
        var effectiveDeadline time.Time
        if hasClientDeadline && clientDeadline.Before(serverDeadline) {
            effectiveDeadline = clientDeadline
        } else {
            effectiveDeadline = serverDeadline
        }

        ctx, cancel := context.WithDeadline(ctx, effectiveDeadline)
        defer cancel()

        // Execute with timeout
        done := make(chan result, 1)
        go func() {
            resp, err := handler(ctx, req)
            done <- result{resp, err}
        }()

        select {
        case r := <-done:
            return r.resp, r.err
        case <-ctx.Done():
            if ctx.Err() == context.DeadlineExceeded {
                return nil, status.Error(codes.DeadlineExceeded, "server timeout exceeded")
            }
            return nil, status.FromContextError(ctx.Err()).Err()
        }
    }
}
```

### Time Budget Allocation Pattern
```go
func allocateTime(ctx context.Context, fraction float64) (context.Context, context.CancelFunc) {
    deadline, ok := ctx.Deadline()
    if !ok {
        return context.WithTimeout(ctx, 10*time.Second)
    }

    remaining := time.Until(deadline)
    allocated := time.Duration(float64(remaining) * fraction)

    // Ensure minimum timeout
    if allocated < 100*time.Millisecond {
        allocated = 100 * time.Millisecond
    }

    return context.WithTimeout(ctx, allocated)
}
```

---

## 4. Testing Patterns

### Unit Test (from middleware_test.go)
```go
func TestUnaryServerInterceptor(t *testing.T) {
    initTestSentry()
    defer sentry.Flush(1)

    interceptor := UnaryServerInterceptor()

    handlerCalled := false
    unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
        handlerCalled = true
        return "response", nil
    }

    info := &grpc.UnaryServerInfo{
        FullMethod: "/test.Service/Method",
        Server:     "test-server",
    }

    resp, err := interceptor(context.Background(), "request", info, unaryHandler)

    assert.NoError(t, err)
    assert.True(t, handlerCalled)
    assert.Equal(t, "response", resp)
}
```

### Trace Propagation Test
```go
func TestGRPCTracePropagation(t *testing.T) {
    initTestSentry()
    defer sentry.Flush(1)

    // Create parent span
    parentCtx := context.Background()
    parentSpan := sentry.StartSpan(parentCtx, "parent")
    defer parentSpan.Finish()

    // Generate trace header
    traceHeader := obsTrace.FormatTraceHeader(parentSpan)

    var capturedTraceID sentry.TraceID
    interceptor := UnaryServerInterceptor()

    unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
        span := sentry.SpanFromContext(ctx)
        if span != nil {
            capturedTraceID = span.TraceID
        }
        return "response", nil
    }

    // Create context with sentry-trace header
    md := metadata.Pairs("sentry-trace", traceHeader)
    ctx := metadata.NewIncomingContext(context.Background(), md)

    interceptor(ctx, "request", &grpc.UnaryServerInfo{FullMethod: "/test"}, unaryHandler)

    // Verify trace ID matches parent
    assert.Equal(t, parentSpan.TraceID, capturedTraceID)
}
```

### Integration Test Pattern
```go
func TestWebhookWithTimeout(t *testing.T) {
    // Setup server with timeout interceptor
    server := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            ServerTimeoutInterceptor(5*time.Second),
            grpcMiddleware.UnaryServerInterceptor(),
        ),
    )

    // Create client with deadline
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    resp, err := client.ProcessIndexerWebhook(ctx, req)

    // Should succeed within deadline
    assert.NoError(t, err)
    assert.True(t, resp.Success)
}
```

---

## 5. Recommended Collection Service Setup

```go
// cmd/main.go

// 1. Initialize Sentry (non-blocking)
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
}

// 2. Create gRPC server with interceptors
grpcServer := grpc.NewServer(
    grpc.MaxRecvMsgSize(10*1024*1024),
    grpc.MaxSendMsgSize(10*1024*1024),
    grpc.ChainUnaryInterceptor(
        grpcMiddleware.UnaryServerInterceptor(),  // Sentry tracing
    ),
)

// 3. Add timeout interceptor for webhook endpoint
// (Per-method timeouts can be configured in interceptor)
```

---

## Key Findings

1. **Collection service lacks observability:** No Sentry interceptors, unlike auth/wallet services
2. **Shared middleware exists:** `shared/observability/middleware/grpc.go` provides production-ready interceptors
3. **Timeout propagation automatic:** Go gRPC automatically propagates deadlines through context
4. **Testing patterns established:** Comprehensive unit tests in `middleware_test.go`
5. **Webhook handler needs timeout awareness:** Currently no explicit deadline checking

---

## Implementation Priority

1. **Add Sentry interceptor** (critical) - Same pattern as auth/wallet services
2. **Add timeout interceptor** (high) - Prevent resource exhaustion
3. **Add deadline checks** (medium) - Check ctx.Done() in long-running operations
4. **Add interceptor tests** (low) - Follow existing test patterns

---

## Unresolved Questions

- Should collection-service use per-method timeout configuration?
- What's the appropriate timeout for webhook processing (currently none)?
- Should webhooks have different timeout than regular RPCs?

---

## Sources

- [Deadlines - gRPC Official Docs](https://grpc.io/docs/guides/deadlines/)
- [How to Handle Deadlines and Timeouts in gRPC - OneUptime](https://oneuptime.com/blog/post/2026-01-08-grpc-deadlines-timeouts/view)
- [gRPC Go Interceptors - Victoriametrics](https://victoriametrics.com/blog/go-grpc-basic-streaming-interceptor/)
