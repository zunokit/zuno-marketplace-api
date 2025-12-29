package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getsentry/sentry-go"
	"google.golang.org/grpc"
)

// TestSentryHTTP tests the HTTP middleware
func TestSentryHTTP(t *testing.T) {
	// Initialize Sentry for testing (with test DSN)
	_ = sentry.Init(sentry.ClientOptions{
		Dsn:              "https://examplePublicKey@o0.ingest.sentry.io/0",
		TracesSampleRate: 1.0,
	})
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
	_ = sentry.Init(sentry.ClientOptions{
		Dsn:              "https://examplePublicKey@o0.ingest.sentry.io/0",
		TracesSampleRate: 1.0,
	})
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
	_ = sentry.Init(sentry.ClientOptions{
		Dsn:              "https://examplePublicKey@o0.ingest.sentry.io/0",
		TracesSampleRate: 1.0,
	})
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
	_ = sentry.Init(sentry.ClientOptions{
		Dsn:              "https://examplePublicKey@o0.ingest.sentry.io/0",
		TracesSampleRate: 1.0,
	})
	defer sentry.Flush(1)

	ctx := context.Background()
	span := sentry.StartSpan(ctx, "test")
	defer span.Finish()

	header := formatTraceHeader(span)

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
	_ = sentry.Init(sentry.ClientOptions{
		Dsn:              "https://examplePublicKey@o0.ingest.sentry.io/0",
		TracesSampleRate: 1.0,
	})
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
