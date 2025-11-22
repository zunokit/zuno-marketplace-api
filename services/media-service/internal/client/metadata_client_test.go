package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewMetadataServiceClient(t *testing.T) {
	baseURL := "http://localhost:3001/api"
	apiKey := "test-api-key"

	client := NewMetadataServiceClient(baseURL, apiKey)

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.baseURL != baseURL {
		t.Errorf("Expected baseURL %s, got %s", baseURL, client.baseURL)
	}

	if client.apiKey != apiKey {
		t.Errorf("Expected apiKey %s, got %s", apiKey, client.apiKey)
	}

	if client.client == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}

func TestUploadMedia_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify headers
		if r.Header.Get("x-api-key") == "" {
			t.Error("Expected x-api-key header to be set")
		}

		if r.Header.Get("x-api-version") != "v1" {
			t.Error("Expected x-api-version header to be v1")
		}

		// Parse multipart form
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			t.Fatalf("Failed to parse multipart form: %v", err)
		}

		// Verify file field exists
		_, _, err = r.FormFile("file")
		if err != nil {
			t.Fatalf("Failed to get file from form: %v", err)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":        "media_123",
				"fileName":  "test.png",
				"fileSize":  1024,
				"mimeType":  "image/png",
				"mediaType": "image",
				"url":       "https://ik.imagekit.io/test/test.png",
				"isPinned":  false,
				"createdAt": "2025-11-22T10:00:00Z",
			},
		})
	}))
	defer server.Close()

	client := NewMetadataServiceClient(server.URL, "test-api-key")

	fileContent := []byte("test file content")
	result, err := client.UploadMedia(fileContent, "test.png", "image/png", "collections", []string{"test"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be returned")
	}

	if result.ID != "media_123" {
		t.Errorf("Expected ID media_123, got %s", result.ID)
	}

	if result.FileName != "test.png" {
		t.Errorf("Expected fileName test.png, got %s", result.FileName)
	}

	if result.URL != "https://ik.imagekit.io/test/test.png" {
		t.Errorf("Expected URL https://ik.imagekit.io/test/test.png, got %s", result.URL)
	}
}

func TestUploadMedia_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": "Invalid file"}`))
	}))
	defer server.Close()

	client := NewMetadataServiceClient(server.URL, "test-api-key")

	fileContent := []byte("test file content")
	_, err := client.UploadMedia(fileContent, "test.png", "image/png", "", nil)

	if err == nil {
		t.Error("Expected error when upload fails")
	}

	if !strings.Contains(err.Error(), "400") {
		t.Errorf("Expected error to contain status code 400, got: %v", err)
	}
}

func TestUploadMedia_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
		})
	}))
	defer server.Close()

	client := NewMetadataServiceClient(server.URL, "test-api-key")

	fileContent := []byte("test file content")
	_, err := client.UploadMedia(fileContent, "test.png", "image/png", "", nil)

	if err == nil {
		t.Error("Expected error when success=false")
	}

	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("Expected error to mention success=false, got: %v", err)
	}
}

func TestBatchUploadMedia_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "/batch") {
			t.Errorf("Expected /batch path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": []map[string]interface{}{
				{
					"id":        "media_1",
					"fileName":  "test1.png",
					"url":       "https://ik.imagekit.io/test/test1.png",
					"isPinned":  false,
					"createdAt": "2025-11-22T10:00:00Z",
				},
				{
					"id":        "media_2",
					"fileName":  "test2.jpg",
					"url":       "https://ik.imagekit.io/test/test2.jpg",
					"isPinned":  false,
					"createdAt": "2025-11-22T10:00:01Z",
				},
			},
		})
	}))
	defer server.Close()

	client := NewMetadataServiceClient(server.URL, "test-api-key")

	files := [][]byte{
		[]byte("file 1 content"),
		[]byte("file 2 content"),
	}
	filenames := []string{"test1.png", "test2.jpg"}

	results, err := client.BatchUploadMedia(files, filenames, "collections")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	if results[0].ID != "media_1" {
		t.Errorf("Expected first result ID media_1, got %s", results[0].ID)
	}

	if results[1].ID != "media_2" {
		t.Errorf("Expected second result ID media_2, got %s", results[1].ID)
	}
}

func TestBatchUploadMedia_LengthMismatch(t *testing.T) {
	client := NewMetadataServiceClient("http://localhost:3001", "test-key")

	files := [][]byte{[]byte("test")}
	filenames := []string{"test1.png", "test2.png"}

	_, err := client.BatchUploadMedia(files, filenames, "")

	if err == nil {
		t.Error("Expected error when files and filenames length mismatch")
	}

	if !strings.Contains(err.Error(), "mismatch") {
		t.Errorf("Expected error to mention mismatch, got: %v", err)
	}
}

func TestBatchUploadMedia_TooManyFiles(t *testing.T) {
	client := NewMetadataServiceClient("http://localhost:3001", "test-key")

	// Create 21 files (exceeds max of 20)
	files := make([][]byte, 21)
	filenames := make([]string, 21)
	for i := 0; i < 21; i++ {
		files[i] = []byte("test")
		filenames[i] = "test.png"
	}

	_, err := client.BatchUploadMedia(files, filenames, "")

	if err == nil {
		t.Error("Expected error when uploading more than 20 files")
	}

	if !strings.Contains(err.Error(), "20") {
		t.Errorf("Expected error to mention max 20 files, got: %v", err)
	}
}

func TestBatchUploadMedia_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": "Server error"}`))
	}))
	defer server.Close()

	client := NewMetadataServiceClient(server.URL, "test-api-key")

	files := [][]byte{[]byte("test")}
	filenames := []string{"test.png"}

	_, err := client.BatchUploadMedia(files, filenames, "")

	if err == nil {
		t.Error("Expected error when batch upload fails")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error to contain status code 500, got: %v", err)
	}
}

func TestUploadMedia_DefaultFolder(t *testing.T) {
	var capturedFolder string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20)
		capturedFolder = r.FormValue("folder")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":        "media_123",
				"fileName":  "test.png",
				"url":       "https://example.com/test.png",
				"isPinned":  false,
				"createdAt": "2025-11-22T10:00:00Z",
			},
		})
	}))
	defer server.Close()

	client := NewMetadataServiceClient(server.URL, "test-api-key")

	// Test with empty folder - should use default
	_, err := client.UploadMedia([]byte("test"), "test.png", "image/png", "", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if capturedFolder != "uploads" {
		t.Errorf("Expected default folder 'uploads', got '%s'", capturedFolder)
	}
}

func TestMediaResponse_AllFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":           "media_full",
				"fileName":     "full.png",
				"fileSize":     2048,
				"mimeType":     "image/png",
				"mediaType":    "image",
				"url":          "https://ik.imagekit.io/test/full.png",
				"thumbnailUrl": "https://ik.imagekit.io/test/full_thumb.png",
				"width":        1920,
				"height":       1080,
				"ipfsHash":     "QmTestHash",
				"ipfsUrl":      "ipfs://QmTestHash",
				"isPinned":     true,
				"pinnedAt":     "2025-11-22T12:00:00Z",
				"createdAt":    "2025-11-22T10:00:00Z",
			},
		})
	}))
	defer server.Close()

	client := NewMetadataServiceClient(server.URL, "test-api-key")

	result, err := client.UploadMedia([]byte("test"), "full.png", "image/png", "test", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify all fields
	if result.ID != "media_full" {
		t.Errorf("ID = %v, want media_full", result.ID)
	}
	if result.ThumbnailURL != "https://ik.imagekit.io/test/full_thumb.png" {
		t.Errorf("ThumbnailURL = %v, want https://ik.imagekit.io/test/full_thumb.png", result.ThumbnailURL)
	}
	if result.Width != 1920 {
		t.Errorf("Width = %v, want 1920", result.Width)
	}
	if result.Height != 1080 {
		t.Errorf("Height = %v, want 1080", result.Height)
	}
	if result.IPFSHash != "QmTestHash" {
		t.Errorf("IPFSHash = %v, want QmTestHash", result.IPFSHash)
	}
	if !result.IsPinned {
		t.Error("IsPinned should be true")
	}
}

func TestUploadMedia_NetworkError(t *testing.T) {
	// Use invalid URL to simulate network error
	client := NewMetadataServiceClient("http://localhost:99999", "test-api-key")

	_, err := client.UploadMedia([]byte("test"), "test.png", "image/png", "", nil)

	if err == nil {
		t.Error("Expected error for network failure")
	}
}
