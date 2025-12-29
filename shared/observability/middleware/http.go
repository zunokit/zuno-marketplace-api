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
