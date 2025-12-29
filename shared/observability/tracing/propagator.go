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
// Note: Actual trace continuation handled by middleware continueFromTraceHeader()
func ExtractTraceContext(ctx context.Context) []sentry.SpanOption {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	values := md[sentryTraceHeader]
	if len(values) == 0 {
		return nil
	}

	// Middleware handles actual trace parsing via continueFromTraceHeader()
	// This function returns nil since middleware already handles continuation
	return nil
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
