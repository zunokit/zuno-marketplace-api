# Phase 01: Migrate Collection-Service to Local Config

## Changes to `internal/config/config.go`

### Before (Hybrid)
```go
import sharedConfig "github.com/zunokit/zuno-marketplace-api/shared/config"

type Config struct {
    Server      sharedConfig.ServerConfig
    Database    sharedConfig.DatabaseConfig
    DatabaseDSN string
    Sentry      SentryConfig
}
```

### After (Fully Local)
```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    Sentry   SentryConfig
}

type ServerConfig struct {
    GRPCPort string
}

type DatabaseConfig struct {
    Mode     string
    Host     string
    Port     string
    User     string
    Password string
    Database string
    SSLMode  string
    URL      string
}

func (c *DatabaseConfig) GetDSN() string { ... }

type RedisConfig struct {
    Mode string
    Host string
    Port string
    URL  string
}

func (c *RedisConfig) GetAddr() string { ... }
```

## Changes to `cmd/main.go`

- Update `collectionRepo := repository.NewCachedCollectionRepository(baseRepo, cache)` - cache comes from config
- Access Redis via `cfg.Redis.GetAddr()` instead of `GetRedisConfig()`
