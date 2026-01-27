package sentry

import (
	"time"

	"github.com/getsentry/sentry-go"
)

// Init initializes Sentry with production-ready defaults
func Init(dsn, environment, service, release string, tracesSampleRate float64) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		Release:          release,
		TracesSampleRate: tracesSampleRate,
		EnableTracing:    true,

		// Security: Auto-scrub sensitive data
		SendDefaultPII: false,
		MaxBreadcrumbs: 100,

		// Privacy scrubbing hooks
		BeforeSend:            scrubEvent,
		BeforeSendTransaction: scrubTransaction,

		// Attach stacktraces
		AttachStacktrace: true,

		// Service identification
		ServerName: service,
	})
}

// Flush ensures all events are sent before shutdown
// Call before program exits, typically with 2 second timeout
func Flush(timeout time.Duration) bool {
	return sentry.Flush(timeout)
}

// CaptureException captures an error and sends to Sentry
func CaptureException(err error) *sentry.EventID {
	return sentry.CaptureException(err)
}

// CaptureMessage captures a message and sends to Sentry
func CaptureMessage(message string) *sentry.EventID {
	return sentry.CaptureMessage(message)
}

// AddBreadcrumb adds a breadcrumb for context
func AddBreadcrumb(message string, level sentry.Level, data map[string]interface{}) {
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Message:   message,
		Level:     level,
		Data:      scrubMap(data),
		Timestamp: time.Now(),
	})
}
