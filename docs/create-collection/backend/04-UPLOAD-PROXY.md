# PHASE 4: UPLOAD PROXY (Metadata Service Integration)

**Phase**: 4 of 6
**Dependencies**: Phase 3 (GraphQL Gateway) complete
**Estimated Effort**: ~300 lines Go
**Output**: Upload endpoints at `/api/upload/*`

---

## 🎯 Overview

Add upload proxy endpoints to GraphQL Gateway để:
1. ✅ **Secure API key** (không expose ở frontend)
2. ✅ **Centralized access control** (auth required)
3. ✅ **File validation** (size, type)
4. ✅ **Audit logging** (track uploads)
5. ✅ **Forward to Metadata Service** (thin proxy)

**Key Decision**: KHÔNG tạo service mới, chỉ add endpoints vào Gateway!

**⚠️ IMPORTANT**: IPFS pinning is **ASYNC** (background job), not immediate!
- Upload returns `isPinned: false` initially
- Pinning happens via cron job (~10-30 seconds)
- Must poll `/api/metadata/{id}` until `isPinned: true`

**📖 Full Spec**: See [`METADATA-SERVICE-SPEC.md`](../METADATA-SERVICE-SPEC.md)

---

## 🏗️ Architecture

```
Frontend                Gateway                Metadata Service
   │                       │                          │
   │  1. Upload file       │                          │
   │  (with JWT)           │                          │
   ├──────────────────────>│                          │
   │                       │  2. Verify JWT           │
   │                       │                          │
   │                       │  3. Validate file        │
   │                       │     (size, type)         │
   │                       │                          │
   │                       │  4. Forward upload       │
   │                       │     (with Backend        │
   │                       │      API key)            │
   │                       ├─────────────────────────>│
   │                       │                          │
   │                       │  5. Process & store      │
   │                       │     (IPFS, ImageKit)     │
   │                       │                          │
   │                       │  6. Return URL + hash    │
   │                       │<─────────────────────────┤
   │  7. Return to FE      │                          │
   │<──────────────────────┤                          │
   │                       │                          │
```

---

## 📋 Prerequisites

- [ ] Phase 3 complete (GraphQL Gateway running)
- [ ] Metadata Service deployed and accessible
- [ ] Backend API key obtained from Metadata Service
- [ ] Environment variables configured

---

## 🗂️ Step 1: Add Environment Variables

**File**: `E:\zuno-marketplace-api\.env`

```env
# ... existing variables ...

# Metadata Service Integration (NEW)
METADATA_SERVICE_URL=http://localhost:3001/api
METADATA_SERVICE_API_KEY=your-backend-api-key-here
```

**File**: `E:\zuno-marketplace-api\.env.example`

```env
# Add to example file
METADATA_SERVICE_URL=http://localhost:3001/api
METADATA_SERVICE_API_KEY=your-backend-api-key
```

---

## 📝 Step 2: Create Metadata Service HTTP Client

**File**: `E:\zuno-marketplace-api\services\graphql-gateway\internal\client\metadata_client.go`

```go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// MetadataServiceClient handles communication with Metadata Service
type MetadataServiceClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewMetadataServiceClient creates a new Metadata Service client
func NewMetadataServiceClient(baseURL, apiKey string) *MetadataServiceClient {
	return &MetadataServiceClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MediaResponse represents response from Metadata Service
type MediaResponse struct {
	ID           string `json:"id"`
	FileName     string `json:"fileName"`
	FileSize     int64  `json:"fileSize"`
	MimeType     string `json:"mimeType"`
	MediaType    string `json:"mediaType"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	IPFSHash     string `json:"ipfsHash,omitempty"`    // Null initially, set after async pinning
	IPFSURL      string `json:"ipfsUrl,omitempty"`     // Null initially, use for baseURI after pinned
	IsPinned     bool   `json:"isPinned"`              // false initially, true after ~10-30s
	PinnedAt     string `json:"pinnedAt,omitempty"`    // Timestamp when pinned to IPFS
	CreatedAt    string `json:"createdAt"`
}

// UploadMedia uploads a single file to Metadata Service
func (c *MetadataServiceClient) UploadMedia(file io.Reader, filename string) (*MediaResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file data
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	// Add optional fields
	writer.WriteField("folder", "collections")
	writer.WriteField("tags", "collection-media")

	// Close writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.baseURL+"/media", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers (both Auth methods work: Bearer or x-api-key)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("x-api-key", c.apiKey)        // Preferred method
	// req.Header.Set("Authorization", "Bearer "+c.apiKey)  // Alternative
	req.Header.Set("x-api-version", "v1")        // REQUIRED!

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response (includes meta field with requestId, timestamp)
	var result struct {
		Success bool          `json:"success"`
		Data    MediaResponse `json:"data"`
		Meta    struct {
			RequestID string `json:"requestId"`
			Timestamp string `json:"timestamp"`
		} `json:"meta,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("metadata service returned success=false")
	}

	return &result.Data, nil
}

// BatchUploadMedia uploads multiple files to Metadata Service
func (c *MetadataServiceClient) BatchUploadMedia(files []io.Reader, filenames []string) ([]*MediaResponse, error) {
	if len(files) != len(filenames) {
		return nil, fmt.Errorf("files and filenames length mismatch")
	}

	if len(files) > 20 {
		return nil, fmt.Errorf("maximum 20 files per batch")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add each file
	for i, file := range files {
		part, err := writer.CreateFormFile("files", filenames[i])
		if err != nil {
			return nil, fmt.Errorf("failed to create form file %d: %w", i, err)
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return nil, fmt.Errorf("failed to copy file %d: %w", i, err)
		}
	}

	// Add folder
	writer.WriteField("folder", "collections")

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.baseURL+"/media/batch", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("x-api-version", "v1")

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("batch upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var result struct {
		Success bool             `json:"success"`
		Data    []*MediaResponse `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}
```

---

## 📝 Step 3: Create Upload Handler

**File**: `E:\zuno-marketplace-api\services\graphql-gateway\internal\handlers\upload_handler.go`

```go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"zuno-marketplace-api/services/graphql-gateway/internal/client"
	"zuno-marketplace-api/services/graphql-gateway/internal/middleware"
)

type UploadHandler struct {
	metadataClient *client.MetadataServiceClient
	logger         *log.Logger
}

func NewUploadHandler(metadataClient *client.MetadataServiceClient, logger *log.Logger) *UploadHandler {
	return &UploadHandler{
		metadataClient: metadataClient,
		logger:         logger,
	}
}

// UploadMedia handles single file uploads
func (h *UploadHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	// Only allow POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Auth required - get user ID from context
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		h.logger.Printf("Upload attempt without authentication")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		h.logger.Printf("Failed to parse multipart form: %v", err)
		http.Error(w, "File too large or invalid request", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Printf("No file provided: %v", err)
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		h.logger.Printf("Invalid file type: %s", contentType)
		http.Error(w, "Invalid file type. Only images allowed (JPEG, PNG, GIF, WebP)", http.StatusBadRequest)
		return
	}

	// Validate file size (max 10MB)
	if header.Size > 10*1024*1024 {
		h.logger.Printf("File too large: %d bytes", header.Size)
		http.Error(w, "File too large. Maximum size is 10MB", http.StatusBadRequest)
		return
	}

	// Log upload attempt
	h.logger.Printf("User %s uploading file: %s (%d bytes, %s)", userID, header.Filename, header.Size, contentType)

	// Forward to Metadata Service
	result, err := h.metadataClient.UploadMedia(file, header.Filename)
	if err != nil {
		h.logger.Printf("Metadata service upload failed: %v", err)
		http.Error(w, "Upload failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	// Log success
	h.logger.Printf("Upload successful: %s -> %s", header.Filename, result.URL)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// BatchUploadMedia handles multiple file uploads
func (h *UploadHandler) BatchUploadMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form (max 50MB for batch)
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		http.Error(w, "Request too large or invalid", http.StatusBadRequest)
		return
	}

	// Get all files
	multipartFiles := r.MultipartForm.File["files"]
	if len(multipartFiles) == 0 {
		http.Error(w, "No files provided", http.StatusBadRequest)
		return
	}

	if len(multipartFiles) > 20 {
		http.Error(w, "Maximum 20 files per batch", http.StatusBadRequest)
		return
	}

	// Open all files
	var files []io.Reader
	var filenames []string

	for _, fileHeader := range multipartFiles {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to open file: "+fileHeader.Filename, http.StatusBadRequest)
			return
		}
		defer file.Close()

		files = append(files, file)
		filenames = append(filenames, fileHeader.Filename)
	}

	h.logger.Printf("User %s uploading %d files", userID, len(files))

	// Forward to Metadata Service
	results, err := h.metadataClient.BatchUploadMedia(files, filenames)
	if err != nil {
		h.logger.Printf("Batch upload failed: %v", err)
		http.Error(w, "Batch upload failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	h.logger.Printf("Batch upload successful: %d files uploaded", len(results))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    results,
	})
}

// isValidImageType checks if content type is valid
func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	contentType = strings.ToLower(strings.TrimSpace(contentType))
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}
```

---

## 📝 Step 4: Register Routes in Gateway

**File**: `E:\zuno-marketplace-api\services\graphql-gateway\cmd\main.go` (update)

```go
package main

import (
	"log"
	"net/http"
	"os"

	"zuno-marketplace-api/services/graphql-gateway/internal/client"
	"zuno-marketplace-api/services/graphql-gateway/internal/handlers"
	"zuno-marketplace-api/services/graphql-gateway/internal/middleware"
	// ... other imports
)

func main() {
	// ... existing code ...

	// Initialize logger
	logger := log.New(os.Stdout, "[GraphQL-Gateway] ", log.LstdFlags)

	// Initialize Metadata Service client (NEW)
	metadataServiceURL := os.Getenv("METADATA_SERVICE_URL")
	metadataAPIKey := os.Getenv("METADATA_SERVICE_API_KEY")

	if metadataServiceURL == "" || metadataAPIKey == "" {
		logger.Fatal("METADATA_SERVICE_URL and METADATA_SERVICE_API_KEY must be set")
	}

	metadataClient := client.NewMetadataServiceClient(metadataServiceURL, metadataAPIKey)
	logger.Printf("Metadata Service client initialized: %s", metadataServiceURL)

	// Initialize upload handler (NEW)
	uploadHandler := handlers.NewUploadHandler(metadataClient, logger)

	// ... existing GraphQL setup ...

	// Create HTTP router
	mux := http.NewServeMux()

	// GraphQL endpoints (existing)
	mux.Handle("/graphql", srv)
	mux.Handle("/playground", playground.Handler("GraphQL", "/graphql"))

	// Upload endpoints (NEW)
	mux.HandleFunc("/api/upload/media", uploadHandler.UploadMedia)
	mux.HandleFunc("/api/upload/batch", uploadHandler.BatchUploadMedia)

	// Health check (existing)
	mux.HandleFunc("/health", healthHandler)

	// Apply middleware
	handler := middleware.AuthMiddleware(mux) // Auth middleware extracts JWT
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.LoggingMiddleware(handler)

	// Start server
	addr := ":" + os.Getenv("GATEWAY_HTTP_ADDR")
	logger.Printf("🚀 GraphQL Gateway running on %s", addr)
	logger.Printf("  - GraphQL API: http://localhost:8081/graphql")
	logger.Printf("  - Playground: http://localhost:8081/playground")
	logger.Printf("  - Upload Media: http://localhost:8081/api/upload/media")
	logger.Printf("  - Batch Upload: http://localhost:8081/api/upload/batch")

	if err := http.ListenAndServe(addr, handler); err != nil {
		logger.Fatal(err)
	}
}
```

---

## ✅ Step 5: Test Upload Proxy

### Test with cURL

```bash
# Get JWT token first (via login)
TOKEN="your-jwt-token-here"

# Test single file upload
curl -X POST http://localhost:8081/api/upload/media \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/path/to/image.png"

# Expected response:
# {
#   "success": true,
#   "data": {
#     "id": "media_abc123",
#     "url": "https://ik.imagekit.io/zuno/image.png",
#     "ipfsHash": "Qm...",
#     "isPinned": false
#   }
# }

# Test without auth (should fail)
curl -X POST http://localhost:8081/api/upload/media \
  -F "file=@/path/to/image.png"

# Expected: 401 Unauthorized

# Test batch upload
curl -X POST http://localhost:8081/api/upload/batch \
  -H "Authorization: Bearer $TOKEN" \
  -F "files=@image1.png" \
  -F "files=@image2.jpg" \
  -F "files=@image3.gif"
```

---

## ✅ Completion Checklist

- [ ] Environment variables configured
- [ ] Metadata Service client implemented
- [ ] Upload handler implemented
- [ ] Routes registered in gateway
- [ ] Auth middleware applied
- [ ] File validation working (size, type)
- [ ] Single upload tested
- [ ] Batch upload tested
- [ ] Error handling tested
- [ ] Logging working
- [ ] Unauthorized access blocked

---

## 🎯 Next Steps

✅ Phase 4 complete! Proceed to:

**[Phase 5: Blockchain Indexer](./05-BLOCKCHAIN-INDEXER.md)**
- Implement webhook endpoints
- Handle blockchain events
- Update collection status

---

**Last Updated**: 2025-11-20
