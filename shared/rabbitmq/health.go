package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// CheckHealth checks if RabbitMQ connection is healthy
func CheckHealth() error {
	conn := GetConnection()
	if conn == nil {
		return ErrNotConnected
	}

	if conn.IsClosed() {
		return ErrConnectionClosed
	}

	ch := GetChannel()
	if ch == nil {
		return ErrChannelClosed
	}

	return nil
}

// Errors
var (
	ErrNotConnected     = &amqp.Error{Code: 0, Reason: "rabbitmq not initialized"}
	ErrConnectionClosed = &amqp.Error{Code: 0, Reason: "connection closed"}
	ErrChannelClosed    = &amqp.Error{Code: 0, Reason: "channel closed"}
)
