# Phase 3: Application Configuration

**Priority**: P1
**Status**: Pending
**Effort**: 2 hours

## Context Links

- [Phase 2](./phase-02-environment-config.md) - Environment setup must be complete
- Current Config: `services/auth-service/internal/config/config.go`
- Shared Env: `shared/env/env.go`

## Overview

Update application configuration code to support both Docker (individual env vars) and serverless (connection URLs) modes via `INFRA_MODE` switch.

**Goal**: Services connect to correct infrastructure based on ENV variable with zero runtime changes.

## Key Insights

1. Current config uses individual vars (POSTGRES_HOST, etc.)
2. Need to add URL parsing while maintaining backward compatibility
3. All services use same config pattern - update once, apply everywhere
4. Redis and RabbitMQ also need URL support

## Requirements

### Functional Requirements
- FR1: Add `INFRA_MODE` to all service configs
- FR2: Add `URL` field to DatabaseConfig, RedisConfig, RabbitMQConfig
- FR3: Update `GetDSN()` to use URL when in serverless mode
- FR4: Update all 4 services (auth, user, wallet, gateway)
- FR5: Add connection URL parsing for Redis and RabbitMQ

### Non-Functional Requirements
- NFR1: Backward compatible - existing Docker mode unchanged
- NFR2: Zero behavior change for existing deployments
- NFR3: Code follows existing patterns
- NFR4: No external dependencies added

## Architecture

```
INFRA_MODE=serverless
    │
    ├── Config Load
    │   ├── Read INFRA_MODE
    │   ├── If serverless: read DATABASE_URL, REDIS_URL, CLOUDAMQP_URL
    │   └── If docker: read individual HOST, PORT, USER vars
    │
    └── Connection
        ├── Database: Use URL or build DSN
        ├── Redis: Use URL or build host:port
        └── RabbitMQ: Use URL or build host:port
```

## Related Code Files

### Files to Modify

**Auth Service:**
- `services/auth-service/internal/config/config.go` - Add URL support

**User Service:**
- `services/user-service/internal/config/config.go` - Add URL support

**Wallet Service:**
- `services/wallet-service/internal/config/config.go` - Add URL support

**GraphQL Gateway:**
- `services/graphql-gateway/internal/config/config.go` - Add URL support

**Shared:**
- `shared/env/env.go` - May need URL parsing helpers (optional)

### Files to Create
- None (config changes only)

## Implementation Steps

### Step 1: Update Auth Service Config

**File:** `services/auth-service/internal/config/config.go`

```go
// Add to DatabaseConfig struct
type DatabaseConfig struct {
    Mode       string // "docker" or "serverless"
    Host       string
    Port       string
    User       string
    Password   string
    Database   string
    SSLMode    string
    URL        string // Full connection URL for serverless mode
}

// Update Load() function
func Load() *Config {
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

    return &Config{
        Server: ServerConfig{
            GRPCPort: env.GetString("AUTH_GRPC_PORT", ":50051"),
        },
        Database: dbConfig,
        JWT: JWTConfig{
            Secret:            env.GetString("JWT_SECRET", "your-jwt-secret-key-change-in-production"),
            RefreshSecret:     env.GetString("REFRESH_SECRET", "your-refresh-secret-key-change-in-production"),
            AccessExpiration:  time.Duration(env.GetInt("JWT_ACCESS_EXPIRATION_HOURS", 1)) * time.Hour,
            RefreshExpiration: time.Duration(env.GetInt("JWT_REFRESH_EXPIRATION_DAYS", 7)) * 24 * time.Hour,
        },
        Services: ServicesConfig{
            UserServiceURL:   env.GetString("USER_SERVICE_URL", "localhost:50052"),
            WalletServiceURL: env.GetString("WALLET_SERVICE_URL", "localhost:50053"),
        },
    }
}

// Update GetDSN() to support URL
func (c *DatabaseConfig) GetDSN() string {
    if c.Mode == "serverless" && c.URL != "" {
        return c.URL
    }
    // Fall back to existing behavior (docker mode)
    return "host=" + c.Host + " port=" + c.Port + " user=" + c.User +
        " password=" + c.Password + " dbname=" + c.Database + " sslmode=" + c.SSLMode
}
```

### Step 2: Add Redis and RabbitMQ Config (Auth Service Only)

```go
// Add to Config struct
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
    Services ServicesConfig
    Redis    RedisConfig    // NEW
    RabbitMQ RabbitMQConfig // NEW
}

// Add new config structs
type RedisConfig struct {
    Mode string // "docker" or "serverless"
    Host string
    Port string
    URL  string // Full URL for serverless
}

type RabbitMQConfig struct {
    Mode     string // "docker" or "serverless"
    Host     string
    Port     string
    User     string
    Password string
    Exchange string
    URL      string // Full URL for serverless
}

// Update Load() function
func Load() *Config {
    mode := env.GetString("INFRA_MODE", "docker")

    // ... existing database config ...

    // Redis config
    redisConfig := RedisConfig{Mode: mode}
    if mode == "serverless" {
        redisConfig.URL = env.GetString("REDIS_URL", "")
    } else {
        redisConfig.Host = env.GetString("REDIS_HOST", "localhost")
        redisConfig.Port = env.GetString("REDIS_PORT", "6379")
    }

    // RabbitMQ config
    rabbitConfig := RabbitMQConfig{Mode: mode}
    if mode == "serverless" {
        rabbitConfig.URL = env.GetString("CLOUDAMQP_URL", "")
    } else {
        rabbitConfig.Host = env.GetString("RABBITMQ_HOST", "localhost")
        rabbitConfig.Port = env.GetString("RABBITMQ_PORT", "5672")
        rabbitConfig.User = env.GetString("RABBITMQ_USER", "guest")
        rabbitConfig.Password = env.GetString("RABBITMQ_PASSWORD", "guest")
        rabbitConfig.Exchange = env.GetString("RABBITMQ_EXCHANGE", "nft_events")
    }

    return &Config{
        // ... existing fields ...
        Redis:    redisConfig,
        RabbitMQ: rabbitConfig,
    }
}

// Add helper methods
func (c *RedisConfig) GetAddr() string {
    if c.Mode == "serverless" && c.URL != "" {
        return c.URL
    }
    return c.Host + ":" + c.Port
}

func (c *RabbitMQConfig) GetURL() string {
    if c.Mode == "serverless" && c.URL != "" {
        return c.URL
    }
    return fmt.Sprintf("amqp://%s:%s@%s:%s/",
        c.User, c.Password, c.Host, c.Port)
}
```

### Step 3: Apply Same Pattern to Other Services

**User Service** (`services/user-service/internal/config/config.go`):
- Same DatabaseConfig changes (no Redis/RabbitMQ needed)

**Wallet Service** (`services/wallet-service/internal/config/config.go`):
- Same DatabaseConfig changes (no Redis/RabbitMQ needed)

**GraphQL Gateway** (`services/graphql-gateway/internal/config/config.go`):
- Add RedisConfig and RabbitMQConfig
- Same URL pattern as auth service

### Step 4: Update Connection Initialization

Check where connections are initialized (likely in `main.go` or repository files):

```go
// Example: services/auth-service/cmd/main.go

// Old way (still works for docker mode)
db, err := sql.Open("postgres", cfg.Database.GetDSN())

// New way works for both modes
db, err := sql.Open("postgres", cfg.Database.GetDSN())

// Redis connection
import "github.com/redis/go-redis/v9"

var redisClient *redis.Client
if cfg.Redis.Mode == "serverless" {
    // Parse URL and create client
    opt, err := redis.ParseURL(cfg.Redis.GetAddr())
    if err != nil {
        log.Fatal(err)
    }
    redisClient = redis.NewClient(opt)
} else {
    redisClient = redis.NewClient(&redis.Options{
        Addr: cfg.Redis.GetAddr(),
    })
}

// RabbitMQ connection
import amqp "github.com/rabbitmq/amqp091-go"

var rabbitConn *amqp.Connection
var err error
rabbitConn, err = amqp.Dial(cfg.RabbitMQ.GetURL())
if err != nil {
    log.Fatal(err)
}
```

## Todo List

- [ ] Update auth-service config with URL support
- [ ] Update auth-service main.go to use new config
- [ ] Update user-service config with URL support
- [ ] Update wallet-service config with URL support
- [ ] Update graphql-gateway config with URL support
- [ ] Add connection tests for serverless mode
- [ ] Test docker mode still works
- [ ] Run existing tests to ensure no breakage

## Success Criteria

- [ ] Services start with `INFRA_MODE=docker` (existing behavior)
- [ ] Services start with `INFRA_MODE=serverless` (new behavior)
- [ ] All existing tests pass
- [ ] No breaking changes to existing deployments

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| URL parsing fails | Low | Medium | Test with real connection strings |
| Existing deployments break | Low | High | Default to docker mode |
| Redis client incompatibility | Low | Low | Use go-redis v9+ (supports URL) |
| RabbitMQ URL format differs | Low | Low | Test with CloudAMQP URL |

## Security Considerations

- **URL parsing**: Validate URLs before using
- **Error handling**: Fail fast on invalid URLs
- **Secrets in logs**: Ensure connection URLs not logged
- **Env var precedence**: URL mode should not read individual vars

## Next Steps

- Proceed to [Phase 4: Documentation & Scripts](./phase-04-documentation-scripts.md)
- Update README and create health check scripts

## Unresolved Questions

- Should we add connection health checks on startup?
- How to handle partial configuration (some docker, some serverless)?
- Should we log which mode is active on startup?

---

**Note**: This phase modifies multiple service config files. Consider implementing incrementally:
1. Auth service first
2. Test thoroughly
3. Apply to other services
4. Final integration test
