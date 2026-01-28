package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"time"
)

const (
	PublishTimeout = 5 * time.Second
	ConsumeTimeout = 30 * time.Second
)

// Publish publishes an event to RabbitMQ
func Publish(event Event) error {
	body, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), PublishTimeout)
	defer cancel()

	if err := PublishWithContext(ctx, event.RoutingKey(), body); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}

	log.Printf("Event published: %s [%s]", event.Type, event.ID)
	return nil
}

// PublishAsync publishes an event asynchronously
func PublishAsync(event Event) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		errChan <- Publish(event)
	}()
	return errChan
}

// PublishLoginEvent publishes a login event
func PublishLoginEvent(source string, data map[string]interface{}) error {
	event := NewEvent(EventTypeAuthLogin, source, data)
	return Publish(event)
}

// PublishUserCreatedEvent publishes a user created event
func PublishUserCreatedEvent(source string, data map[string]interface{}) error {
	event := NewEvent(EventTypeUserCreated, source, data)
	return Publish(event)
}

// PublishWalletCreatedEvent publishes a wallet created event
func PublishWalletCreatedEvent(source string, data map[string]interface{}) error {
	event := NewEvent(EventTypeWalletCreated, source, data)
	return Publish(event)
}
