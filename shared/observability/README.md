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

---

## Middleware Usage

### HTTP Middleware (Chi)

```go
import obshttp "github.com/quangdang46/NFT-Marketplace/shared/observability/middleware"

router := chi.NewRouter()
router.Use(obshttp.SentryHTTP)  // Add BEFORE other middleware
router.Use(middleware.Logger)
router.Use(middleware.Recoverer)
```

### gRPC Server Interceptor

```go
import obsgrpc "github.com/quangdang46/NFT-Marketplace/shared/observability/middleware"

grpcServer := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        obsgrpc.UnaryServerInterceptor(),
    ),
)
```

### gRPC Client Interceptor

```go
conn, err := grpc.Dial(
    serviceURL,
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithChainUnaryInterceptor(
        obsgrpc.UnaryClientInterceptor(),
    ),
)
```

### GraphQL Middleware

```go
import obsgraphql "github.com/quangdang46/NFT-Marketplace/shared/observability/middleware"

// In GraphQL server setup
srv := handler.NewDefaultServer(schema)
srv.Use(obsgraphql.GraphQLFieldMiddleware())
srv.Use(obsgraphql.GraphQLResponseMiddleware())
```
