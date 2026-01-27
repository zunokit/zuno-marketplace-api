---
title: "Phase 4: Testing & Validation"
description: "Write tests, validate both docker and serverless modes work"
status: pending
priority: P1
effort: 2h
---

# Phase 4: Testing & Validation

## Overview

Write unit tests for Redis and RabbitMQ wrappers, integration tests with real services, validate both docker and serverless modes.

## Context

**Previous Phases:**
- Phase 1: Neon Database integration
- Phase 2: Upstash Redis integration
- Phase 3: CloudAMQP RabbitMQ integration

**Current State:**
- All code implemented
- Need comprehensive tests
- Need validation of both modes

## Requirements

### Functional
- [ ] Unit tests for Redis wrapper
- [ ] Unit tests for RabbitMQ wrapper
- [ ] Integration tests with real Upstash
- [ ] Integration tests with real CloudAMQP
- [ ] Docker mode validation
- [ ] Serverless mode validation

### Non-Functional
- Test isolation
- Mock external dependencies where appropriate
- Fast feedback loop

## Implementation Steps

### 1. Unit Tests (Already Created)

From previous phases:
- `shared/redis/client_test.go`
- `shared/rabbitmq/publisher_test.go`

### 2. Integration Tests

**File:** `tests/integration/redis_integration_test.go`

```go
// +build integration

package integration_test

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/zunokit/zuno-marketplace-api/shared/redis"
)

func TestRedisIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    redisURL := os.Getenv("REDIS_URL")
    if redisURL == "" {
        t.Skip("REDIS_URL not set")
    }

    // Initialize
    err := redis.Init(redisURL)
    require.NoError(t, err)
    defer redis.Close()

    ctx := context.Background()
    cache := redis.NewCache()

    // Test cache
    t.Run("Cache", func(t *testing.T) {
        key := "test:integration"
        value := map[string]string{"foo": "bar"}

        // Set
        err := cache.Set(ctx, key, value, 5*time.Minute)
        require.NoError(t, err)

        // Get
        var result map[string]string
        err = cache.Get(ctx, key, &result)
        require.NoError(t, err)
        assert.Equal(t, "bar", result["foo"])

        // Delete
        err = cache.Delete(ctx, key)
        require.NoError(t, err)

        // Verify gone
        err = cache.Get(ctx, key, &result)
        assert.Error(t, err)
    })

    // Test rate limiting
    t.Run("RateLimit", func(t *testing.T) {
        rl := redis.NewRateLimiter()

        // First 5 allowed
        for i := 0; i < 5; i++ {
            allowed, _, _, err := rl.Allow(ctx, "test:rl", 5, time.Minute)
            require.NoError(t, err)
            assert.True(t, allowed)
        }

        // 6th denied
        allowed, _, _, err := rl.Allow(ctx, "test:rl", 5, time.Minute)
        require.NoError(t, err)
        assert.False(t, allowed)
    })
}
```

**File:** `tests/integration/rabbitmq_integration_test.go`

```go
// +build integration

package integration_test

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/zunokit/zuno-marketplace-api/shared/rabbitmq"
)

func TestRabbitMQIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    amqpURL := os.Getenv("CLOUDAMQP_URL")
    if amqpURL == "" {
        t.Skip("CLOUDAMQP_URL not set")
    }

    // Initialize
    err := rabbitmq.Init(rabbitmq.Config{URL: amqpURL})
    require.NoError(t, err)
    defer rabbitmq.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Test publish
    t.Run("Publish", func(t *testing.T) {
        pub := rabbitmq.NewPublisher()

        event := rabbitmq.NewEvent(
            rabbitmq.EventAuthLogin,
            "test-service",
            map[string]string{"userId": "test123"},
        )

        err := pub.Publish(ctx, "auth.login", event)
        require.NoError(t, err)
    })

    // Test consume
    t.Run("Consume", func(t *testing.T) {
        received := make(chan rabbitmq.Event, 1)

        consumer := rabbitmq.NewConsumer("auth_events", func(ctx context.Context, event rabbitmq.Event) error {
            received <- event
            return nil
        })

        go consumer.Consume(ctx)

        // Publish test event
        pub := rabbitmq.NewPublisher()
        event := rabbitmq.NewEvent(
            rabbitmq.EventAuthLogin,
            "test-service",
            map[string]string{"userId": "test456"},
        )
        err := pub.Publish(ctx, "auth.login", event)
        require.NoError(t, err)

        // Wait for consumption
        select {
        case <-time.After(5 * time.Second):
            t.Fatal("Timeout waiting for event")
        case e := <-received:
            assert.Equal(t, rabbitmq.EventAuthLogin, e.Type)
        }
    })
}
```

### 3. Docker Mode Validation

```bash
# Start docker compose
docker compose up -d

# Run tests
go test ./...

# Verify services start
tilt up
```

### 4. Serverless Mode Validation

```bash
# Set environment
export INFRA_MODE=serverless
export DATABASE_URL="postgresql://..."
export REDIS_URL="redis://..."
export CLOUDAMQP_URL="amqp://..."

# Run tests
go test ./...

# Verify services start
tilt up
```

### 5. Health Check Tests

**File:** `tests/integration/health_test.go`

```go
// +build integration

package integration_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
    "github.com/zunokit/zuno-marketplace-api/shared/rabbitmq"
    "github.com/zunokit/zuno-marketplace-api/shared/redis"
)

func TestHealthChecks(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    t.Run("RedisHealth", func(t *testing.T) {
        err := redis.Health(ctx)
        require.NoError(t, err)
    })

    t.Run("RabbitMQHealth", func(t *testing.T) {
        err := rabbitmq.Health(ctx)
        require.NoError(t, err)
    })
}
```

## Test Commands

```bash
# Unit tests only
go test -short ./...

# Integration tests
go test -tags=integration ./tests/integration/...

# All tests with coverage
go test -cover ./...

# Race detection
go test -race ./...

# Specific package
go test -v ./shared/redis/...
```

## Validation Checklist

### Docker Mode
- [ ] Services start with docker compose
- [ ] PostgreSQL connection works
- [ ] Redis connection works
- [ ] RabbitMQ connection works
- [ ] All tests pass

### Serverless Mode
- [ ] Services start with Tilt
- [ ] Neon connection works (sslmode=require)
- [ ] Upstash connection works (TLS)
- [ ] CloudAMQP connection works
- [ ] All tests pass

### Caching
- [ ] Session cache reduces DB queries
- [ ] Cache misses hit DB correctly
- [ ] TTL expiration works

### Rate Limiting
- [ ] Requests within limit pass
- [ ] Requests over limit blocked
- [ ] Headers set correctly

### Events
- [ ] Events published on actions
- [ ] Consumers receive events
- [ ] Failed messages requeued

## Related Code Files

**Create:**
- `tests/integration/redis_integration_test.go`
- `tests/integration/rabbitmq_integration_test.go`
- `tests/integration/health_test.go`

**Verify:**
- `shared/redis/client_test.go`
- `shared/rabbitmq/publisher_test.go`
- `docker-compose.yml`
- `Tiltfile.development`
- `Tiltfile.production`

## Todo Checklist

- [ ] Create redis_integration_test.go
- [ ] Create rabbitmq_integration_test.go
- [ ] Create health_test.go
- [ ] Run unit tests (all pass)
- [ ] Run integration tests with Docker mode
- [ ] Run integration tests with Serverless mode
- [ ] Verify caching works
- [ ] Verify rate limiting works
- [ ] Verify event publishing/consuming
- [ ] Document any mode-specific issues

## Success Criteria

- [ ] All unit tests pass
- [ ] Integration tests pass with real services
- [ ] Docker mode fully functional
- [ ] Serverless mode fully functional
- [ ] Cache hit rate > 80% for sessions
- [ ] Rate limiting blocks excess requests
- [ ] Events flow through system correctly

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Integration tests flaky | Medium | Use testcontainers, proper cleanup |
| Cloud service rate limits | Low | Use dedicated test instances |
| Mode switch issues | Medium | Test both modes in CI |

## Security Considerations

- Test credentials separate from prod
- No secrets in test code
- Cleanup test data after runs

## Unresolved Questions

1. Should we use testcontainers for isolated integration tests?
2. How to handle rate limits in integration tests?
3. Need performance benchmarks?

## Next Steps

- Update documentation
- Deploy to staging
- Monitor production metrics
