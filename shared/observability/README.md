# Observability Package

Shared Sentry integration for all Zuno NFT Marketplace microservices.

## Usage

### Initialization

```go
import (
    obs "github.com/quangdang46/NFT-Marketplace/shared/observability/sentry"
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

### Capturing Errors

```go
import obs "github.com/quangdang46/NFT-Marketplace/shared/observability/sentry"

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
