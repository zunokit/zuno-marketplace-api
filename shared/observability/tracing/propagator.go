package tracing

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

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
	traceHeader := FormatTraceHeader(span)

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

// FormatTraceHeader formats span data as sentry-trace header
// Format: {trace_id}-{span_id}-{sampled}
// Example: 12345678901234567890123456789012-1234567890123456-1
func FormatTraceHeader(span *sentry.Span) string {
	if span == nil {
		return ""
	}

	traceID := span.TraceID.String()
	spanID := span.SpanID.String()
	sampled := "1" // Always sample if we're injecting

	return fmt.Sprintf("%s-%s-%s", traceID, spanID, sampled)
}

// ContinueFromTraceHeader creates a span option from sentry-trace header
// Header format: {trace_id}-{parent_span_id}-{sampled}
// Example: 12345678901234567890123456789012-1234567890123456-1
// Used by both HTTP and gRPC middleware for distributed tracing
func ContinueFromTraceHeader(header string) sentry.SpanOption {
	return func(span *sentry.Span) {
		parts := strings.Split(header, "-")
		if len(parts) != 3 {
			return
		}

		traceIDStr := parts[0]
		parentSpanIDStr := parts[1]
		// parts[2] is sampled flag

		// Parse trace ID (32 hex chars -> 16 bytes)
		traceIDBytes, err := hex.DecodeString(traceIDStr)
		if err != nil || len(traceIDBytes) != 16 {
			return
		}

		// Parse parent span ID (16 hex chars -> 8 bytes)
		parentSpanIDBytes, err := hex.DecodeString(parentSpanIDStr)
		if err != nil || len(parentSpanIDBytes) != 8 {
			return
		}

		var traceID sentry.TraceID
		copy(traceID[:], traceIDBytes)

		var parentSpanID sentry.SpanID
		copy(parentSpanID[:], parentSpanIDBytes)

		// Set the trace context to continue the parent trace
		span.TraceID = traceID
		span.ParentSpanID = parentSpanID

		// Mark this as continuing from a trace header
		span.SetData("sentry.trace_source", "header")
	}
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
