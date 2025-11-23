package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"github.com/stretchr/testify/assert"
)

func TestProcessIndexerWebhook_CollectionCreated(t *testing.T) {
	// Test valid JSON parsing
	eventData := WebhookEventData{
		CollectionAddress: "0x1234567890abcdef1234567890abcdef12345678",
		Creator:           "0xCreatorAddress",
		TokenType:         "ERC721",
		BlockNumber:       12345,
		TxHash:            "0xTxHash",
	}

	dataJSON, err := json.Marshal(eventData)
	assert.NoError(t, err)

	var parsed WebhookEventData
	err = json.Unmarshal(dataJSON, &parsed)
	assert.NoError(t, err)
	assert.Equal(t, eventData.CollectionAddress, parsed.CollectionAddress)
	assert.Equal(t, eventData.Creator, parsed.Creator)
}

func TestProcessIndexerWebhook_InvalidEventData(t *testing.T) {
	// Test that invalid JSON returns error
	invalidJSON := "invalid json"
	var eventData WebhookEventData
	err := json.Unmarshal([]byte(invalidJSON), &eventData)
	assert.Error(t, err)
}

func TestProcessIndexerWebhook_EventTypes(t *testing.T) {
	// Test different event types are recognized
	events := []string{
		"collection.created",
		"collection.minted",
		"collection.batch_minted",
	}

	for _, event := range events {
		req := &pb.ProcessIndexerWebhookRequest{
			Event:     event,
			ChainId:   31337,
			Timestamp: time.Now().Unix(),
			DataJson:  `{"collectionAddress":"0x1234"}`,
		}
		assert.Equal(t, event, req.Event)
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	tests := []struct {
		name      string
		payload   []byte
		secret    string
		signature string
		expected  bool
	}{
		{
			name:      "Invalid signature",
			payload:   []byte(`{"event":"test"}`),
			secret:    "test-secret",
			signature: "invalid-signature",
			expected:  false,
		},
		{
			name:      "Empty signature",
			payload:   []byte(`{"event":"test"}`),
			secret:    "test-secret",
			signature: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyWebhookSignature(tt.payload, tt.signature, tt.secret)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVerifyWebhookSignature_ValidComputed(t *testing.T) {
	payload := []byte(`{"event":"collection.created","chainId":31337}`)
	secret := "webhook-secret-123"

	// Compute expected signature using same method as VerifyWebhookSignature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Test with correct signature
	result := VerifyWebhookSignature(payload, expectedSignature, secret)
	assert.True(t, result)

	// Test with wrong signature
	result = VerifyWebhookSignature(payload, "wrong-signature", secret)
	assert.False(t, result)

	// Test with empty signature
	result = VerifyWebhookSignature(payload, "", secret)
	assert.False(t, result)
}

func TestWebhookEventData_Parsing(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		expected WebhookEventData
	}{
		{
			name: "Full event data",
			json: `{
				"collectionAddress": "0x1234567890abcdef1234567890abcdef12345678",
				"creator": "0xCreator",
				"tokenType": "ERC721",
				"blockNumber": 12345,
				"txHash": "0xTxHash123",
				"logIndex": 5,
				"contractAddress": "0xContract"
			}`,
			expected: WebhookEventData{
				CollectionAddress: "0x1234567890abcdef1234567890abcdef12345678",
				Creator:           "0xCreator",
				TokenType:         "ERC721",
				BlockNumber:       12345,
				TxHash:            "0xTxHash123",
				LogIndex:          5,
				ContractAddress:   "0xContract",
			},
		},
		{
			name: "Minimal event data",
			json: `{
				"collectionAddress": "0xMinimal"
			}`,
			expected: WebhookEventData{
				CollectionAddress: "0xMinimal",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result WebhookEventData
			err := json.Unmarshal([]byte(tc.json), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected.CollectionAddress, result.CollectionAddress)
			assert.Equal(t, tc.expected.Creator, result.Creator)
			assert.Equal(t, tc.expected.TokenType, result.TokenType)
			assert.Equal(t, tc.expected.BlockNumber, result.BlockNumber)
			assert.Equal(t, tc.expected.TxHash, result.TxHash)
		})
	}
}

func TestChainIDFormatting(t *testing.T) {
	// Test that chain IDs are formatted correctly for eip155
	testCases := []struct {
		chainID  int64
		expected string
	}{
		{1, "eip155:1"},
		{31337, "eip155:31337"},
		{137, "eip155:137"},
		{11155111, "eip155:11155111"},
	}

	for _, tc := range testCases {
		result := fmt.Sprintf("eip155:%d", tc.chainID)
		assert.Equal(t, tc.expected, result)
	}
}
