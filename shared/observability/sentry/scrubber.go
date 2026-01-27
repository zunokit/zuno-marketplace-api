package sentry

import (
	"regexp"
	"strings"

	"github.com/getsentry/sentry-go"
)

var (
	// CAIP-10 account IDs (chain namespace:address) - MUST be before ETH_ADDRESS pattern
	// Format: eip155:1:0x... or bip122:000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f
	caip10Pattern = regexp.MustCompile(`\beip[0-9]+:[a-zA-Z0-9_-]+:[a-zA-Z0-9_-]+\b`)

	// Ethereum address: 0x followed by 40 hex characters
	ethereumAddrPattern = regexp.MustCompile(`\b0x[a-fA-F0-9]{40}\b`)

	// JWT: header.payload.signature OR header.payload (base64url)
	// Matches tokens with or without signature
	jwtPattern = regexp.MustCompile(`eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+(?:\.[a-zA-Z0-9_-]+)?`)

	// Email: standard email pattern
	emailPattern = regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`)

	// Private keys, seed phrases (64-char hex)
	privateKeyPattern = regexp.MustCompile(`\b[0-9a-fA-F]{64}\b`)
)

// scrubEvent scrubs sensitive data from error events
func scrubEvent(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	// Scrub request headers
	if event.Request != nil {
		event.Request.Headers = scrubHeaders(event.Request.Headers)
		// Scrub request body data (string type)
		event.Request.Data = scrubString(event.Request.Data)
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
	event.User = scrubUser(&event.User)

	// Scrub contexts - Context is type alias for map[string]interface{}
	if event.Contexts != nil {
		for key, ctx := range event.Contexts {
			event.Contexts[key] = scrubMap(ctx)
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
// Order matters: CAIP-10 must be before ETH_ADDRESS to prevent partial matches
func scrubString(s string) string {
	// Check CAIP-10 first (e.g., "eip155:1:0x...") before ETH_ADDRESS
	s = caip10Pattern.ReplaceAllString(s, "[FILTERED:CAIP10]")
	s = ethereumAddrPattern.ReplaceAllString(s, "[FILTERED:ETH_ADDRESS]")
	s = jwtPattern.ReplaceAllString(s, "[FILTERED:JWT]")
	s = emailPattern.ReplaceAllString(s, "[FILTERED:EMAIL]")
	s = privateKeyPattern.ReplaceAllString(s, "[FILTERED:PRIVATE_KEY]")
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

	// Sensitive headers that should be completely filtered
	sensitiveKeys := []string{
		"authorization", "cookie", "set-cookie",
		"x-api-key", "x-auth-token", "x-session-id",
		"x-csrf-token", "proxy-authorization", "sec-websocket-key",
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
func scrubUser(user *sentry.User) sentry.User {
	result := *user
	if result.Email != "" {
		result.Email = "[FILTERED]"
	}
	if result.IPAddress != "" {
		result.IPAddress = "[FILTERED]"
	}
	if result.Data != nil {
		// Convert map[string]string to map[string]interface{} for scrubbing
		data := make(map[string]interface{})
		for k, v := range result.Data {
			data[k] = v
		}
		scrubbed := scrubMap(data)
		// Convert back to map[string]string
		result.Data = make(map[string]string)
		for k, v := range scrubbed {
			result.Data[k] = toString(v)
		}
	}
	return result
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
