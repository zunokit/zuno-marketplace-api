# Code Review Report: Phase 01 - Core Sentry Package

**Date**: 2025-12-29
**Reviewer**: Code Review Agent
**Phase**: Phase 01: Core Sentry Package

---

## Scope

- Files reviewed: 5
  - `shared/observability/sentry/sentry.go` (59 lines)
  - `shared/observability/sentry/scrubber.go` (195 lines)
  - `shared/observability/sentry/scrubber_test.go` (325 lines)
  - `shared/observability/README.md` (58 lines)
  - `go.mod` (dependency added)
- Lines of code analyzed: ~637
- Review focus: Security, performance, architecture, YAGNI/KISS/DRY principles

---

## Overall Assessment

**Grade: C (Critical Issues Must Be Fixed)**

The implementation shows good architectural design with proper separation of concerns. The scrubbing logic is well-organized. However, **3 test failures indicate broken security functionality** - sensitive data patterns are NOT being properly scrubbed.

**Build Status**: PASS
**Test Status**: FAIL (4/9 tests failing)

---

## Critical Issues

### 1. JWT Pattern Without Signature Not Scrubbed (SECURITY)

**File**: `shared/observability/sentry/scrubber.go:15`

**Issue**: JWTs without signature (only header.payload) are NOT being filtered.

**Pattern**: `\beyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]*\b`

**Problem**: The pattern requires 3 parts (header.payload.signature) even though `*` allows empty signature. Word boundary `\b` causes issues when the JWT ends without trailing characters.

**Test Failing**: `TestScrubString_JWT/jwt_without_signature`

**Impact**: JWT tokens in certain formats leak to Sentry.

**Fix**:
```go
// Current (broken)
jwtPattern = regexp.MustCompile(`\beyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]*\b`)

// Suggested fix
jwtPattern = regexp.MustCompile(`eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)?`)
```

### 2. CAIP-10 Pattern Superseded by ETH_ADDRESS (SECURITY)

**File**: `shared/observability/sentry/scrubber.go:24`

**Issue**: CAIP-10 account IDs (e.g., `eip155:1:0xab16...`) are partially scrubbed - only the address portion is filtered, not the full CAIP-10 identifier.

**Test Failing**: `TestScrubString_CAIP10/eip155_account_id`

**Root Cause**: Regex execution order in `scrubString()` - ETH_ADDRESS pattern matches first, replacing just the address portion.

**Current Output**: `Account: eip155:1:[FILTERED:ETH_ADDRESS]`
**Expected Output**: `Account: [FILTERED:CAIP10]`

**Impact**: CAIP-10 account IDs leak chain namespace information.

**Fix Options**:
1. Move CAIP-10 pattern check before ETH_ADDRESS in `scrubString()`
2. Update pattern to be more specific: `\b[eip][0-9]+:[a-zA-Z0-9_-]+:\b0x[a-fA-F0-9]{40}\b`

### 3. Nested JWT Not Scrubbed in Map (SECURITY)

**File**: `shared/observability/sentry/scrubber_test.go:150`

**Test Failing**: `TestScrubMap_Nested`

**Issue**: JWT token without signature in nested map not being scrubbed.

**Impact**: Nested sensitive data in complex structures leaks to Sentry.

**Root Cause**: Related to Issue #1 - JWT pattern doesn't match properly.

### 4. Multiple Patterns in Same String (SECURITY)

**File**: `shared/observability/sentry/scrubber_test.go:298`

**Test Failing**: `TestMultiplePatterns`

**Issue**: When multiple sensitive patterns exist in one string, JWT pattern fails to match.

**Impact**: Mixed sensitive data strings leak JWT tokens.

---

## High Priority Findings

### 1. No Error Handling in Init() (RELIABILITY)

**File**: `shared/observability/sentry/sentry.go:10`

```go
func Init(dsn, environment, service, release string, tracesSampleRate float64) error {
    return sentry.Init(sentry.ClientOptions{...})
}
```

**Issue**: If Sentry init fails (e.g., invalid DSN, network issues), error is returned but caller may ignore it. Sentry will be unavailable but app continues.

**Impact**: Errors silently not captured; debugging harder in production.

**Suggestion**: Add graceful degradation or logger:
```go
func Init(dsn, environment, service, release string, tracesSampleRate float64) error {
    err := sentry.Init(...)
    if err != nil {
        // Log but don't fail - Sentry is optional
        log.Printf("WARNING: Sentry init failed: %v", err)
    }
    return err
}
```

### 2. Memory Allocation on Every Scrub (PERFORMANCE)

**File**: `shared/observability/sentry/scrubber.go:84-101`

**Issue**: `scrubMap()` and `scrubSlice()` create new map/slice on every call. For large payloads (e.g., request bodies with 1000s of entries), this causes significant GC pressure.

**Impact**: High memory allocation rate for observability path.

**Suggestion**: Consider in-place modification for large maps (trade-off: mutability).

### 3. Private Key Pattern Over-Matching (SECURITY)

**File**: `shared/observability/sentry/scrubber.go:21`

```go
privateKeyPattern = regexp.MustCompile(`\b[0-9a-fA-F]{64}\b`)
```

**Issue**: Matches ANY 64-char hex string, including:
- Transaction hashes
- Block hashes
- Random hex identifiers

**Impact**: Legitimate data is falsely marked as private keys.

**Suggestion**: Be more specific with context (e.g., look for "privatekey", "secret" keywords nearby) or require word boundary markers like "0x" prefix check negative.

### 4. Missing Sensitive Headers (SECURITY)

**File**: `shared/observability/sentry/scrubber.go:137-140`

```go
sensitiveKeys := []string{
    "authorization", "cookie", "set-cookie",
    "x-api-key", "x-auth-token", "x-session-id",
}
```

**Missing headers**:
- `x-csrf-token` / `x-csrf-header`
- `proxy-authorization`
- `sec-websocket-key` (for WS upgrades)
- `sec-ch-ua` (possibly privacy-sensitive client hints)

---

## Medium Priority Improvements

### 1. Inefficient Header Scrubbing (PERFORMANCE)

**File**: `shared/observability/sentry/scrubber.go:142-152`

**Issue**: Nested loop with `strings.Contains` on every header. O(n*m) complexity where n=headers, m=sensitive keys.

**Suggestion**: Use map for O(1) lookup:
```go
sensitiveKeysMap := map[string]bool{
    "authorization": true, "cookie": true, ...
}
if sensitiveKeysMap[lowerKey] || strings.HasPrefix(lowerKey, "x-") {
    // ...
}
```

### 2. README Module Path Inconsistent (DOCUMENTATION)

**File**: `shared/observability/README.md:11`

```go
import obs "github.com/quangdang46/NFT-Marketplace/shared/observability/sentry"
```

**Issue**: Hardcoded username `quangdang46` in module path.

**Suggestion**: Use placeholder or relative import in docs.

### 3. No Integration Tests (TESTING)

**Missing**: Integration tests for:
- End-to-end event capture and scrubbing
- Performance benchmarks for scrubbing
- Sentry DSN validation

### 4. Test Helper Functions Not Exported (CODE QUALITY)

**File**: `shared/observability/sentry/scrubber_test.go:313-324`

Functions `contains()` and `containsMiddle()` are test-only but could use `strings.Contains()` from stdlib.

---

## Low Priority Suggestions

### 1. Go Doc Comments

**File**: `shared/observability/sentry/sentry.go`

Add package documentation:
```go
// Package sentry provides Sentry error tracking integration with automatic
// sensitive data scrubbing for Ethereum addresses, JWT tokens, emails, and
// private keys.
package sentry
```

### 2. Constants for Filter Labels

**File**: `shared/observability/sentry/scrubber.go:105-110`

Magic strings like `[FILTERED:ETH_ADDRESS]` should be constants:
```go
const (
    filterEthAddr    = "[FILTERED:ETH_ADDRESS]"
    filterJWT        = "[FILTERED:JWT]"
    filterEmail      = "[FILTERED:EMAIL]"
    // ...
)
```

### 3. sentry-go Version (DEPENDENCY)

**File**: `go.mod:22`

Current: `github.com/getsentry/sentry-go v0.40.0`

Check if this is latest. As of review, v0.41.0+ may be available.

---

## Positive Observations

1. **Clean Architecture**: Proper separation between core sentry functions and scrubbing logic
2. **Comprehensive Test Coverage**: 9 test functions covering edge cases
3. **Recursive Scrubbing**: Handles nested maps and slices correctly
4. **Security-First**: SendDefaultPII set to false, proper hook registration
5. **Good Documentation**: README with usage examples
6. **Type Safety**: Proper use of sentry.User, sentry.Event types
7. **Table-Driven Tests**: Follows Go best practices

---

## Architectural Assessment

| Principle | Status | Notes |
|-----------|--------|-------|
| YAGNI | PASS | No unnecessary features detected |
| KISS | PASS | Code is simple and readable |
| DRY | PASS | Minimal repetition, good helper functions |
| SOLID | PASS | Single responsibility per function |
| Security | FAIL | 3 test failures = broken scrubbing |

---

## Recommended Actions

### Must Fix (Before Merge)

1. **Fix JWT regex pattern** to match tokens without signature
2. **Fix CAIP-10 pattern** to match before ETH_ADDRESS
3. **Verify all tests pass** - currently 4/9 failing
4. **Add missing sensitive headers** (csrf, proxy-auth)

### Should Fix (Before Production)

5. Add graceful degradation in Init() if Sentry unavailable
6. Fix private key pattern over-matching
7. Optimize header scrubbing with map lookup

### Nice to Have

8. Add integration tests
9. Extract filter label constants
10. Update go.mod to latest sentry-go
11. Add package godoc comment

---

## Metrics

| Metric | Value | Target |
|--------|-------|--------|
| Test Pass Rate | 56% (5/9) | 100% |
| Build Status | PASS | PASS |
| go vet | PASS | PASS |
| Code Coverage | ~70% | 80% |
| Security Issues | 4 critical | 0 |

---

## Unresolved Questions

1. Should CAIP-10 be a separate filter or is ETH_ADDRESS scrubbing sufficient?
2. Is graceful degradation acceptable if Sentry is unavailable?
3. Should we add benchmark tests for scrubbing performance?
4. What is the acceptable memory overhead for scrubbing large payloads?
5. Should we add opt-out config for specific scrubbing patterns?

---

## Summary

The Phase 01 implementation has solid architectural foundations but **fails critical security tests**. The scrubbing logic for JWT, CAIP-10, and nested patterns is broken. Must fix regex patterns and ensure all tests pass before merging to main.

**Recommendation**: **DO NOT MERGE** until critical issues are resolved.

**Estimated Fix Time**: 1-2 hours for regex fixes + test verification.
