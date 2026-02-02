package service

import (
	"testing"

	"github.com/zunokit/zuno-marketplace-api/services/media-service/internal/client"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
)

func TestMediaResponseToProto(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		result := MediaResponseToProto(nil)
		if result != nil {
			t.Error("Expected nil for nil input")
		}
	})

	t.Run("full conversion", func(t *testing.T) {
		input := &client.MediaResponse{
			ID:           "media_123",
			FileName:     "test.png",
			FileSize:     1024,
			MimeType:     "image/png",
			MediaType:    "image",
			URL:          "https://example.com/test.png",
			ThumbnailURL: "https://example.com/test_thumb.png",
			Width:        800,
			Height:       600,
			IPFSHash:     "Qm123",
			IPFSURL:      "ipfs://Qm123",
			IsPinned:     true,
			PinnedAt:     "2025-01-01T00:00:00Z",
			CreatedAt:    "2025-01-01T00:00:00Z",
		}

		result := MediaResponseToProto(input)

		if result.Id != input.ID {
			t.Errorf("Id = %v, want %v", result.Id, input.ID)
		}
		if result.FileName != input.FileName {
			t.Errorf("FileName = %v, want %v", result.FileName, input.FileName)
		}
		if result.FileSize != input.FileSize {
			t.Errorf("FileSize = %v, want %v", result.FileSize, input.FileSize)
		}
		if result.MimeType != input.MimeType {
			t.Errorf("MimeType = %v, want %v", result.MimeType, input.MimeType)
		}
		if result.Url != input.URL {
			t.Errorf("Url = %v, want %v", result.Url, input.URL)
		}
		if result.Width != int32(input.Width) {
			t.Errorf("Width = %v, want %v", result.Width, input.Width)
		}
		if result.Height != int32(input.Height) {
			t.Errorf("Height = %v, want %v", result.Height, input.Height)
		}
		if result.IsPinned != input.IsPinned {
			t.Errorf("IsPinned = %v, want %v", result.IsPinned, input.IsPinned)
		}
	})
}

func TestMediaResponsesToProto(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		result := MediaResponsesToProto(nil)
		if result != nil {
			t.Error("Expected nil for nil input")
		}
	})

	t.Run("multiple items", func(t *testing.T) {
		input := []*client.MediaResponse{
			{ID: "media_1", FileName: "test1.png"},
			{ID: "media_2", FileName: "test2.png"},
		}

		result := MediaResponsesToProto(input)

		if len(result) != 2 {
			t.Fatalf("Expected 2 items, got %d", len(result))
		}
		if result[0].Id != "media_1" {
			t.Errorf("First item Id = %v, want media_1", result[0].Id)
		}
		if result[1].Id != "media_2" {
			t.Errorf("Second item Id = %v, want media_2", result[1].Id)
		}
	})
}

func TestExtractFileData(t *testing.T) {
	files := []*pb.FileData{
		{Data: []byte("data1"), Filename: "file1.png"},
		{Data: []byte("data2"), Filename: "file2.jpg"},
	}

	fileData, filenames := ExtractFileData(files)

	if len(fileData) != 2 {
		t.Fatalf("Expected 2 file data, got %d", len(fileData))
	}
	if len(filenames) != 2 {
		t.Fatalf("Expected 2 filenames, got %d", len(filenames))
	}

	if string(fileData[0]) != "data1" {
		t.Errorf("First file data = %v, want data1", string(fileData[0]))
	}
	if filenames[0] != "file1.png" {
		t.Errorf("First filename = %v, want file1.png", filenames[0])
	}
	if string(fileData[1]) != "data2" {
		t.Errorf("Second file data = %v, want data2", string(fileData[1]))
	}
	if filenames[1] != "file2.jpg" {
		t.Errorf("Second filename = %v, want file2.jpg", filenames[1])
	}
}
