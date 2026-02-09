# Phase 02: Migrate Media-Service to Local Config

## Changes to `internal/config/config.go`

### Before
```go
import sharedConfig "github.com/zunokit/zuno-marketplace-api/shared/config"

type Config struct {
    Server   sharedConfig.ServerConfig
    Database DatabaseConfig
    Storage  StorageConfig
    Sentry   SentryConfig
}
```

### After (Fully Local)
```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Storage  StorageConfig
    Sentry   SentryConfig
}

type ServerConfig struct {
    GRPCPort string
}
```

## Changes to `cmd/main.go`

- Update config usage from `sharedConfig.ServerConfig` to local `ServerConfig`
