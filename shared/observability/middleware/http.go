package middleware

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

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

		// Extract sentry-trace header before starting transaction for distributed tracing
		var traceParentOption sentry.SpanOption
		if traceHeader := r.Header.Get("sentry-trace"); traceHeader != "" {
			traceParentOption = continueFromTraceHeaderHTTP(traceHeader)
		}

		// Start transaction with trace continuation option if available
		var transaction *sentry.Span
		if traceHeader := r.Header.Get("sentry-trace"); traceHeader != "" {
			// Create transaction with parent trace context
			transaction = sentry.StartTransaction(
				r.Context(),
				fmt.Sprintf("%s %s", r.Method, r.URL.Path),
				sentry.WithTransactionName(r.URL.Path),
				sentry.WithOpName("http.server"),
				sentry.WithDescription(fmt.Sprintf("%s request to %s", r.Method, r.URL.Path)),
			)
			// Apply trace continuation
			traceParentOption(transaction)
		} else {
			// No parent trace, start new transaction
			transaction = sentry.StartTransaction(
				r.Context(),
				fmt.Sprintf("%s %s", r.Method, r.URL.Path),
				sentry.WithTransactionName(r.URL.Path),
				sentry.WithOpName("http.server"),
				sentry.WithDescription(fmt.Sprintf("%s request to %s", r.Method, r.URL.Path)),
			)
		}
		defer transaction.Finish()

		// Add HTTP context data
		transaction.SetData("http.method", r.Method)
		transaction.SetData("http.url", r.URL.String())
		transaction.SetData("http.scheme", r.URL.Scheme)
		transaction.SetData("http.host", r.Host)
		transaction.SetData("http.path", r.URL.Path)
		transaction.SetData("http.query", r.URL.RawQuery)
		transaction.SetData("http.remote_addr", r.RemoteAddr)

		// Continue with trace context
		r = r.WithContext(transaction.Context())

		// Wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)

		// Set HTTP status on transaction
		transaction.SetData("http.status_code", rw.status)

		// Map HTTP status to span status
		if rw.status >= 400 && rw.status < 500 {
			transaction.Status = sentry.SpanStatusInvalidArgument
		} else if rw.status >= 500 {
			transaction.Status = sentry.SpanStatusInternalError
		} else {
			transaction.Status = sentry.SpanStatusOK
		}
	})
}

// continueFromTraceHeaderHTTP creates a span option from sentry-trace header for HTTP
// Header format: {trace_id}-{parent_span_id}-{sampled}
func continueFromTraceHeaderHTTP(header string) sentry.SpanOption {
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

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
