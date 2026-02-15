# Phase 1: Setup Redis Configuration

## Overview
Add Redis configuration and initialization to the collection service. Supports both **Upstash (serverless)** for development and **Docker** for production.

**Status:** Pending
**Blocked By:** None
**Blocks:** Phase 2, 3, 4, 5

---

## Context

The collection service currently uses `shared/config` which has no Redis. The auth-service has its own config with `RedisConfig` supporting:
- **Serverless mode (Upstash)**: Uses `REDIS_URL` from environment
- **Docker mode**: Uses `REDIS_HOST` and `REDIS_PORT`

We need to add the same pattern to collection-service.

---

## Files to Modify

### 1. `services/collection-service/internal/config/config.go`

Add `RedisConfig` struct and update `Load()` function:

```go
package config

import (
	"strings"

	sharedConfig "github.com/zunokit/zuno-marketplace-api/shared/config"
	"github.com/zunokit/zuno-marketplace-api/shared/env"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Mode     string // "docker" or "serverless"
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
	URL      string // Full connection URL for serverless mode
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Mode string // "docker" or "serverless"
	Host string
	Port string
	URL  string // Full URL for serverless (Upstash)
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	if c.Mode == "serverless" && c.URL != "" {
		if !strings.Contains(c.URL, "sslmode=") {
			return c.URL + "&sslmode=require"
		}
		return c.URL
	}
	return "host=" + c.Host + " port=" + c.Port + " user=" + c.User +
		" password=" + c.Password + " dbname=" + c.Database + " sslmode=" + c.SSLMode
}

// GetAddr returns the Redis address
func (c *RedisConfig) GetAddr() string {
	if c.Mode == "serverless" && c.URL != "" {
		return c.URL
	}
	return c.Host + ":" + c.Port
}

// Load loads configuration from environment variables
func Load() *sharedConfig.Config {
	mode := env.GetString("INFRA_MODE", "docker")

	// Database config
	dbConfig := DatabaseConfig{Mode: mode}
	if mode == "serverless" {
		dbConfig.URL = env.GetString("DATABASE_URL", "")
	} else {
		dbConfig.Host = env.GetString("POSTGRES_HOST", "localhost")
		dbConfig.Port = env.GetString("POSTGRES_PORT", "5432")
		dbConfig.User = env.GetString("POSTGRES_USER", "postgres")
		dbConfig.Password = env.GetString("POSTGRES_PASSWORD", "postgres")
		dbConfig.Database = env.GetString("POSTGRES_DATABASE", "nft_marketplace")
		dbConfig.SSLMode = env.GetString("POSTGRES_SSL_MODE", "disable")
	}

	// Redis config (NEW)
	redisConfig := RedisConfig{Mode: mode}
	if mode == "serverless" {
		redisConfig.URL = env.GetString("REDIS_URL", "")
	} else {
		redisConfig.Host = env.GetString("REDIS_HOST", "localhost")
		redisConfig.Port = env.GetString("REDIS_PORT", "6379")
	}

	return &sharedConfig.Config{
		Server: sharedConfig.ServerConfig{
			GRPCPort: env.GetString("COLLECTION_GRPC_PORT", ":50054"),
		},
		Database: sharedConfig.DatabaseConfig{
			Host:     dbConfig.Host,
			Port:     dbConfig.Port,
			User:     dbConfig.User,
			Password: dbConfig.Password,
			Database: dbConfig.Database,
			SSLMode:  dbConfig.SSLMode,
		},
		DatabaseDSN: dbConfig.GetDSN(),
		// Note: Redis config not in sharedConfig, handle separately in main.go
	}
}

// GetRedisConfig returns Redis configuration (call this in main.go)
func GetRedisConfig() RedisConfig {
	mode := env.GetString("INFRA_MODE", "docker")
	redisConfig := RedisConfig{Mode: mode}
	if mode == "serverless" {
		redisConfig.URL = env.GetString("REDIS_URL", "")
	} else {
		redisConfig.Host = env.GetString("REDIS_HOST", "localhost")
		redisConfig.Port = env.GetString("REDIS_PORT", "6379")
	}
	return redisConfig
}
```

### 2. `services/collection-service/cmd/main.go`

Initialize Redis client (non-blocking) and pass to repository:

```go
package main

import (
	// ... existing imports ...
	sharedredis "github.com/zunokit/zuno-marketplace-api/shared/redis"
)

func main() {
	// ... existing setup (logger, config, db) ...

	// Load full config with Redis
	cfg := config.Load()
	redisConfig := config.GetRedisConfig()

	// Initialize Redis (non-blocking - continues without cache if Redis fails)
	if err := sharedredis.Init(redisConfig.GetAddr()); err != nil {
		log.Infof("Redis init failed (continuing without cache): %v", err)
	} else {
		log.Info("Redis connected")
		defer sharedredis.Close()
	}

	// Create cache instance
	cache := sharedredis.NewCache()

	// Create repository with caching
	baseRepo := repository.NewCollectionRepository(db)
	cachedRepo := repository.NewCachedCollectionRepository(baseRepo, cache)

	// Pass cachedRepo to service
	collectionService := service.NewCollectionService(
		cachedRepo,
		allowlistRepo,
		metadataRepo,
	)

	// ... rest of initialization ...
}
```

---

## Environment Variables

### Serverless (Development with Upstash)
```bash
INFRA_MODE=serverless
DATABASE_URL=postgresql://...
REDIS_URL=rediss://default:...@...-db.upstash.io:6379
```

### Docker (Production)
```bash
INFRA_MODE=docker
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
REDIS_HOST=redis
REDIS_PORT=6379
```

---

## Implementation Steps

1. **Add `RedisConfig` struct** to `config.go`
   - Add `Mode`, `Host`, `Port`, `URL` fields
   - Add `GetAddr()` method (returns URL for serverless, host:port for docker)

2. **Add `GetRedisConfig()` function** to return Redis config
   - Read `INFRA_MODE` environment variable
   - Load `REDIS_URL` for serverless or `REDIS_HOST/PORT` for docker

3. **Initialize Redis in main.go**
   - Call `sharedredis.Init()` with address from config (MUST happen before `NewCache()`)
   - Non-blocking: log and continue if Redis fails
   - Create `sharedredis.NewCache()` instance (uses the initialized global client)
   - Wrap repository with `NewCachedCollectionRepository()`
   - **Important:** Order matters - `Init()` must be called before `NewCache()`

4. **Update `.env.example`** (if exists)
   - Add `REDIS_URL` for serverless mode
   - Add `REDIS_HOST` and `REDIS_PORT` for docker mode

---

## Testing Checklist

- [ ] Service starts **without Redis** (fallback mode works)
- [ ] Service connects to **Upstash Redis** when `INFRA_MODE=serverless`
- [ ] Service connects to **Docker Redis** when `INFRA_MODE=docker`
- [ ] Config loads Redis URL from environment correctly
- [ ] `GetAddr()` returns correct format for each mode

---

## References

- Auth service pattern: `services/auth-service/internal/config/config.go`
- Auth service main: `services/auth-service/cmd/main.go`
- Shared Redis client: `shared/redis/client.go`
- Redis already has Upstash optimizations built-in

---

## Next Phase

[Phase 2: Cache Key Utilities](./phase-02-cache-keys.md)
