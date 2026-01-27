package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Handler handles incoming messages
type Handler func(ctx context.Context, event Event) error

// Consumer consumes messages from a queue
type Consumer struct {
	queueName string
	routingKey string
	handler   Handler
}

// NewConsumer creates a new consumer
func NewConsumer(queueName, routingKey string, handler Handler) *Consumer {
	return &Consumer{
		queueName:  queueName,
		routingKey: routingKey,
		handler:    handler,
	}
}

// Start starts consuming messages
func (c *Consumer) Start() error {
	queue, err := DeclareQueue(c.queueName, c.routingKey)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	ch := GetChannel()
	if ch == nil {
		return fmt.Errorf("rabbitmq not initialized")
	}

	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer tag
		false,      // auto-ack (manual ack)
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("register consumer: %w", err)
	}

	log.Printf("Consumer started on queue '%s' with routing key '%s'", c.queueName, c.routingKey)

	go c.consumeLoop(msgs)

	return nil
}

func (c *Consumer) consumeLoop(msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		if err := c.handleMessage(msg); err != nil {
			log.Printf("Error handling message: %v", err)
			msg.Nack(false, true) // requeue on error
		} else {
			msg.Ack(false)
		}
	}
}

func (c *Consumer) handleMessage(msg amqp.Delivery) error {
	ctx, cancel := context.WithTimeout(context.Background(), ConsumeTimeout)
	defer cancel()

	var event Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	log.Printf("Received event: %s [%s]", event.Type, event.ID)

	if err := c.handler(ctx, event); err != nil {
		return fmt.Errorf("handler error: %w", err)
	}

	return nil
}

// ConsumeMessages is a helper for simple message consumption
func ConsumeMessages(queueName, routingKey string, handler Handler) error {
	consumer := NewConsumer(queueName, routingKey, handler)
	return consumer.Start()
}
