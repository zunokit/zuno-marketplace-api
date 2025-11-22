package service

import (
	"fmt"
	"strings"

	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
)

// Validation constants
const (
	MaxFileSizeBytes  = 10 * 1024 * 1024 // 10MB
	MaxBatchSizeBytes = 50 * 1024 * 1024 // 50MB
	MaxFilesPerBatch  = 20
)

// ValidMediaTypes contains allowed media content types
var ValidMediaTypes = []string{
	"image/jpeg",
	"image/jpg",
	"image/png",
	"image/gif",
	"image/webp",
}

// IsValidMediaType checks if content type is valid
func IsValidMediaType(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	for _, validType := range ValidMediaTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

// ValidateUploadRequest validates a single file upload request
func ValidateUploadRequest(fileData []byte, filename, contentType string) error {
	// Validate content type
	if !IsValidMediaType(contentType) {
		return fmt.Errorf("invalid content type: %s. Only images allowed (JPEG, PNG, GIF, WebP)", contentType)
	}

	// Validate file size
	if len(fileData) > MaxFileSizeBytes {
		return fmt.Errorf("file too large: %d bytes. Maximum size is %dMB", len(fileData), MaxFileSizeBytes/(1024*1024))
	}

	// Validate filename
	if filename == "" {
		return fmt.Errorf("filename is required")
	}

	return nil
}

// ValidateBatchUploadRequest validates a batch upload request
func ValidateBatchUploadRequest(files []*pb.FileData) error {
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}

	if len(files) > MaxFilesPerBatch {
		return fmt.Errorf("maximum %d files per batch", MaxFilesPerBatch)
	}

	var totalSize int64
	for i, f := range files {
		// Validate content type
		if !IsValidMediaType(f.ContentType) {
			return fmt.Errorf("file %d: invalid content type: %s", i, f.ContentType)
		}

		// Track total size
		totalSize += int64(len(f.Data))
		if totalSize > MaxBatchSizeBytes {
			return fmt.Errorf("total upload size exceeds %dMB limit", MaxBatchSizeBytes/(1024*1024))
		}
	}

	return nil
}
