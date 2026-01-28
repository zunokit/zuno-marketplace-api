package server

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
)

var (
	// Validation errors
	ErrInvalidEvent         = errors.New("invalid event type")
	ErrMissingRequiredField = errors.New("missing required field")
	ErrInvalidAddress       = errors.New("invalid Ethereum address format")
	ErrInvalidTxHash        = errors.New("invalid transaction hash format")
	ErrInvalidChainID       = errors.New("invalid chain ID format")
	ErrInvalidTokenType     = errors.New("invalid token type")
	ErrInvalidBlockNumber   = errors.New("invalid block number")
	ErrInvalidLogIndex      = errors.New("invalid log index")
	ErrInvalidTimestamp     = errors.New("invalid timestamp")
)

// WebhookValidator validates webhook payloads
type WebhookValidator struct {
	// allowExternalCollections determines if we should accept events for collections not in our DB
	allowExternalCollections bool
}

// NewWebhookValidator creates a new webhook validator
func NewWebhookValidator(allowExternalCollections bool) *WebhookValidator {
	return &WebhookValidator{
		allowExternalCollections: allowExternalCollections,
	}
}

// ValidateRequest validates the webhook request
func (v *WebhookValidator) ValidateRequest(eventType string, chainID int64, timestamp int64, data map[string]interface{}) error {
	// Validate event type
	if err := v.ValidateEventType(eventType); err != nil {
		return err
	}

	// Validate chain ID
	if err := v.ValidateChainID(chainID); err != nil {
		return err
	}

	// Validate timestamp
	if err := v.ValidateTimestamp(timestamp); err != nil {
		return err
	}

	// Validate event-specific data
	switch eventType {
	case "collection.created":
		return v.ValidateCollectionCreated(data)
	case "collection.minted":
		return v.ValidateCollectionMinted(data)
	case "collection.batch_minted":
		return v.ValidateCollectionBatchMinted(data)
	default:
		return fmt.Errorf("%w: %s", ErrInvalidEvent, eventType)
	}
}

// ValidateEventType validates the event type
func (v *WebhookValidator) ValidateEventType(eventType string) error {
	validTypes := []string{
		"collection.created",
		"collection.minted",
		"collection.batch_minted",
	}

	for _, validType := range validTypes {
		if eventType == validType {
			return nil
		}
	}

	return fmt.Errorf("%w: %s (must be one of: %v)", ErrInvalidEvent, eventType, validTypes)
}

// ValidateChainID validates the chain ID
func (v *WebhookValidator) ValidateChainID(chainID int64) error {
	if chainID <= 0 {
		return fmt.Errorf("%w: chain ID must be positive, got %d", ErrInvalidChainID, chainID)
	}

	// Accept any positive chain ID (allows custom chains for testing)
	return nil
}

// ValidateTimestamp validates the timestamp
func (v *WebhookValidator) ValidateTimestamp(timestamp int64) error {
	if timestamp <= 0 {
		return fmt.Errorf("%w: timestamp must be positive, got %d", ErrInvalidTimestamp, timestamp)
	}

	// Optionally check if timestamp is reasonable (not too far in the future/past)
	// For now, we accept any positive timestamp

	return nil
}

// ValidateCollectionCreated validates collection.created event data
func (v *WebhookValidator) ValidateCollectionCreated(data map[string]interface{}) error {
	// Required fields
	requiredFields := map[string]string{
		"collectionAddress": "string",
		"creator":           "string",
		"tokenType":         "string",
		"blockNumber":       "float64",
		"txHash":            "string",
		"logIndex":          "float64",
	}

	// Check required fields exist and have correct types
	for field, expectedType := range requiredFields {
		value, exists := data[field]
		if !exists {
			return fmt.Errorf("%w: %s", ErrMissingRequiredField, field)
		}

		if err := v.validateFieldType(field, value, expectedType); err != nil {
			return err
		}
	}

	// Validate collection address
	collectionAddress := data["collectionAddress"].(string)
	if err := v.ValidateEthereumAddress(collectionAddress); err != nil {
		return fmt.Errorf("collectionAddress: %w", err)
	}

	// Validate creator address
	creator := data["creator"].(string)
	if err := v.ValidateEthereumAddress(creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate token type
	tokenType := data["tokenType"].(string)
	if err := v.ValidateTokenType(tokenType); err != nil {
		return fmt.Errorf("tokenType: %w", err)
	}

	// Validate block number
	blockNumber := int64(data["blockNumber"].(float64))
	if err := v.ValidateBlockNumber(blockNumber); err != nil {
		return fmt.Errorf("blockNumber: %w", err)
	}

	// Validate transaction hash
	txHash := data["txHash"].(string)
	if err := v.ValidateTxHash(txHash); err != nil {
		return fmt.Errorf("txHash: %w", err)
	}

	// Validate log index
	logIndex := int64(data["logIndex"].(float64))
	if err := v.ValidateLogIndex(logIndex); err != nil {
		return fmt.Errorf("logIndex: %w", err)
	}

	return nil
}

// ValidateCollectionMinted validates collection.minted event data
func (v *WebhookValidator) ValidateCollectionMinted(data map[string]interface{}) error {
	// Required fields
	requiredFields := map[string]string{
		"contractAddress": "string",
		"blockNumber":     "float64",
		"txHash":          "string",
		"logIndex":        "float64",
	}

	// Check required fields
	for field, expectedType := range requiredFields {
		value, exists := data[field]
		if !exists {
			return fmt.Errorf("%w: %s", ErrMissingRequiredField, field)
		}

		if err := v.validateFieldType(field, value, expectedType); err != nil {
			return err
		}
	}

	// Validate contract address
	contractAddress := data["contractAddress"].(string)
	if err := v.ValidateEthereumAddress(contractAddress); err != nil {
		return fmt.Errorf("contractAddress: %w", err)
	}

	// Validate block number
	blockNumber := int64(data["blockNumber"].(float64))
	if err := v.ValidateBlockNumber(blockNumber); err != nil {
		return fmt.Errorf("blockNumber: %w", err)
	}

	// Validate transaction hash
	txHash := data["txHash"].(string)
	if err := v.ValidateTxHash(txHash); err != nil {
		return fmt.Errorf("txHash: %w", err)
	}

	// Validate log index
	logIndex := int64(data["logIndex"].(float64))
	if err := v.ValidateLogIndex(logIndex); err != nil {
		return fmt.Errorf("logIndex: %w", err)
	}

	return nil
}

// ValidateCollectionBatchMinted validates collection.batch_minted event data
func (v *WebhookValidator) ValidateCollectionBatchMinted(data map[string]interface{}) error {
	// Same validation as collection.minted for now
	// In the future, we might validate batch-specific fields (batchSize, etc.)
	return v.ValidateCollectionMinted(data)
}

// ValidateEthereumAddress validates an Ethereum address
func (v *WebhookValidator) ValidateEthereumAddress(address string) error {
	if address == "" {
		return fmt.Errorf("%w: address is empty", ErrInvalidAddress)
	}

	// Check for 0x prefix
	if !strings.HasPrefix(address, "0x") {
		return fmt.Errorf("%w: missing 0x prefix", ErrInvalidAddress)
	}

	// Check length (0x + 40 hex characters)
	if len(address) != 42 {
		return fmt.Errorf("%w: invalid length (expected 42, got %d)", ErrInvalidAddress, len(address))
	}

	// Check hex characters
	hexPart := address[2:]
	matched, _ := regexp.MatchString("^[0-9a-fA-F]{40}$", hexPart)
	if !matched {
		return fmt.Errorf("%w: invalid hex characters", ErrInvalidAddress)
	}

	return nil
}

// ValidateTxHash validates a transaction hash
func (v *WebhookValidator) ValidateTxHash(txHash string) error {
	if txHash == "" {
		return fmt.Errorf("%w: txHash is empty", ErrInvalidTxHash)
	}

	// Check for 0x prefix
	if !strings.HasPrefix(txHash, "0x") {
		return fmt.Errorf("%w: missing 0x prefix", ErrInvalidTxHash)
	}

	// Check length (0x + 64 hex characters)
	if len(txHash) != 66 {
		return fmt.Errorf("%w: invalid length (expected 66, got %d)", ErrInvalidTxHash, len(txHash))
	}

	// Check hex characters
	hexPart := txHash[2:]
	matched, _ := regexp.MatchString("^[0-9a-fA-F]{64}$", hexPart)
	if !matched {
		return fmt.Errorf("%w: invalid hex characters", ErrInvalidTxHash)
	}

	return nil
}

// ValidateTokenType validates the token type
func (v *WebhookValidator) ValidateTokenType(tokenType string) error {
	validTypes := []string{
		string(models.TokenStandardERC721),
		string(models.TokenStandardERC1155),
	}

	for _, validType := range validTypes {
		if tokenType == validType {
			return nil
		}
	}

	return fmt.Errorf("%w: %s (must be ERC721 or ERC1155)", ErrInvalidTokenType, tokenType)
}

// ValidateBlockNumber validates a block number
func (v *WebhookValidator) ValidateBlockNumber(blockNumber int64) error {
	if blockNumber < 0 {
		return fmt.Errorf("%w: block number cannot be negative", ErrInvalidBlockNumber)
	}

	return nil
}

// ValidateLogIndex validates a log index
func (v *WebhookValidator) ValidateLogIndex(logIndex int64) error {
	if logIndex < 0 {
		return fmt.Errorf("%w: log index cannot be negative", ErrInvalidLogIndex)
	}

	return nil
}

// validateFieldType validates the type of a field
func (v *WebhookValidator) validateFieldType(fieldName string, value interface{}, expectedType string) error {
	if value == nil {
		return fmt.Errorf("%w: %s is nil", ErrMissingRequiredField, fieldName)
	}

	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%w: %s must be string, got %T", ErrMissingRequiredField, fieldName, value)
		}
	case "float64":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("%w: %s must be number, got %T", ErrMissingRequiredField, fieldName, value)
		}
	}

	return nil
}

// BuildEventID builds a unique event ID from txHash and logIndex
func BuildEventID(txHash string, logIndex int64) string {
	return fmt.Sprintf("%s:%d", txHash, logIndex)
}
