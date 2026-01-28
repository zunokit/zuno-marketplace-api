package main

import (
	"context"
	"log"

	sharedrabbitmq "github.com/zunokit/zuno-marketplace-api/shared/rabbitmq"
)

// startEventConsumer starts the event consumer for user service
func startEventConsumer() {
	handler := func(ctx context.Context, event sharedrabbitmq.Event) error {
		log.Printf("User service received event: type=%s, id=%s, source=%s",
			event.Type, event.ID, event.Source)

		// Handle different event types
		switch event.Type {
		case sharedrabbitmq.EventTypeAuthLogin:
			handleAuthLogin(ctx, event)
		case sharedrabbitmq.EventTypeUserCreated:
			handleUserCreated(ctx, event)
		case sharedrabbitmq.EventTypeWalletCreated:
			handleWalletCreated(ctx, event)
		default:
			log.Printf("Unhandled event type: %s", event.Type)
		}

		return nil
	}

	// Consume user events
	if err := sharedrabbitmq.ConsumeMessages("user_events", "user.*", handler); err != nil {
		log.Printf("Failed to start consumer: %v", err)
	}
}

func handleAuthLogin(ctx context.Context, event sharedrabbitmq.Event) {
	log.Printf("Auth login event: %+v", event.Data)
}

func handleUserCreated(ctx context.Context, event sharedrabbitmq.Event) {
	log.Printf("User created event: %+v", event.Data)
}

func handleWalletCreated(ctx context.Context, event sharedrabbitmq.Event) {
	log.Printf("Wallet created event: %+v", event.Data)
}
