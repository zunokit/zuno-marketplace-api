package service

import (
	"testing"

	"github.com/spruceid/siwe-go"
)

func TestSIWEService_ParseMessage(t *testing.T) {
	service := NewSIWEService()

	validMessage := `localhost wants you to sign in with your Ethereum account:
0x1234567890123456789012345678901234567890

URI: http://localhost
Version: 1
Chain ID: 1
Nonce: test-nonce-12345
Issued At: 2024-01-01T00:00:00Z`

	message, err := service.ParseMessage(validMessage)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if message.GetDomain() != "localhost" {
		t.Errorf("Expected domain 'localhost', got '%s'", message.GetDomain())
	}

	if message.GetNonce() != "test-nonce-12345" {
		t.Errorf("Expected nonce 'test-nonce-12345', got '%s'", message.GetNonce())
	}
}

func TestSIWEService_ValidateSIWEMessage(t *testing.T) {
	service := NewSIWEService()

	message := &siwe.Message{
		Domain:  "localhost",
		ChainID: 1,
		Nonce:   "test-nonce",
	}

	err := service.ValidateSIWEMessage(message, "localhost", "eip155:1", "test-nonce")
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}

	// Test domain mismatch
	err = service.ValidateSIWEMessage(message, "wrongdomain", "eip155:1", "test-nonce")
	if err == nil {
		t.Error("Expected domain mismatch error")
	}

	// Test chain ID mismatch
	err = service.ValidateSIWEMessage(message, "localhost", "eip155:137", "test-nonce")
	if err == nil {
		t.Error("Expected chain ID mismatch error")
	}

	// Test nonce mismatch
	err = service.ValidateSIWEMessage(message, "localhost", "eip155:1", "wrong-nonce")
	if err == nil {
		t.Error("Expected nonce mismatch error")
	}
}
