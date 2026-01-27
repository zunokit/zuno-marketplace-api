package middleware

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	obsTrace "github.com/zunokit/zuno-marketplace-api/shared/observability/tracing"
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
			traceParentOption = obsTrace.ContinueFromTraceHeader(traceHeader)
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

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
