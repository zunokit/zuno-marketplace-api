package middleware

import (
	"context"
	"fmt"

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
		if cc != nil {
			span.SetData("grpc.target", cc.Target())
		}

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
	return sentry.WithTransactionSource("custom")
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
