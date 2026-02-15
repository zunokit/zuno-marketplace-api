package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebhookValidator_ValidateEventType(t *testing.T) {
	validator := NewWebhookValidator(true)

	t.Run("Valid event types", func(t *testing.T) {
		validTypes := []string{
			"collection.created",
			"collection.minted",
			"collection.batch_minted",
		}

		for _, eventType := range validTypes {
			err := validator.ValidateEventType(eventType)
			assert.NoError(t, err, "Event type %s should be valid", eventType)
		}
	})

	t.Run("Invalid event type", func(t *testing.T) {
		err := validator.ValidateEventType("invalid.event")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid event type")
	})
}

func TestWebhookValidator_ValidateChainID(t *testing.T) {
	validator := NewWebhookValidator(true)

	t.Run("Valid chain IDs", func(t *testing.T) {
		validChainIDs := []int64{1, 5, 11155111, 137, 80001, 31337}

		for _, chainID := range validChainIDs {
			err := validator.ValidateChainID(chainID)
			assert.NoError(t, err, "Chain ID %d should be valid", chainID)
		}
	})

	t.Run("Invalid chain ID (zero)", func(t *testing.T) {
		err := validator.ValidateChainID(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid chain ID")
	})

	t.Run("Invalid chain ID (negative)", func(t *testing.T) {
		err := validator.ValidateChainID(-1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid chain ID")
	})
}

func TestWebhookValidator_ValidateTimestamp(t *testing.T) {
	validator := NewWebhookValidator(true)

	t.Run("Valid timestamp", func(t *testing.T) {
		err := validator.ValidateTimestamp(1700000000)
		assert.NoError(t, err)
	})

	t.Run("Invalid timestamp (zero)", func(t *testing.T) {
		err := validator.ValidateTimestamp(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid timestamp")
	})

	t.Run("Invalid timestamp (negative)", func(t *testing.T) {
		err := validator.ValidateTimestamp(-1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid timestamp")
	})
}

func TestWebhookValidator_ValidateCollectionCreated(t *testing.T) {
	validator := NewWebhookValidator(true)

	t.Run("Valid collection.created event", func(t *testing.T) {
		data := map[string]interface{}{
			"collectionAddress": "0x1234567890123456789012345678901234567890",
			"creator":           "0x0987654321098765432109876543210987654321",
			"tokenType":         "ERC721",
			"blockNumber":       float64(12345),
			"txHash":            "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"logIndex":          float64(0),
		}

		err := validator.ValidateCollectionCreated(data)
		assert.NoError(t, err)
	})

	t.Run("Missing required field", func(t *testing.T) {
		data := map[string]interface{}{
			"collectionAddress": "0x1234567890123456789012345678901234567890",
			// Missing: creator, tokenType, blockNumber, txHash, logIndex
		}

		err := validator.ValidateCollectionCreated(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required field")
	})

	t.Run("Invalid collection address (missing 0x)", func(t *testing.T) {
		data := map[string]interface{}{
			"collectionAddress": "1234567890123456789012345678901234567890",
			"creator":           "0x0987654321098765432109876543210987654321",
			"tokenType":         "ERC721",
			"blockNumber":       float64(12345),
			"txHash":            "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"logIndex":          float64(0),
		}

		err := validator.ValidateCollectionCreated(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "collectionAddress")
		assert.Contains(t, err.Error(), "invalid Ethereum address")
	})

	t.Run("Invalid token type", func(t *testing.T) {
		data := map[string]interface{}{
			"collectionAddress": "0x1234567890123456789012345678901234567890",
			"creator":           "0x0987654321098765432109876543210987654321",
			"tokenType":         "ERC20",
			"blockNumber":       float64(12345),
			"txHash":            "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"logIndex":          float64(0),
		}

		err := validator.ValidateCollectionCreated(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tokenType")
		assert.Contains(t, err.Error(), "invalid token type")
	})

	t.Run("Invalid transaction hash (wrong length)", func(t *testing.T) {
		data := map[string]interface{}{
			"collectionAddress": "0x1234567890123456789012345678901234567890",
			"creator":           "0x0987654321098765432109876543210987654321",
			"tokenType":         "ERC721",
			"blockNumber":       float64(12345),
			"txHash":            "0xabc123",
			"logIndex":          float64(0),
		}

		err := validator.ValidateCollectionCreated(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "txHash")
		assert.Contains(t, err.Error(), "invalid length")
	})
}

func TestWebhookValidator_ValidateCollectionMinted(t *testing.T) {
	validator := NewWebhookValidator(true)

	t.Run("Valid collection.minted event", func(t *testing.T) {
		data := map[string]interface{}{
			"contractAddress": "0x1234567890123456789012345678901234567890",
			"blockNumber":     float64(12346),
			"txHash":          "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"logIndex":        float64(1),
		}

		err := validator.ValidateCollectionMinted(data)
		assert.NoError(t, err)
	})

	t.Run("Missing required field", func(t *testing.T) {
		data := map[string]interface{}{
			"contractAddress": "0x1234567890123456789012345678901234567890",
			// Missing: blockNumber, txHash, logIndex
		}

		err := validator.ValidateCollectionMinted(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required field")
	})
}

func TestWebhookValidator_ValidateEthereumAddress(t *testing.T) {
	validator := NewWebhookValidator(true)

	t.Run("Valid addresses", func(t *testing.T) {
		validAddresses := []string{
			"0x1234567890123456789012345678901234567890",
			"0xabcdef1234567890abcdef1234567890abcdef12",
			"0xABCDEF1234567890ABCDEF1234567890ABCDEF12",
			"0x0000000000000000000000000000000000000001",
		}

		for _, address := range validAddresses {
			err := validator.ValidateEthereumAddress(address)
			assert.NoError(t, err, "Address %s should be valid", address)
		}
	})

	t.Run("Invalid address (empty)", func(t *testing.T) {
		err := validator.ValidateEthereumAddress("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "address is empty")
	})

	t.Run("Invalid address (missing 0x)", func(t *testing.T) {
		err := validator.ValidateEthereumAddress("1234567890123456789012345678901234567890")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing 0x prefix")
	})

	t.Run("Invalid address (wrong length)", func(t *testing.T) {
		err := validator.ValidateEthereumAddress("0x123456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid length")
	})

	t.Run("Invalid address (invalid hex)", func(t *testing.T) {
		err := validator.ValidateEthereumAddress("0xGHKLmnopqrstuvwxzyz123456789012345")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid hex characters")
	})
}

func TestWebhookValidator_ValidateTxHash(t *testing.T) {
	validator := NewWebhookValidator(true)

	t.Run("Valid transaction hash", func(t *testing.T) {
		txHash := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		err := validator.ValidateTxHash(txHash)
		assert.NoError(t, err)
	})

	t.Run("Invalid txHash (empty)", func(t *testing.T) {
		err := validator.ValidateTxHash("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "txHash is empty")
	})

	t.Run("Invalid txHash (missing 0x)", func(t *testing.T) {
		err := validator.ValidateTxHash("abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing 0x prefix")
	})

	t.Run("Invalid txHash (wrong length)", func(t *testing.T) {
		err := validator.ValidateTxHash("0xabc123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid length")
	})
}

func TestBuildEventID(t *testing.T) {
	t.Run("Build event ID correctly", func(t *testing.T) {
		txHash := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		logIndex := int64(5)

		eventID := BuildEventID(txHash, logIndex)
		expected := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890:5"

		assert.Equal(t, expected, eventID)
	})

	t.Run("Build event ID with log index 0", func(t *testing.T) {
		txHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		logIndex := int64(0)

		eventID := BuildEventID(txHash, logIndex)
		expected := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef:0"

		assert.Equal(t, expected, eventID)
	})
}

func TestWebhookValidator_ValidateRequest(t *testing.T) {
	validator := NewWebhookValidator(true)

	t.Run("Valid collection.created request", func(t *testing.T) {
		data := map[string]interface{}{
			"collectionAddress": "0x1234567890123456789012345678901234567890",
			"creator":           "0x0987654321098765432109876543210987654321",
			"tokenType":         "ERC721",
			"blockNumber":       float64(12345),
			"txHash":            "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"logIndex":          float64(0),
		}

		err := validator.ValidateRequest("collection.created", 1, 1700000000, data)
		assert.NoError(t, err)
	})

	t.Run("Invalid event type", func(t *testing.T) {
		data := map[string]interface{}{
			"collectionAddress": "0x1234567890123456789012345678901234567890",
			"creator":           "0x0987654321098765432109876543210987654321",
			"tokenType":         "ERC721",
			"blockNumber":       float64(12345),
			"txHash":            "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"logIndex":          float64(0),
		}

		err := validator.ValidateRequest("invalid.event", 1, 1700000000, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid event type")
	})
}
