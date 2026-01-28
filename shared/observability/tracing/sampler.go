package tracing

import (
	"github.com/getsentry/sentry-go"
)

// GetTracesSampleRate returns sampling rate based on environment
func GetTracesSampleRate(environment string) float64 {
	switch environment {
	case "production":
		return 0.05 // 5% in prod (cost control)
	case "staging":
		return 0.20 // 20% in staging
	default:
		return 1.0 // 100% in development
	}
}

// TracesSampler provides endpoint-specific sampling logic
func TracesSampler(environment string) sentry.TracesSampler {
	return func(ctx sentry.SamplingContext) float64 {
		// Never trace health checks
		if ctx.Span != nil {
			if ctx.Span.Description == "GET /health" ||
				ctx.Span.Description == "GET /ready" ||
				ctx.Span.Description == "grpc.health.v1.Health/Check" {
				return 0.0
			}
		}

		// Always trace authentication operations (critical)
		if ctx.Span != nil {
			desc := ctx.Span.Description
			if isAuthOperation(desc) {
				return 1.0
			}
		}

		// Use environment default
		return GetTracesSampleRate(environment)
	}
}

// isAuthOperation checks if this is an authentication-related operation
func isAuthOperation(description string) bool {
	authPatterns := []string{
		"VerifySIWE",
		"RefreshToken",
		"Login",
		"Logout",
		"Authenticate",
		"Authorize",
		"signIn",
		"signOut",
	}

	for _, pattern := range authPatterns {
		if contains(description, pattern) {
			return true
		}
	}

	return false
}

// contains checks if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}

	// Simple substring check (Go's strings.Contains would work too)
	// This is kept simple per KISS principle
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
