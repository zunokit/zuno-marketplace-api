package client

import (
	"strings"
	"testing"
)

func TestValidateBatchUploadParams(t *testing.T) {
	tests := []struct {
		name        string
		files       [][]byte
		filenames   []string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid params",
			files:     [][]byte{[]byte("test1"), []byte("test2")},
			filenames: []string{"file1.png", "file2.png"},
			wantErr:   false,
		},
		{
			name:        "length mismatch - more files",
			files:       [][]byte{[]byte("test1"), []byte("test2")},
			filenames:   []string{"file1.png"},
			wantErr:     true,
			errContains: "mismatch",
		},
		{
			name:        "length mismatch - more filenames",
			files:       [][]byte{[]byte("test1")},
			filenames:   []string{"file1.png", "file2.png"},
			wantErr:     true,
			errContains: "mismatch",
		},
		{
			name:        "too many files",
			files:       make([][]byte, MaxFilesPerBatch+1),
			filenames:   make([]string, MaxFilesPerBatch+1),
			wantErr:     true,
			errContains: "maximum",
		},
		{
			name:      "exactly max files",
			files:     make([][]byte, MaxFilesPerBatch),
			filenames: make([]string, MaxFilesPerBatch),
			wantErr:   false,
		},
		{
			name:      "empty files is valid at this level",
			files:     [][]byte{},
			filenames: []string{},
			wantErr:   false, // Empty check is done at service level
		},
	}

	// Initialize files for max tests
	for i := range tests {
		if tests[i].name == "too many files" || tests[i].name == "exactly max files" {
			for j := range tests[i].files {
				tests[i].files[j] = []byte("test")
				tests[i].filenames[j] = "test.png"
			}
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBatchUploadParams(tt.files, tt.filenames)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBatchUploadParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ValidateBatchUploadParams() error = %v, want error containing %q", err, tt.errContains)
			}
		})
	}
}

func TestMaxFilesPerBatchConstant(t *testing.T) {
	if MaxFilesPerBatch != 20 {
		t.Errorf("MaxFilesPerBatch = %d, want 20", MaxFilesPerBatch)
	}
}
