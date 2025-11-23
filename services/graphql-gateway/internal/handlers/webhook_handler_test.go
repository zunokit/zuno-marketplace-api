package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// MockCollectionServiceClient mocks the gRPC collection service client
type MockCollectionServiceClient struct {
	mock.Mock
}

func (m *MockCollectionServiceClient) CreateCollection(ctx context.Context, req *pb.CreateCollectionRequest, opts ...grpc.CallOption) (*pb.CreateCollectionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreateCollectionResponse), args.Error(1)
}

func (m *MockCollectionServiceClient) GetCollection(ctx context.Context, req *pb.GetCollectionRequest, opts ...grpc.CallOption) (*pb.GetCollectionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetCollectionResponse), args.Error(1)
}

func (m *MockCollectionServiceClient) UpdateCollection(ctx context.Context, req *pb.UpdateCollectionRequest, opts ...grpc.CallOption) (*pb.UpdateCollectionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UpdateCollectionResponse), args.Error(1)
}

func (m *MockCollectionServiceClient) DeleteCollection(ctx context.Context, req *pb.DeleteCollectionRequest, opts ...grpc.CallOption) (*pb.DeleteCollectionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.DeleteCollectionResponse), args.Error(1)
}

func (m *MockCollectionServiceClient) ListCollections(ctx context.Context, req *pb.ListCollectionsRequest, opts ...grpc.CallOption) (*pb.ListCollectionsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ListCollectionsResponse), args.Error(1)
}

func (m *MockCollectionServiceClient) ProcessIndexerWebhook(ctx context.Context, req *pb.ProcessIndexerWebhookRequest, opts ...grpc.CallOption) (*pb.ProcessIndexerWebhookResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ProcessIndexerWebhookResponse), args.Error(1)
}

func (m *MockCollectionServiceClient) AddToAllowlist(ctx context.Context, req *pb.AddToAllowlistRequest, opts ...grpc.CallOption) (*pb.AddToAllowlistResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AddToAllowlistResponse), args.Error(1)
}

func (m *MockCollectionServiceClient) ListCollectionsByUser(ctx context.Context, req *pb.ListCollectionsByUserRequest, opts ...grpc.CallOption) (*pb.ListCollectionsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ListCollectionsResponse), args.Error(1)
}

func TestHandleIndexerWebhook_Success(t *testing.T) {
	mockClient := new(MockCollectionServiceClient)

	// Use empty secret to skip verification for this test
	handler := &WebhookHandler{
		collectionClient: mockClient,
		webhookSecret:    "",
	}

	// Setup mock expectations
	mockClient.On("ProcessIndexerWebhook", mock.Anything, mock.MatchedBy(func(req *pb.ProcessIndexerWebhookRequest) bool {
		return req.Event == "collection.created" && req.ChainId == 31337
	})).Return(&pb.ProcessIndexerWebhookResponse{
		Success: true,
		Message: "Collection indexed successfully",
	}, nil)

	// Create test request
	payload := IndexerWebhookPayload{
		Event:     "collection.created",
		ChainID:   31337,
		Timestamp: 1234567890,
		Data: map[string]interface{}{
			"collectionAddress": "0x1234567890abcdef1234567890abcdef12345678",
			"creator":           "0xCreatorAddress",
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/webhook/indexer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Execute
	handler.HandleIndexerWebhook(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Collection indexed successfully", response["message"])

	mockClient.AssertExpectations(t)
}

func TestHandleIndexerWebhook_InvalidSignature(t *testing.T) {
	mockClient := new(MockCollectionServiceClient)

	handler := &WebhookHandler{
		collectionClient: mockClient,
		webhookSecret:    "test-secret", // With secret
	}

	payload := IndexerWebhookPayload{
		Event:     "collection.created",
		ChainID:   31337,
		Timestamp: 1234567890,
		Data:      map[string]interface{}{},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/webhook/indexer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", "invalid-signature")

	w := httptest.NewRecorder()

	// Execute
	handler.HandleIndexerWebhook(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid webhook signature")
}

func TestHandleIndexerWebhook_ValidSignature(t *testing.T) {
	mockClient := new(MockCollectionServiceClient)
	secret := "webhook-secret-123"

	handler := &WebhookHandler{
		collectionClient: mockClient,
		webhookSecret:    secret,
	}

	// Setup mock expectations
	mockClient.On("ProcessIndexerWebhook", mock.Anything, mock.Anything).
		Return(&pb.ProcessIndexerWebhookResponse{
			Success: true,
			Message: "Success",
		}, nil)

	payload := IndexerWebhookPayload{
		Event:     "collection.created",
		ChainID:   31337,
		Timestamp: 1234567890,
		Data:      map[string]interface{}{"test": "data"},
	}
	body, _ := json.Marshal(payload)

	// Compute valid signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/api/webhook/indexer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)

	w := httptest.NewRecorder()

	// Execute
	handler.HandleIndexerWebhook(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockClient.AssertExpectations(t)
}

func TestHandleIndexerWebhook_InvalidJSON(t *testing.T) {
	mockClient := new(MockCollectionServiceClient)

	handler := &WebhookHandler{
		collectionClient: mockClient,
		webhookSecret:    "",
	}

	req := httptest.NewRequest(http.MethodPost, "/api/webhook/indexer", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Execute
	handler.HandleIndexerWebhook(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid JSON")
}

func TestHandleIndexerWebhook_GRPCError(t *testing.T) {
	mockClient := new(MockCollectionServiceClient)

	handler := &WebhookHandler{
		collectionClient: mockClient,
		webhookSecret:    "",
	}

	// Setup mock to return error
	mockClient.On("ProcessIndexerWebhook", mock.Anything, mock.Anything).
		Return(nil, assert.AnError)

	payload := IndexerWebhookPayload{
		Event:     "collection.created",
		ChainID:   31337,
		Timestamp: 1234567890,
		Data:      map[string]interface{}{},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/webhook/indexer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Execute
	handler.HandleIndexerWebhook(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockClient.AssertExpectations(t)
}

func TestVerifySignature(t *testing.T) {
	tests := []struct {
		name      string
		payload   []byte
		secret    string
		signature string
		expected  bool
	}{
		{
			name:      "Empty secret skips verification",
			payload:   []byte(`{"test":"data"}`),
			secret:    "",
			signature: "any-signature",
			expected:  true,
		},
		{
			name:      "Invalid signature with secret",
			payload:   []byte(`{"test":"data"}`),
			secret:    "my-secret",
			signature: "invalid",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &WebhookHandler{webhookSecret: tt.secret}
			result := handler.verifySignature(tt.payload, tt.signature)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVerifySignature_Valid(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"event":"test","chainId":1}`)

	// Compute valid signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	handler := &WebhookHandler{webhookSecret: secret}
	result := handler.verifySignature(payload, signature)
	assert.True(t, result)
}
