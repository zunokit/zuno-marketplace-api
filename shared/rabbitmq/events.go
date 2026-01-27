package rabbitmq

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Event types
const (
	EventTypeAuthLogin      = "auth.login"
	EventTypeAuthLogout     = "auth.logout"
	EventTypeUserCreated    = "user.created"
	EventTypeUserUpdated    = "user.updated"
	EventTypeWalletCreated  = "wallet.created"
	EventTypeWalletUpdated  = "wallet.updated"
	EventTypeTransactionCreated = "transaction.created"
)

// NewEvent creates a new event
func NewEvent(eventType, source string, data map[string]interface{}) Event {
	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
	}
}

// ToJSON converts event to JSON bytes
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// RoutingKey returns the routing key for the event
func (e *Event) RoutingKey() string {
	return e.Type
}
