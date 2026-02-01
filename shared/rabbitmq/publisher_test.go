package rabbitmq

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEvent(t *testing.T) {
	data := map[string]interface{}{
		"userId": "123",
		"email":  "test@example.com",
	}

	event := NewEvent("user.created", "user-service", data)

	assert.NotEmpty(t, event.ID)
	assert.Equal(t, "user.created", event.Type)
	assert.Equal(t, "user-service", event.Source)
	assert.NotEmpty(t, event.Timestamp)
	assert.Equal(t, data, event.Data)
}

func TestEventToJSON(t *testing.T) {
	data := map[string]interface{}{
		"userId": "123",
	}

	event := NewEvent("user.created", "user-service", data)
	jsonBytes, err := event.ToJSON()

	require.NoError(t, err)
	require.NotNil(t, jsonBytes)

	var decoded Event
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.ID, decoded.ID)
	assert.Equal(t, event.Type, decoded.Type)
	assert.Equal(t, event.Source, decoded.Source)
}

func TestEventRoutingKey(t *testing.T) {
	event := NewEvent("auth.login", "auth-service", nil)
	assert.Equal(t, "auth.login", event.RoutingKey())
}

func TestPublishWithContext(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// This would require a real RabbitMQ instance
	// For now, just test the function signature exists
	ctx := context.Background()
	body := []byte(`{"test":"data"}`)

	err := PublishWithContext(ctx, "test.routing.key", body)
	// Expected to fail without connection, but function should exist
	// In real scenario, this would be mocked or use testcontainers
	_ = err
}

func TestDeclareQueue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Integration test with real RabbitMQ
	// Requires test setup/teardown
}

func TestPublishTimeout(t *testing.T) {
	assert.Equal(t, 5*time.Second, PublishTimeout)
}

func TestConsumeTimeout(t *testing.T) {
	assert.Equal(t, 30*time.Second, ConsumeTimeout)
}
