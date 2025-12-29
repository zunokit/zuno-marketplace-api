package tracing

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestExtractTraceContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantNil bool
	}{
		{
			name:    "no metadata returns nil",
			ctx:     context.Background(),
			wantNil: true,
		},
		{
			name:    "empty metadata returns nil",
			ctx:     metadata.NewIncomingContext(context.Background(), metadata.New(nil)),
			wantNil: true,
		},
		{
			name: "with sentry-trace header returns nil (middleware handles continuation)",
			ctx: metadata.NewIncomingContext(context.Background(),
				metadata.Pairs("sentry-trace", "12345678901234567890123456789012-1234567890123456-1")),
			wantNil: true, // Function returns nil since middleware handles actual continuation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTraceContext(tt.ctx)
			if (result == nil) != tt.wantNil {
				t.Errorf("ExtractTraceContext() = %v, wantNil %v", result, tt.wantNil)
			}
		})
	}
}

func TestInjectTraceContext(t *testing.T) {
	ctx := context.Background()

	// Without span in context, should return same context
	result := InjectTraceContext(ctx)
	if result != ctx {
		t.Error("InjectTraceContext() without span should return same context")
	}

	// With metadata already in context (still no span, so returns same context)
	md := metadata.New(map[string]string{"key": "value"})
	ctx = metadata.NewOutgoingContext(ctx, md)

	result = InjectTraceContext(ctx)
	// Without a sentry span in context, returns same context
	if result == nil && ctx != nil {
		t.Error("InjectTraceContext() should not return nil")
	}
}

func TestFormatTraceHeader(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "nil span returns empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTraceHeader(nil)
			if result != "" {
				t.Errorf("formatTraceHeader(nil) = %q, want empty", result)
			}
		})
	}
}

func TestGetTraceID(t *testing.T) {
	ctx := context.Background()

	// Without span in context
	result := GetTraceID(ctx)
	if result != "" {
		t.Errorf("GetTraceID() = %q, want empty", result)
	}
}

func TestGetSpanID(t *testing.T) {
	ctx := context.Background()

	// Without span in context
	result := GetSpanID(ctx)
	if result != "" {
		t.Errorf("GetSpanID() = %q, want empty", result)
	}
}
