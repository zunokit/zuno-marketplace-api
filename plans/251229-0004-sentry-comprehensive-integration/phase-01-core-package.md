# Phase 01: Core Sentry Package

**Context**: `plan.md` | **Priority**: P1 | **Effort**: 1.5h

---

## Overview

Create shared observability package with Sentry initialization, event scrubbing, and core utilities. This is foundation for all other phases.

**Status**: ✅ COMPLETED (2025-12-29) - All tasks implemented, security fixes applied, tests passing (10/10), coverage 71.8%
**Review**: `plans/reports/code-reviewer-251229-0041-sentry-phase01-security-fixes.md`

---

## Related Files

- Brainstorming: `plans/reports/brainstormer-251229-0004-sentry-comprehensive-integration.md`
- Sentry SDK: https://docs.sentry.io/platforms/go/

---

## Requirements

### Functional

- Init Sentry with DSN, environment, service name, release, sampling
- Flush events on shutdown with timeout
- Scrub Ethereum addresses (0x + 40 hex chars)
- Scrub JWT tokens (3-part base64)
- Scrub email addresses
- Attach stacktraces to errors

### Non-Functional

- Zero dependencies beyond sentry-go
- Thread-safe initialization
- Graceful degradation if Sentry unavailable

---

## Architecture

```
shared/observability/sentry/
├── sentry.go       # Init(), Flush(), BeforeSend hooks
└── scrubber.go     # Regex patterns, scrubMap()
```

---

## Implementation Steps

### Step 1: Create Package Directory

```bash
mkdir -p shared/observability/sentry
```

### Step 2: Create sentry.go

**File**: `shared/observability/sentry/sentry.go`

```go
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
		SendDefaultPII:      false,
		MaxRequestBodyBytes: 10_000,
		MaxBreadcrumbs:      100,

		// Privacy scrubbing hooks
		BeforeSend:          scrubEvent,
		BeforeSendTransaction: scrubTransaction,

		// Attach stacktraces
		AttachStacktrace: true,
		StacktraceConfig: sentry.StacktraceConfig{
			ContextLines: 5,
		},
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
	sentry.AddBreadcrumb(sentry.Breadcrumb{
		Message:     message,
		Level:       level,
		Data:        scrubMap(data),
		Timestamp:   time.Now(),
	})
}
```

### Step 3: Create scrubber.go

**File**: `shared/observability/sentry/scrubber.go`

```go
package sentry

import (
	"regexp"
	"strings"

	"github.com/getsentry/sentry-go"
)

var (
	// Ethereum address: 0x followed by 40 hex characters
	ethereumAddrPattern = regexp.MustCompile(`\b0x[a-fA-F0-9]{40}\b`)

	// JWT: header.payload.signature (base64url)
	jwtPattern = regexp.MustCompile(`\beyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]*\b`)

	// Email: standard email pattern
	emailPattern = regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`)

	// Private keys, seed phrases
	privateKeyPattern = regexp.MustCompile(`\b[0-9a-fA-F]{64}\b`)

	// CAIP-10 account IDs (chain namespace:address)
	caip10Pattern = regexp.MustCompile(`\b[eip][0-9]+:[a-zA-Z0-9_-]+:[a-zA-Z0-9_-]+\b`)
)

// scrubEvent scrubs sensitive data from error events
func scrubEvent(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	// Scrub request headers
	if event.Request != nil {
		event.Request.Headers = scrubHeaders(event.Request.Headers)

		// Scrub request body data
		if event.Request.Data != nil {
			event.Request.Data = scrubMap(event.Request.Data)
		}
	}

	// Scrub breadcrumbs
	if event.Breadcrumbs != nil {
		for i := range event.Breadcrumbs {
			if event.Breadcrumbs[i].Data != nil {
				event.Breadcrumbs[i].Data = scrubMap(event.Breadcrumbs[i].Data)
			}
		}
	}

	// Scrub extra context
	if event.Extra != nil {
		event.Extra = scrubMap(event.Extra)
	}

	// Scrub user context
	if event.User != nil {
		event.User = scrubUser(event.User)
	}

	// Scrub contexts
	if event.Contexts != nil {
		for key, ctx := range event.Contexts {
			if data, ok := ctx.(map[string]interface{}); ok {
				event.Contexts[key] = scrubMap(data)
			}
		}
	}

	// Scrub exception data
	if event.Exception != nil {
		for i := range event.Exception {
			if event.Exception[i].Stacktrace != nil {
				for j := range event.Exception[i].Stacktrace.Frames {
					if event.Exception[i].Stacktrace.Frames[j].Vars != nil {
						event.Exception[i].Stacktrace.Frames[j].Vars =
							scrubMap(event.Exception[i].Stacktrace.Frames[j].Vars)
					}
				}
			}
		}
	}

	return event
}

// scrubTransaction scrubs sensitive data from transaction events
func scrubTransaction(tx *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	// Reuse event scrubbing logic
	return scrubEvent(tx, hint)
}

// scrubMap recursively scrubs sensitive values from map
func scrubMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		switch v := value.(type) {
		case string:
			result[key] = scrubString(v)
		case map[string]interface{}:
			result[key] = scrubMap(v)
		case []interface{}:
			result[key] = scrubSlice(v)
		default:
			result[key] = value
		}
	}

	return result
}

// scrubString scrubs sensitive patterns from string
func scrubString(s string) string {
	s = ethereumAddrPattern.ReplaceAllString(s, "[FILTERED:ETH_ADDRESS]")
	s = jwtPattern.ReplaceAllString(s, "[FILTERED:JWT]")
	s = emailPattern.ReplaceAllString(s, "[FILTERED:EMAIL]")
	s = privateKeyPattern.ReplaceAllString(s, "[FILTERED:PRIVATE_KEY]")
	s = caip10Pattern.ReplaceAllString(s, "[FILTERED:CAIP10]")
	return s
}

// scrubSlice scrubs sensitive values from slice
func scrubSlice(slice []interface{}) []interface{} {
	result := make([]interface{}, len(slice))

	for i, value := range slice {
		switch v := value.(type) {
		case string:
			result[i] = scrubString(v)
		case map[string]interface{}:
			result[i] = scrubMap(v)
		case []interface{}:
			result[i] = scrubSlice(v)
		default:
			result[i] = value
		}
	}

	return result
}

// scrubHeaders scrubs sensitive headers
func scrubHeaders(headers map[string]string) map[string]string {
	result := make(map[string]string)

	sensitiveKeys := []string{
		"authorization", "cookie", "set-cookie",
		"x-api-key", "x-auth-token", "x-session-id",
	}

	for key, value := range headers {
		lowerKey := strings.ToLower(key)

		// Check if this is a sensitive header
		isSensitive := false
		for _, sensitive := range sensitiveKeys {
			if strings.Contains(lowerKey, sensitive) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			result[key] = "[FILTERED]"
		} else {
			result[key] = scrubString(value)
		}
	}

	return result
}

// scrubUser scrubs sensitive user data
func scrubUser(user *sentry.User) *sentry.User {
	if user.Email != "" {
		user.Email = "[FILTERED]"
	}
	if user.IPAddress != "" {
		user.IPAddress = "[FILTERED]"
	}
	if user.Data != nil {
		user.Data = scrubMap(user.Data)
	}
	return user
}
```

### Step 4: Create README.md

**File**: `shared/observability/README.md`

````markdown
# Observability Package

Shared Sentry integration for all Zuno NFT Marketplace microservices.

## Usage

### Initialization

```go
import (
    obs "github.com/zunokit/zuno-marketplace-api/shared/observability/sentry"
)

func main() {
    // Initialize Sentry
    if err := obs.Init(
        cfg.Sentry.DSN,
        cfg.Sentry.Environment,
        "auth-service",  // service name
        "v1.0.0",        // release version
        0.2,             // 20% trace sampling
    ); err != nil {
        log.Printf("Sentry init failed: %v", err)
    }
    defer obs.Flush(2 * time.Second)
}
```
````

### Capturing Errors

```go
import obs "github.com/zunokit/zuno-marketplace-api/shared/observability/sentry"

// Capture exception
if err != nil {
    obs.CaptureException(err)
}

// Capture message
obs.CaptureMessage("User login failed")

// Add breadcrumb
obs.AddBreadcrumb("User action", sentry.LevelInfo, map[string]interface{}{
    "action": "click_button",
    "button": "submit",
})
```

## Privacy Scrubbing

The following patterns are automatically scrubbed:

- Ethereum addresses (0x + 40 hex chars)
- JWT tokens
- Email addresses
- Private keys
- CAIP-10 account IDs
- Sensitive HTTP headers (Authorization, Cookie, etc.)

````

### Step 5: Update go.mod

**File**: `go.mod`

Add dependency:
```bash
go get github.com/getsentry/sentry-go@latest
````

---

## Todo List

- [x] Create `shared/observability/sentry/` directory
- [x] Implement `sentry.go` with Init, Flush, Capture functions
- [x] Implement `scrubber.go` with regex patterns and scrubbing logic
- [x] Create `README.md` with usage documentation
- [x] Update `go.mod` with sentry-go dependency
- [x] Write unit tests for scrubbing logic
- [x] Verify compilation with `go build ./...`

---

## Success Criteria

- [x] Package compiles without errors
- [x] Init() accepts all parameters and configures Sentry
- [x] Scrubber filters all test patterns (ETH addr, JWT, email, CAIP-10, private keys)
- [x] README documents usage clearly

---

## Risk Assessment

| Risk                  | Mitigation                                  |
| --------------------- | ------------------------------------------- |
| Regex false positives | Test with real wallet addresses             |
| Performance overhead  | Scrubbing is synchronous but minimal        |
| Scrubbing misses data | Multi-layer approach (headers, body, extra) |

---

## Next Steps

→ Phase 02: Middleware Layer
