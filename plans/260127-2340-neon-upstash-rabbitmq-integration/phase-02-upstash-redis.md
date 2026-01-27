---
title: "Phase 2: Upstash Redis Integration"
description: "Add go-redis client, wrapper package, and implement caching for auth/sessions"
status: pending
priority: P1
effort: 6h
---

# Phase 2: Upstash Redis Integration

## Overview

Add Redis client using go-redis/v9, create shared wrapper package, implement caching for auth sessions, rate limiting, and token blacklisting.

## Context

**Current State:**
- Config has RedisConfig with GetAddr()
- No Redis client initialization
- No caching layer
- Session storage in Postgres only

**Required:**
- Add go-redis/v9 dependency
- Create shared/redis package
- Initialize client in main.go
- Use for session caching, rate limiting

## Requirements

### Functional
- [ ] Add go-redis/v9 to go.mod
- [ ] Create shared/redis wrapper package
- [ ] Initialize Redis client in all service main.go files
- [ ] Implement session caching (auth-service)
- [ ] Implement rate limiting middleware (gateway)
- [ ] Implement token blacklist (logout)
- [ ] Graceful shutdown (close client)

### Non-Functional
- TLS support for Upstash
- Connection pooling
- Retry logic for transient failures
- Circuit breaker pattern

## Architecture

```
┌───────────────────────────────────────────────────────────┐
│  shared/redis/                                             │
│  ├── client.go    (singleton, connection mgmt)            │
│  ├── cache.go     (Get, Set, Delete with TTL)             │
│  ├── rate_limit.go (Sliding window)                        │
│  ├── lock.go      (Distributed lock)                       │
│  └── health.go    (Ping check)                             │
└───────────────────────────────────────────────────────────┘
         │
         ▼
┌───────────────────────────────────────────────────────────┐
│  services/auth-service/cmd/main.go                         │
│  redisClient := redis.NewClient(cfg.Redis.GetAddr())       │
└───────────────────────────────────────────────────────────┘
         │
         ▼
┌───────────────────────────────────────────────────────────┐
│  services/auth-service/internal/repository/                │
│  └── session_repository.go (cache + db)                   │
└───────────────────────────────────────────────────────────┘
```

## Implementation Steps

### 1. Add Dependency

```bash
go get github.com/redis/go-redis/v9
```

### 2. Create shared/redis Package

**File:** `shared/redis/client.go`

```go
package redis

import (
   "context"
   "fmt"
   "time"

    "github.com/redis/go-redis/v9"
)

var (
    client *redis.Client
)

// Init initializes the Redis client
func Init(addr string) error {
    opts, err := redis.ParseURL(addr)
    if err != nil {
        return fmt.Errorf("parse redis url: %w", err)
    }

    // Upstash-specific optimizations
    opts.PoolSize = 10
    opts.MinIdleConns = 5
    opts.MaxRetries = 3
    opts.DialTimeout = 5 * time.Second
    opts.ReadTimeout = 3 * time.Second
    opts.WriteTimeout = 3 * time.Second

    client = redis.NewClient(opts)

    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        return fmt.Errorf("redis ping failed: %w", err)
    }

    return nil
}

// GetClient returns the Redis client
func GetClient() *redis.Client {
    return client
}

// Close closes the Redis client
func Close() error {
    if client != nil {
        return client.Close()
    }
    return nil
}
```

**File:** `shared/redis/cache.go`

```go
package redis

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

// Cache provides simple cache operations
type Cache struct {
    client *redis.Client
}

// NewCache creates a new Cache instance
func NewCache() *Cache {
    return &Cache{client: client}
}

// Get retrieves a value and unmarshals it
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := c.client.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(val), dest)
}

// Set stores a value with JSON marshaling and TTL
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return c.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a key
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
    return c.client.Del(ctx, keys...).Err()
}

// Exists checks if keys exist
func (c *Cache) Exists(ctx context.Context, keys ...string) (int64, error) {
    return c.client.Exists(ctx, keys...).Result()
}
```

**File:** `shared/redis/rate_limit.go`

```go
package redis

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

// RateLimiter implements sliding window rate limiting
type RateLimiter struct {
    client *redis.Client
}

// NewRateLimiter creates a new RateLimiter
func NewRateLimiter() *RateLimiter {
    return &RateLimiter{client: client}
}

// Allow checks if action is allowed within rate limit
// Returns (allowed, remaining, resetTime, error)
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time, error) {
    now := time.Now().Unix()
    windowSec := int64(window.Seconds())

    pipe := rl.client.Pipeline()

    // Remove old entries
    zrem := pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", now-windowSec))

    // Count current requests
    count := pipe.ZCard(ctx, key)

    // Add current request
    pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

    // Set expiry
    pipe.Expire(ctx, key, window)

    _, err := pipe.Exec(ctx)
    if err != nil {
        return false, 0, time.Time{}, err
    }

    current := int(count.Val())
    allowed := current < limit
    remaining := max(0, limit-current-1)
    reset := time.Now().Add(window)

    return allowed, remaining, reset, nil
}
```

**File:** `shared/redis/lock.go`

```go
package redis

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

// Lock provides distributed locking
type Lock struct {
    client *redis.Client
    key    string
    ttl    time.Duration
}

// NewLock creates a new distributed lock
func NewLock(key string, ttl time.Duration) *Lock {
    return &Lock{
        client: client,
        key:    fmt.Sprintf("lock:%s", key),
        ttl:    ttl,
    }
}

// TryAcquire attempts to acquire the lock
func (l *Lock) TryAcquire(ctx context.Context) (bool, error) {
    return l.client.SetNX(ctx, l.key, "1", l.ttl).Result()
}

// Release releases the lock
func (l *Lock) Release(ctx context.Context) error {
    return l.client.Del(ctx, l.key).Err()
}
```

**File:** `shared/redis/health.go`

```go
package redis

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

// Health checks Redis connectivity
func Health(ctx context.Context) error {
    if client == nil {
        return fmt.Errorf("redis client not initialized")
    }

    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    return client.Ping(ctx).Err()
}
```

**File:** `shared/redis/client_test.go`

```go
package redis

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
    // Skip if no Redis
    if client == nil {
        t.Skip("Redis not initialized")
    }

    ctx := context.Background()
    cache := NewCache()

    // Set
    err := cache.Set(ctx, "test:key", map[string]string{"foo": "bar"}, 5*time.Minute)
    require.NoError(t, err)

    // Get
    var result map[string]string
    err = cache.Get(ctx, "test:key", &result)
    require.NoError(t, err)
    assert.Equal(t, "bar", result["foo"])

    // Delete
    err = cache.Delete(ctx, "test:key")
    require.NoError(t, err)

    // Verify gone
    err = cache.Get(ctx, "test:key", &result)
    assert.Error(t, err)
}

func TestRateLimiter(t *testing.T) {
    if client == nil {
        t.Skip("Redis not initialized")
    }

    ctx := context.Background()
    rl := NewRateLimiter()

    // First 5 should be allowed
    for i := 0; i < 5; i++ {
        allowed, _, _, err := rl.Allow(ctx, "test:rl", 5, time.Minute)
        require.NoError(t, err)
        assert.True(t, allowed, "request %d should be allowed", i)
    }

    // 6th should be denied
    allowed, _, _, err := rl.Allow(ctx, "test:rl", 5, time.Minute)
    require.NoError(t, err)
    assert.False(t, allowed, "6th request should be denied")
}
```

### 3. Initialize in Services

**File:** `services/auth-service/cmd/main.go`

```go
import (
    "github.com/zunokit/zuno-marketplace-api/shared/redis"
)

func main() {
    // ... existing code ...

    // Initialize Redis
    if err := redis.Init(cfg.Redis.GetAddr()); err != nil {
        log.Printf("Redis init failed (continuing without cache): %v", err)
    } else {
        log.Println("Redis connected")
        defer redis.Close()
    }

    // ... rest of code ...
}
```

Repeat for:
- `services/user-service/cmd/main.go`
- `services/wallet-service/cmd/main.go`
- `services/graphql-gateway/cmd/main.go`

### 4. Implement Session Caching

**File:** `services/auth-service/internal/repository/session_repository.go`

Add cache layer:

```go
type SessionRepository struct {
    db    *gorm.DB
    cache *redis.Cache
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
    return &SessionRepository{
        db:    db,
        cache: redis.NewCache(),
    }
}

func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
    if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
        return err
    }

    // Cache with TTL matching expiration
    key := fmt.Sprintf("session:%s", session.Token)
    ttl := time.Until(session.ExpiresAt)
    return r.cache.Set(ctx, key, session, ttl)
}

func (r *SessionRepository) FindByToken(ctx context.Context, token string) (*models.Session, error) {
    // Try cache first
    key := fmt.Sprintf("session:%s", token)
    var session models.Session

    if err := r.cache.Get(ctx, key, &session); err == nil {
        return &session, nil
    }

    // Fallback to DB
    err := r.db.WithContext(ctx).Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error
    if err != nil {
        return nil, err
    }

    // Populate cache
    ttl := time.Until(session.ExpiresAt)
    _ = r.cache.Set(ctx, key, &session, ttl)

    return &session, nil
}
```

### 5. Implement Rate Limiting Middleware

**File:** `services/graphql-gateway/internal/middleware/rate_limit.go`

```go
package middleware

import (
    "context"
    "fmt"
    "net/http"

    "github.com/zunokit/zuno-marketplace-api/shared/redis"
)

func RateLimit(rl *redis.RateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Use IP or user ID for rate limit key
            key := fmt.Sprintf("ratelimit:%s", r.RemoteAddr)

            allowed, remaining, reset, err := rl.Allow(r.Context(), key, 100, time.Hour)
            if err != nil {
                // Log but continue on Redis errors
                log.Printf("Rate limit error: %v", err)
                next.ServeHTTP(w, r)
                return
            }

            w.Header().Set("X-RateLimit-Limit", "100")
            w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
            w.Header().Set("X-RateLimit-Reset", reset.Format(time.RFC3339))

            if !allowed {
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

## Related Code Files

**Create:**
- `shared/redis/client.go`
- `shared/redis/cache.go`
- `shared/redis/rate_limit.go`
- `shared/redis/lock.go`
- `shared/redis/health.go`
- `shared/redis/client_test.go`

**Modify:**
- `go.mod` - Add go-redis/v9
- `services/auth-service/cmd/main.go` - Init Redis
- `services/user-service/cmd/main.go` - Init Redis
- `services/wallet-service/cmd/main.go` - Init Redis
- `services/graphql-gateway/cmd/main.go` - Init Redis
- `services/auth-service/internal/repository/session_repository.go` - Add caching
- `services/graphql-gateway/internal/middleware/rate_limit.go` - Create

**Verify:**
- `.env.development.example` - REDIS_URL format
- `.env.production.example` - REDIS_URL format

## Todo Checklist

- [ ] Add go-redis/v9 dependency
- [ ] Create shared/redis/client.go
- [ ] Create shared/redis/cache.go
- [ ] Create shared/redis/rate_limit.go
- [ ] Create shared/redis/lock.go
- [ ] Create shared/redis/health.go
- [ ] Create shared/redis/client_test.go
- [ ] Initialize Redis in auth-service main.go
- [ ] Initialize Redis in user-service main.go
- [ ] Initialize Redis in wallet-service main.go
- [ ] Initialize Redis in graphql-gateway main.go
- [ ] Add caching to session repository
- [ ] Create rate limit middleware
- [ ] Add graceful shutdown for Redis client
- [ ] Test with real Upstash instance

## Success Criteria

- [ ] All services connect to Upstash successfully
- [ ] Session caching reduces DB queries
- [ ] Rate limiting blocks excessive requests
- [ ] Unit tests pass for cache/rate limit
- [ ] Docker mode still works (Redis container)
- [ ] Serverless mode uses Upstash

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Upstash TLS handshake issues | High | Test connection early, verify URL format |
| Redis connection exhaustion | Medium | Configure pool size, monitor |
| Cache inconsistency | Medium | Short TTL, cache invalidation on mutations |
| Rate limit false positives | Low | Log warnings, don't block on Redis errors |

## Security Considerations

- TLS for Upstash (rediss://)
- Connection string in env
- Rate limit per IP/user
- No sensitive data in cache (session tokens OK)

## Next Steps

- **Phase 3:** CloudAMQP RabbitMQ integration
- **Phase 4:** Testing & validation
