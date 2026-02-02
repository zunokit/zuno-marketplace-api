package client

import "fmt"

// Validation constants
const (
	MaxFilesPerBatch = 20
)

// ValidateBatchUploadParams validates batch upload parameters
func ValidateBatchUploadParams(files [][]byte, filenames []string) error {
	if len(files) != len(filenames) {
		return fmt.Errorf("files and filenames length mismatch")
	}

	if len(files) > MaxFilesPerBatch {
		return fmt.Errorf("maximum %d files per batch", MaxFilesPerBatch)
	}

	return nil
}
