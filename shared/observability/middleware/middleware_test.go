package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getsentry/sentry-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	obsTrace "github.com/zunokit/zuno-marketplace-api/shared/observability/tracing"
)

// initTestSentry initializes Sentry for testing with disabled transport
// Using empty DSN prevents actual network calls while testing middleware logic
func initTestSentry() {
	_ = sentry.Init(sentry.ClientOptions{
		Dsn:              "", // Empty DSN disables event sending
		TracesSampleRate: 1.0,
	})
}

// TestSentryHTTP tests the HTTP middleware
func TestSentryHTTP(t *testing.T) {
	// Initialize Sentry for testing
	initTestSentry()
	defer sentry.Flush(1)

	tests := []struct {
		name           string
		path           string
		method         string
		expectSkip     bool
		expectedStatus int
	}{
		{
			name:           "health endpoint skipped",
			path:           "/health",
			method:         "GET",
			expectSkip:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ready endpoint skipped",
			path:           "/ready",
			method:         "GET",
			expectSkip:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "regular endpoint traced",
			path:           "/api/v1/users",
			method:         "GET",
			expectSkip:     false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "post request traced",
			path:           "/api/v1/users",
			method:         "POST",
			expectSkip:     false,
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.expectedStatus)
				w.Write([]byte("OK"))
			})

			// Wrap with Sentry middleware
			mw := SentryHTTP(handler)

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Execute
			mw.ServeHTTP(w, req)

			// Verify response
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestResponseWriter tests the responseWriter wrapper
func TestResponseWriter(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{name: "200 OK", statusCode: http.StatusOK},
		{name: "404 Not Found", statusCode: http.StatusNotFound},
		{name: "500 Internal Server Error", statusCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			rw.WriteHeader(tt.statusCode)

			if rw.status != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, rw.status)
			}
		})
	}
}

// TestUnaryServerInterceptor tests the gRPC server interceptor
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

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !handlerCalled {
		t.Error("handler was not called")
	}

	if resp != "response" {
		t.Errorf("expected response 'response', got %v", resp)
	}
}

// TestUnaryClientInterceptor tests the gRPC client interceptor
func TestUnaryClientInterceptor(t *testing.T) {
	initTestSentry()
	defer sentry.Flush(1)

	interceptor := UnaryClientInterceptor()

	invokerCalled := false
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		invokerCalled = true
		return nil
	}

	// Mock connection (we won't actually call it)
	ctx := context.Background()

	err := interceptor(ctx, "/test.Service/Method", "req", "reply", nil, invoker)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !invokerCalled {
		t.Error("invoker was not called")
	}
}

// TestFormatTraceHeader tests trace header formatting
func TestFormatTraceHeader(t *testing.T) {
	initTestSentry()
	defer sentry.Flush(1)

	ctx := context.Background()
	span := sentry.StartSpan(ctx, "test")
	defer span.Finish()

	header := obsTrace.FormatTraceHeader(span)

	if header == "" {
		t.Error("expected non-empty trace header")
	}

	// Header format should be: {trace_id}-{span_id}-{sampled}
	if len(header) < 10 {
		t.Errorf("trace header too short: %s", header)
	}
}

// TestStreamServerInterceptor tests the gRPC streaming server interceptor
func TestStreamServerInterceptor(t *testing.T) {
	initTestSentry()
	defer sentry.Flush(1)

	interceptor := StreamServerInterceptor()

	// Mock server stream
	mockStream := &mockServerStream{
		ctx: context.Background(),
	}

	handlerCalled := false
	streamHandler := func(srv interface{}, ss grpc.ServerStream) error {
		handlerCalled = true
		return nil
	}

	info := &grpc.StreamServerInfo{
		FullMethod: "/test.Service/StreamMethod",
	}

	err := interceptor(nil, mockStream, info, streamHandler)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !handlerCalled {
		t.Error("stream handler was not called")
	}
}

// mockServerStream is a mock implementation of grpc.ServerStream
type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

// TestHTTPTracePropagation tests that HTTP middleware continues trace from header
func TestHTTPTracePropagation(t *testing.T) {
	initTestSentry()
	defer sentry.Flush(1)

	// Create a parent span to get a valid trace ID
	parentCtx := context.Background()
	parentSpan := sentry.StartSpan(parentCtx, "parent")
	defer parentSpan.Finish()

	// Generate trace header from parent span
	traceHeader := obsTrace.FormatTraceHeader(parentSpan)

	// Create test handler that validates trace continuation
	var capturedTraceID sentry.TraceID
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the transaction from context
		span := sentry.SpanFromContext(r.Context())
		if span != nil {
			capturedTraceID = span.TraceID
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with Sentry middleware
	mw := SentryHTTP(handler)

	// Create request with sentry-trace header
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("sentry-trace", traceHeader)
	w := httptest.NewRecorder()

	// Execute
	mw.ServeHTTP(w, req)

	// Verify trace ID matches parent (trace propagation works)
	if capturedTraceID.String() == "" {
		t.Error("expected non-empty trace ID in child span")
	}

	// The child should have the same trace ID as the parent
	if capturedTraceID != parentSpan.TraceID {
		t.Errorf("trace propagation failed: expected trace ID %s, got %s",
			parentSpan.TraceID, capturedTraceID)
	}
}

// TestGRPCTracePropagation tests that gRPC server interceptor continues trace from header
func TestGRPCTracePropagation(t *testing.T) {
	initTestSentry()
	defer sentry.Flush(1)

	// Create a parent span to get a valid trace ID
	parentCtx := context.Background()
	parentSpan := sentry.StartSpan(parentCtx, "parent")
	defer parentSpan.Finish()

	// Generate trace header from parent span
	traceHeader := obsTrace.FormatTraceHeader(parentSpan)

	var capturedTraceID sentry.TraceID
	var capturedParentSpanID sentry.SpanID

	interceptor := UnaryServerInterceptor()
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Get the span from context
		span := sentry.SpanFromContext(ctx)
		if span != nil {
			capturedTraceID = span.TraceID
			capturedParentSpanID = span.ParentSpanID
		}
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
		Server:     "test-server",
	}

	// Create context with sentry-trace header in metadata
	md := metadata.Pairs("sentry-trace", traceHeader)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	// Execute interceptor
	resp, err := interceptor(ctx, "request", info, unaryHandler)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if resp != "response" {
		t.Errorf("expected response 'response', got %v", resp)
	}

	// Verify trace ID matches parent (trace propagation works)
	if capturedTraceID.String() == "" {
		t.Error("expected non-empty trace ID in child span")
	}

	if capturedTraceID != parentSpan.TraceID {
		t.Errorf("trace propagation failed: expected trace ID %s, got %s",
			parentSpan.TraceID, capturedTraceID)
	}

	// Verify parent span ID matches
	if capturedParentSpanID != parentSpan.SpanID {
		t.Errorf("parent span ID mismatch: expected %s, got %s",
			parentSpan.SpanID, capturedParentSpanID)
	}
}

// TestTraceHeaderFormat tests that trace header is correctly formatted
func TestTraceHeaderFormat(t *testing.T) {
	initTestSentry()
	defer sentry.Flush(1)

	ctx := context.Background()
	span := sentry.StartSpan(ctx, "test")
	defer span.Finish()

	header := obsTrace.FormatTraceHeader(span)

	// Header format: {trace_id}-{span_id}-{sampled}
	// Example: 12345678901234567890123456789012-1234567890123456-1
	parts := strings.Split(header, "-")
	if len(parts) != 3 {
		t.Errorf("expected 3 parts in trace header, got %d: %s", len(parts), header)
	}

	// Check trace ID (32 hex chars)
	if len(parts[0]) != 32 {
		t.Errorf("expected trace ID length 32, got %d: %s", len(parts[0]), parts[0])
	}

	// Check span ID (16 hex chars)
	if len(parts[1]) != 16 {
		t.Errorf("expected span ID length 16, got %d: %s", len(parts[1]), parts[1])
	}

	// Check sampled flag
	if parts[2] != "1" {
		t.Errorf("expected sampled flag '1', got %s", parts[2])
	}
}
