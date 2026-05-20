package tracing

import (
	"strings"
	"testing"

	"github.com/getsentry/sentry-go"
)

// applySpanOption builds a span and runs a SpanOption against it so we can
// observe the mutations the option performs. We use sentry.StartSpan so the
// span is fully initialized before the option runs.
func applySpanOption(t *testing.T, opt sentry.SpanOption) *sentry.Span {
	t.Helper()
	// Use a fresh hub to keep tests isolated.
	hub := sentry.NewHub(nil, sentry.NewScope())
	ctx := sentry.SetHubOnContext(t.Context(), hub)
	span := sentry.StartSpan(ctx, "test")
	defer span.Finish()
	opt(span)
	return span
}

func TestContinueFromTraceHeader_ValidHeaderSetsTraceAndParentSpan(t *testing.T) {
	header := "0123456789abcdef0123456789abcdef-fedcba9876543210-1"
	opt := ContinueFromTraceHeader(header)

	span := applySpanOption(t, opt)

	if got := span.TraceID.String(); got != "0123456789abcdef0123456789abcdef" {
		t.Errorf("trace id = %q, want %q", got, "0123456789abcdef0123456789abcdef")
	}
	if got := span.ParentSpanID.String(); got != "fedcba9876543210" {
		t.Errorf("parent span id = %q, want %q", got, "fedcba9876543210")
	}
	if got := span.Data["sentry.trace_source"]; got != "header" {
		t.Errorf("sentry.trace_source data = %v, want %q", got, "header")
	}
}

func TestContinueFromTraceHeader_RejectsMalformedHeaders(t *testing.T) {
	cases := []struct {
		name   string
		header string
	}{
		{"empty", ""},
		{"two parts", "abc-def"},
		{"four parts", "a-b-c-d"},
		{"trace id non hex", "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz-fedcba9876543210-1"},
		{"trace id wrong length", "deadbeef-fedcba9876543210-1"},
		{"span id non hex", "0123456789abcdef0123456789abcdef-zzzzzzzzzzzzzzzz-1"},
		{"span id wrong length", "0123456789abcdef0123456789abcdef-deadbeef-1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opt := ContinueFromTraceHeader(tc.header)
			span := applySpanOption(t, opt)

			// Span should still have its auto-generated trace id; we only assert
			// the data marker wasn't set, since both IDs are random per span.
			if _, ok := span.Data["sentry.trace_source"]; ok {
				t.Errorf("expected sentry.trace_source to be unset for malformed header %q", tc.header)
			}

			if strings.HasPrefix(span.TraceID.String(), "0123") {
				// Random trace IDs are extremely unlikely to start with this
				// prefix; if they do, the option mutated the span which is the
				// bug we are guarding against.
				t.Errorf("trace id should not have been overwritten for malformed header %q", tc.header)
			}
		})
	}
}
