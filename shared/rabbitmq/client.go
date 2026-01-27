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
	connection *amqp.Connection
	channel    *amqp.Channel
	once       sync.Once
	mu         sync.RWMutex
)

const (
	exchangeName = "nft_events"
	exchangeType = "topic"
)

// Init initializes RabbitMQ connection and channel
func Init(url string) error {
	var initErr error
	once.Do(func() {
		initErr = connect(url)
	})
	return initErr
}

// connect establishes connection and sets up topology
func connect(url string) error {
	log.Printf("Connecting to RabbitMQ...")

	config := amqp.Config{
		Vhost:      "/",
		Heartbeat:  10 * time.Second,
		Locale:     "en_US",
		Properties: amqp.Table{
			"connection_name": "zuno-marketplace-api",
			"product":         "Zuno Marketplace",
		},
	}

	var err error
	connection, err = amqp.DialConfig(url, config)
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}

	log.Println("RabbitMQ connection established")

	// Create channel
	channel, err = connection.Channel()
	if err != nil {
		connection.Close()
		return fmt.Errorf("open channel: %w", err)
	}

	// Set QoS for fair dispatch
	err = channel.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		channel.Close()
		connection.Close()
		return fmt.Errorf("set qos: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		exchangeName, // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		channel.Close()
		connection.Close()
		return fmt.Errorf("declare exchange: %w", err)
	}

	log.Printf("RabbitMQ exchange '%s' declared", exchangeName)

	// Setup connection error handler
	err = connection.Close()
	connection, err = amqp.DialConfig(url, config)
	if err != nil {
		return fmt.Errorf("re-dial rabbitmq: %w", err)
	}

	channel, err = connection.Channel()
	if err != nil {
		connection.Close()
		return fmt.Errorf("re-open channel: %w", err)
	}

	channel.Qos(10, 0, false)
	channel.ExchangeDeclare(exchangeName, exchangeType, true, false, false, false, nil)

	// Handle connection closes
	go func() {
		errChan := make(chan *amqp.Error, 1)
		connection.NotifyClose(errChan)
		if err := <-errChan; err != nil {
			log.Printf("RabbitMQ connection closed: %v, attempting reconnect...", err)
			time.Sleep(5 * time.Second)
			mu.Lock()
			connection = nil
			channel = nil
			once = sync.Once{}
			mu.Unlock()
			Init(url)
		}
	}()

	return nil
}

func handleConnectionLoss(url string) {
	errChan := make(chan *amqp.Error, 1)
	connection.NotifyClose(errChan)

	if err := <-errChan; err != nil {
		log.Printf("RabbitMQ connection closed: %v, attempting reconnect...", err)
		time.Sleep(5 * time.Second)
		mu.Lock()
		connection = nil
		channel = nil
		once = sync.Once{}
		mu.Unlock()
		Init(url)
	}
}

// GetChannel returns the RabbitMQ channel
func GetChannel() *amqp.Channel {
	mu.RLock()
	defer mu.RUnlock()
	return channel
}

// GetConnection returns the RabbitMQ connection
func GetConnection() *amqp.Connection {
	mu.RLock()
	defer mu.RUnlock()
	return connection
}

// Close closes RabbitMQ connection and channel
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	var errs []error

	if channel != nil {
		if err := channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close channel: %w", err))
		}
		channel = nil
	}

	if connection != nil {
		if err := connection.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close connection: %w", err))
		}
		connection = nil
	}

	once = sync.Once{}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	return nil
}

// DeclareQueue declares and binds a queue
func DeclareQueue(queueName, routingKey string) (amqp.Queue, error) {
	ch := GetChannel()
	if ch == nil {
		return amqp.Queue{}, fmt.Errorf("rabbitmq not initialized")
	}

	queue, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("declare queue: %w", err)
	}

	err = ch.QueueBind(
		queue.Name,   // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("bind queue: %w", err)
	}

	log.Printf("Queue '%s' bound to exchange '%s' with routing key '%s'", queueName, exchangeName, routingKey)

	return queue, nil
}

// IsConnected returns true if RabbitMQ is connected
func IsConnected() bool {
	mu.RLock()
	defer mu.RUnlock()
	return connection != nil && !connection.IsClosed()
}

// PublishWithContext publishes a message with context
func PublishWithContext(ctx context.Context, routingKey string, body []byte) error {
	ch := GetChannel()
	if ch == nil {
		return fmt.Errorf("rabbitmq not initialized")
	}

	return ch.PublishWithContext(
		ctx,
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // persistent
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
}
