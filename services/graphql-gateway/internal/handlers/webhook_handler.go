package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc"
)

// WebhookHandler handles webhook requests from the indexer
type WebhookHandler struct {
	collectionClient pb.CollectionServiceClient
	webhookSecret    string
}

// IndexerWebhookPayload represents the webhook payload from indexer
type IndexerWebhookPayload struct {
	Event     string                 `json:"event"`
	ChainID   int64                  `json:"chainId"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(collectionConn *grpc.ClientConn, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{
		collectionClient: pb.NewCollectionServiceClient(collectionConn),
		webhookSecret:    webhookSecret,
	}
}

// HandleIndexerWebhook processes webhook events from the indexer
func (h *WebhookHandler) HandleIndexerWebhook(w http.ResponseWriter, r *http.Request) {
	// 1. Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	// 2. Verify HMAC signature
	signature := r.Header.Get("X-Webhook-Signature")
	if !h.verifySignature(body, signature) {
		writeJSONError(w, http.StatusUnauthorized, "invalid webhook signature")
		return
	}

	// 3. Parse payload
	var payload IndexerWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	// 4. Convert data to JSON string
	dataJSON, err := json.Marshal(payload.Data)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to marshal event data")
		return
	}

	// 5. Forward to Collection Service via gRPC
	resp, err := h.collectionClient.ProcessIndexerWebhook(r.Context(), &pb.ProcessIndexerWebhookRequest{
		Event:     payload.Event,
		ChainId:   payload.ChainID,
		Timestamp: payload.Timestamp,
		DataJson:  string(dataJSON),
	})

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 6. Return success response
	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": resp.Success,
		"message": resp.Message,
	})
}

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSONResponse(w, status, map[string]string{"error": message})
}

// verifySignature verifies the HMAC-SHA256 signature
func (h *WebhookHandler) verifySignature(payload []byte, signature string) bool {
	if h.webhookSecret == "" {
		// If no secret configured, skip verification (for development)
		return true
	}

	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
