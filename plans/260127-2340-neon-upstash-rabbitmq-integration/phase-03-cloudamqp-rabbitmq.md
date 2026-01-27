---
title: "Phase 3: CloudAMQP RabbitMQ Integration"
description: "Add amqp091-go client, wrapper package, and implement pub/sub for events"
status: pending
priority: P1
effort: 4h
---

# Phase 3: CloudAMQP RabbitMQ Integration

## Overview

Add RabbitMQ client using amqp091-go, create shared wrapper package, implement pub/sub for domain events (user.created, wallet.updated, auth.login).

## Context

**Current State:**
- Config has RabbitMQConfig with GetURL()
- No RabbitMQ client initialization
- No event system

**Required:**
- Add amqp091-go dependency
- Create shared/rabbitmq package
- Initialize client in main.go
- Define event schema
- Publish events from services
- Consume events in handlers

## Requirements

### Functional
- [ ] Add amqp091-go to go.mod
- [ ] Create shared/rabbitmq wrapper package
- [ ] Initialize RabbitMQ connection in services
- [ ] Define event schema (JSON)
- [ ] Implement publisher for auth events
- [ ] Implement publisher for user events
- [ ] Implement publisher for wallet events
- [ ] Implement consumer example
- [ ] Graceful shutdown (close channels)

### Non-Functional
- Connection recovery on failure
- Message acknowledgment
- Dead letter queue for failures
- Idempotent event handlers

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  shared/rabbitmq/                                            │
│  ├── client.go       (connection mgmt, channel pool)        │
│  ├── publisher.go    (Publish with routing key)             │
│  ├── consumer.go     (Consume with ack/reject)              │
│  ├── events.go       (Event type definitions)                │
│  └── health.go       (Connection check)                      │
└─────────────────────────────────────────────────────────────┘
         │
         ├──> auth-service ──> Publish: auth.login, auth.logout
         ├──> user-service  ──> Publish: user.created, user.updated
         └──> wallet-service ──> Publish: wallet.created, wallet.updated
```

### Event Schema

```json
{
  "id": "uuid",
  "type": "user.created",
  "source": "user-service",
  "timestamp": "2025-01-27T10:00:00Z",
  "data": {
    "userId": "123",
    "walletAddress": "0x123..."
  }
}
```

### Exchange/Queue Topology

```
Exchange: nft_events (topic)
  ├── queue.auth_events ──> routing key: auth.*
  ├── queue.user_events ──> routing key: user.*
  └── queue.wallet_events ──> routing key: wallet.*
```

## Implementation Steps

### 1. Add Dependency

```bash
go get github.com/rabbitmq/amqp091-go
```

### 2. Create shared/rabbitmq Package

**File:** `shared/rabbitmq/client.go`

```go
package rabbitmq

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

var (
    conn      *amqp.Connection
    channelMu sync.Mutex
    channels  []*amqp.Channel
)

// Config holds RabbitMQ configuration
type Config struct {
    URL string
}

// Init initializes the RabbitMQ connection
func Init(cfg Config) error {
    var err error

    // Connect with retry logic
    for i := 0; i < 5; i++ {
        conn, err = amqp.Dial(cfg.URL)
        if err == nil {
            break
        }
        log.Printf("RabbitMQ connection attempt %d failed: %v", i+1, err)
        time.Sleep(time.Second * time.Duration(i+1))
    }

    if err != nil {
        return fmt.Errorf("rabbitmq dial failed after retries: %w", err)
    }

    log.Println("RabbitMQ connected")

    // Setup topology
    if err := setupTopology(); err != nil {
        return fmt.Errorf("setup topology: %w", err)
    }

    return nil
}

// setupTopology creates exchanges and queues
func setupTopology() error {
    ch, err := GetChannel()
    if err != nil {
        return err
    }

    // Declare exchange
    err = ch.ExchangeDeclare(
        "nft_events", // name
        "topic",      // type
        true,         // durable
        false,        // auto-deleted
        false,        // internal
        false,        // no-wait
        nil,          // arguments
    )
    if err != nil {
        return fmt.Errorf("declare exchange: %w", err)
    }

    // Declare queues
    queues := []string{"auth_events", "user_events", "wallet_events"}
    for _, q := range queues {
        _, err := ch.QueueDeclare(
            q,     // name
            true,  // durable
            false, // delete when unused
            false, // exclusive
            false, // no-wait
            nil,   // arguments
        )
        if err != nil {
            return fmt.Errorf("declare queue %s: %w", q, err)
        }

        // Bind queue to exchange
        routingKey := fmt.Sprintf("%s.*", q[:len(q)-6]) // auth_events -> auth.*
        err = ch.QueueBind(
            q,
            routingKey,
            "nft_events",
            false,
            nil,
        )
        if err != nil {
            return fmt.Errorf("bind queue %s: %w", q, err)
        }
    }

    return nil
}

// GetChannel returns a channel from the pool
func GetChannel() (*amqp.Channel, error) {
    channelMu.Lock()
    defer channelMu.Unlock()

    if conn == nil || conn.IsClosed() {
        return nil, fmt.Errorf("connection not open")
    }

    ch, err := conn.Channel()
    if err != nil {
        return nil, fmt.Errorf("open channel: %w", err)
    }

    channels = append(channels, ch)
    return ch, nil
}

// Close closes the RabbitMQ connection
func Close() error {
    if conn != nil && !conn.IsClosed() {
        return conn.Close()
    }
    return nil
}
```

**File:** `shared/rabbitmq/events.go`

```go
package rabbitmq

import (
    "time"

    "github.com/google/uuid"
)

// EventType represents the type of event
type EventType string

const (
    EventAuthLogin    EventType = "auth.login"
    EventAuthLogout   EventType = "auth.logout"
    EventUserCreated  EventType = "user.created"
    EventUserUpdated  EventType = "user.updated"
    EventWalletCreated EventType = "wallet.created"
    EventWalletUpdated EventType = "wallet.updated"
)

// Event represents a domain event
type Event struct {
    ID        string    `json:"id"`
    Type      EventType `json:"type"`
    Source    string    `json:"source"`
    Timestamp time.Time `json:"timestamp"`
    Data      any       `json:"data"`
}

// NewEvent creates a new event
func NewEvent(eventType EventType, source string, data any) Event {
    return Event{
        ID:        uuid.New().String(),
        Type:      eventType,
        Source:    source,
        Timestamp: time.Now().UTC(),
        Data:      data,
    }
}
```

**File:** `shared/rabbitmq/publisher.go`

```go
package rabbitmq

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher publishes events to RabbitMQ
type Publisher struct {
    exchange string
}

// NewPublisher creates a new publisher
func NewPublisher() *Publisher {
    return &Publisher{
        exchange: "nft_events",
    }
}

// Publish publishes an event
func (p *Publisher) Publish(ctx context.Context, routingKey string, event Event) error {
    ch, err := GetChannel()
    if err != nil {
        return fmt.Errorf("get channel: %w", err)
    }

    body, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("marshal event: %w", err)
    }

    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    err = ch.PublishWithContext(
        ctx,
        p.exchange,
        routingKey,
        false, // mandatory
        false, // immediate
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,
            MessageId:    event.ID,
            Timestamp:    time.Now(),
        },
    )

    if err != nil {
        return fmt.Errorf("publish: %w", err)
    }

    log.Printf("Published event: %s (%s)", event.Type, event.ID)
    return nil
}

// PublishAuthLogin publishes auth login event
func (p *Publisher) PublishAuthLogin(ctx context.Context, userID, walletAddress string) error {
    event := NewEvent(EventAuthLogin, "auth-service", map[string]string{
        "userId":         userID,
        "walletAddress": walletAddress,
    })
    return p.Publish(ctx, "auth.login", event)
}

// PublishUserCreated publishes user created event
func (p *Publisher) PublishUserCreated(ctx context.Context, userID, walletAddress string) error {
    event := NewEvent(EventUserCreated, "user-service", map[string]string{
        "userId":         userID,
        "walletAddress": walletAddress,
    })
    return p.Publish(ctx, "user.created", event)
}
```

**File:** `shared/rabbitmq/consumer.go`

```go
package rabbitmq

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    amqp "github.com/rabbitmq/amqp091-go"
)

// Handler handles consumed events
type Handler func(ctx context.Context, event Event) error

// Consumer consumes events from RabbitMQ
type Consumer struct {
    queue   string
    handler Handler
}

// NewConsumer creates a new consumer
func NewConsumer(queue string, handler Handler) *Consumer {
    return &Consumer{
        queue:   queue,
        handler: handler,
    }
}

// Consume starts consuming messages
func (c *Consumer) Consume(ctx context.Context) error {
    ch, err := GetChannel()
    if err != nil {
        return fmt.Errorf("get channel: %w", err)
    }

    msgs, err := ch.Consume(
        c.queue,
        "",    // consumer tag
        false, // auto-ack (we'll manually ack)
        false, // exclusive
        false, // no-local
        false, // no-wait
        nil,   // args
    )
    if err != nil {
        return fmt.Errorf("consume: %w", err)
    }

    log.Printf("Consuming from queue: %s", c.queue)

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg, ok := <-msgs:
            if !ok {
                return fmt.Errorf("message channel closed")
            }

            // Parse event
            var event Event
            if err := json.Unmarshal(msg.Body, &event); err != nil {
                log.Printf("Failed to unmarshal event: %v", err)
                msg.Nack(false, false) // Reject, don't requeue
                continue
            }

            // Handle event
            if err := c.handler(ctx, event); err != nil {
                log.Printf("Handler error for event %s: %v", event.ID, err)
                msg.Nack(false, true) // Requeue on error
                continue
            }

            msg.Ack(false) // Acknowledge success
        }
    }
}
```

**File:** `shared/rabbitmq/health.go`

```go
package rabbitmq

import (
    "context"
    "time"
)

// Health checks RabbitMQ connectivity
func Health(ctx context.Context) error {
    if conn == nil || conn.IsClosed() {
        return fmt.Errorf("rabbitmq connection not open")
    }

    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    // Try to get a channel
    _, err := GetChannel()
    return err
}
```

**File:** `shared/rabbitmq/publisher_test.go`

```go
package rabbitmq

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestPublisher(t *testing.T) {
    if conn == nil {
        t.Skip("RabbitMQ not initialized")
    }

    ctx := context.Background()
    pub := NewPublisher()

    event := NewEvent(EventAuthLogin, "test", map[string]string{"userId": "123"})

    err := pub.Publish(ctx, "auth.login", event)
    require.NoError(t, err)
}
```

### 3. Initialize in Services

**File:** `services/auth-service/cmd/main.go`

```go
import (
    "github.com/zunokit/zuno-marketplace-api/shared/rabbitmq"
)

func main() {
    // ... existing code ...

    // Initialize RabbitMQ
    if cfg.RabbitMQ.GetURL() != "" {
        if err := rabbitmq.Init(rabbitmq.Config{URL: cfg.RabbitMQ.GetURL()}); err != nil {
            log.Printf("RabbitMQ init failed (continuing without events): %v", err)
        } else {
            log.Println("RabbitMQ connected")
            defer rabbitmq.Close()
        }
    }

    // ... rest of code ...
}
```

### 4. Publish Events from Services

**File:** `services/auth-service/internal/server/auth_server.go`

```go
import (
    "github.com/zunokit/zuno-marketplace-api/shared/rabbitmq"
)

type AuthServer struct {
    // ... existing fields ...
    eventPublisher *rabbitmq.Publisher
}

func NewAuthServer(..., pub *rabbitmq.Publisher) *AuthServer {
    return &AuthServer{
        // ... existing fields ...
        eventPublisher: pub,
    }
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
    // ... existing login logic ...

    // Publish event
    if s.eventPublisher != nil {
        _ = s.eventPublisher.PublishAuthLogin(ctx, userID, walletAddress)
    }

    // ... return response ...
}
```

### 5. Consumer Example

**File:** `services/user-service/cmd/consumer.go`

```go
package main

import (
    "context"
    "log"

    "github.com/zunokit/zuno-marketplace-api/shared/rabbitmq"
)

func startEventConsumer(ctx context.Context) {
    consumer := rabbitmq.NewConsumer("user_events", func(ctx context.Context, event rabbitmq.Event) error {
        log.Printf("Received event: %s", event.Type)

        switch event.Type {
        case rabbitmq.EventAuthLogin:
            // Handle auth login event
            log.Printf("User logged in: %v", event.Data)
        }

        return nil
    })

    go func() {
        if err := consumer.Consume(ctx); err != nil {
            log.Printf("Consumer error: %v", err)
        }
    }()
}
```

## Related Code Files

**Create:**
- `shared/rabbitmq/client.go`
- `shared/rabbitmq/events.go`
- `shared/rabbitmq/publisher.go`
- `shared/rabbitmq/consumer.go`
- `shared/rabbitmq/health.go`
- `shared/rabbitmq/publisher_test.go`
- `services/user-service/cmd/consumer.go`

**Modify:**
- `go.mod` - Add amqp091-go
- `services/auth-service/cmd/main.go` - Init RabbitMQ
- `services/user-service/cmd/main.go` - Init RabbitMQ
- `services/wallet-service/cmd/main.go` - Init RabbitMQ
- `services/auth-service/internal/server/auth_server.go` - Publish events
- `services/user-service/internal/server/user_server.go` - Publish events
- `services/wallet-service/internal/server/wallet_server.go` - Publish events

**Verify:**
- `.env.development.example` - CLOUDAMQP_URL format
- `.env.production.example` - CLOUDAMQP_URL format

## Todo Checklist

- [ ] Add amqp091-go dependency
- [ ] Create shared/rabbitmq/client.go
- [ ] Create shared/rabbitmq/events.go
- [ ] Create shared/rabbitmq/publisher.go
- [ ] Create shared/rabbitmq/consumer.go
- [ ] Create shared/rabbitmq/health.go
- [ ] Create shared/rabbitmq/publisher_test.go
- [ ] Initialize RabbitMQ in auth-service main.go
- [ ] Initialize RabbitMQ in user-service main.go
- [ ] Initialize RabbitMQ in wallet-service main.go
- [ ] Publish auth.login events from auth-service
- [ ] Publish user.created events from user-service
- [ ] Publish wallet.created events from wallet-service
- [ ] Create example consumer in user-service
- [ ] Add graceful shutdown for RabbitMQ
- [ ] Test with real CloudAMQP instance

## Success Criteria

- [ ] All services connect to CloudAMQP successfully
- [ ] Events are published to correct exchanges
- [ ] Consumers receive and process events
- [ ] Message acknowledgments work
- [ ] Failed messages are requeued
- [ ] Docker mode still works (RabbitMQ container)
- [ ] Serverless mode uses CloudAMQP

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| CloudAMQP connection limits | Medium | Use connection pooling, monitor |
| Message loss on failure | High | Persistent messages, publisher confirms |
| Slow consumers block queue | Medium | Multiple consumer instances, prefetch |
| Duplicate events | Low | Idempotent handlers |

## Security Considerations

- TLS for CloudAMQP (amqps://)
- Connection string in env
- No sensitive data in event payloads
- Validate event schemas

## Next Steps

- **Phase 4:** Testing & validation
