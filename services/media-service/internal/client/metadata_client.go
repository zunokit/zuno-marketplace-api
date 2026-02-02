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
	IPFSHash     string `json:"ipfsHash,omitempty"`
	IPFSURL      string `json:"ipfsUrl,omitempty"`
	IsPinned     bool   `json:"isPinned"`
	PinnedAt     string `json:"pinnedAt,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

// UploadMedia uploads a single file to Metadata Service
func (c *MetadataServiceClient) UploadMedia(fileData []byte, filename, contentType, folder string, tags []string) (*MediaResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file data
	_, err = io.Copy(part, bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	// Add optional fields
	if folder != "" {
		writer.WriteField("folder", folder)
	} else {
		writer.WriteField("folder", "uploads")
	}

	for _, tag := range tags {
		writer.WriteField("tags", tag)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.baseURL+"/media", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("x-api-version", "v1")

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

	// Parse response
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
func (c *MetadataServiceClient) BatchUploadMedia(files [][]byte, filenames []string, folder string) ([]*MediaResponse, error) {
	// Validate parameters
	if err := ValidateBatchUploadParams(files, filenames); err != nil {
		return nil, err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add each file
	for i, fileData := range files {
		part, err := writer.CreateFormFile("files", filenames[i])
		if err != nil {
			return nil, fmt.Errorf("failed to create form file %d: %w", i, err)
		}

		_, err = io.Copy(part, bytes.NewReader(fileData))
		if err != nil {
			return nil, fmt.Errorf("failed to copy file %d: %w", i, err)
		}
	}

	// Add folder
	if folder != "" {
		writer.WriteField("folder", folder)
	} else {
		writer.WriteField("folder", "uploads")
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.baseURL+"/media/batch", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("x-api-key", c.apiKey)
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
