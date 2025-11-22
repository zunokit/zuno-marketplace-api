package service

import (
	"strings"
	"testing"

	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
)

func TestIsValidMediaType(t *testing.T) {
	tests := []struct {
		contentType string
		expected    bool
	}{
		{"image/jpeg", true},
		{"image/jpg", true},
		{"image/png", true},
		{"image/gif", true},
		{"image/webp", true},
		{"IMAGE/PNG", true},     // case insensitive
		{"  image/png  ", true}, // with whitespace
		{"image/svg+xml", false},
		{"application/pdf", false},
		{"text/plain", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			result := IsValidMediaType(tt.contentType)
			if result != tt.expected {
				t.Errorf("IsValidMediaType(%q) = %v, want %v", tt.contentType, result, tt.expected)
			}
		})
	}
}

func TestValidateUploadRequest(t *testing.T) {
	tests := []struct {
		name        string
		fileData    []byte
		filename    string
		contentType string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid request",
			fileData:    []byte("test data"),
			filename:    "test.png",
			contentType: "image/png",
			wantErr:     false,
		},
		{
			name:        "invalid content type",
			fileData:    []byte("test data"),
			filename:    "test.pdf",
			contentType: "application/pdf",
			wantErr:     true,
			errContains: "invalid content type",
		},
		{
			name:        "empty filename",
			fileData:    []byte("test data"),
			filename:    "",
			contentType: "image/png",
			wantErr:     true,
			errContains: "filename is required",
		},
		{
			name:        "file too large",
			fileData:    make([]byte, MaxFileSizeBytes+1),
			filename:    "test.png",
			contentType: "image/png",
			wantErr:     true,
			errContains: "file too large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUploadRequest(tt.fileData, tt.filename, tt.contentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUploadRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ValidateUploadRequest() error = %v, want error containing %q", err, tt.errContains)
			}
		})
	}
}

func TestValidateBatchUploadRequest(t *testing.T) {
	tests := []struct {
		name        string
		files       []*pb.FileData
		wantErr     bool
		errContains string
	}{
		{
			name: "valid batch",
			files: []*pb.FileData{
				{Data: []byte("test1"), Filename: "test1.png", ContentType: "image/png"},
				{Data: []byte("test2"), Filename: "test2.jpg", ContentType: "image/jpeg"},
			},
			wantErr: false,
		},
		{
			name:        "empty files",
			files:       []*pb.FileData{},
			wantErr:     true,
			errContains: "no files provided",
		},
		{
			name:        "too many files",
			files:       make([]*pb.FileData, MaxFilesPerBatch+1),
			wantErr:     true,
			errContains: "maximum",
		},
		{
			name: "invalid content type in batch",
			files: []*pb.FileData{
				{Data: []byte("test"), Filename: "test.pdf", ContentType: "application/pdf"},
			},
			wantErr:     true,
			errContains: "invalid content type",
		},
	}

	// Initialize too many files test case
	for i := range tests {
		if tests[i].name == "too many files" {
			for j := 0; j <= MaxFilesPerBatch; j++ {
				tests[i].files[j] = &pb.FileData{
					Data:        []byte("test"),
					Filename:    "test.png",
					ContentType: "image/png",
				}
			}
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBatchUploadRequest(tt.files)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBatchUploadRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ValidateBatchUploadRequest() error = %v, want error containing %q", err, tt.errContains)
			}
		})
	}
}
