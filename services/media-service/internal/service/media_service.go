package service

import (
	"fmt"

	"github.com/zunokit/zuno-marketplace-api/services/media-service/internal/client"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
)

// MediaService handles media upload business logic
type MediaService struct {
	metadataClient *client.MetadataServiceClient
}

// NewMediaService creates a new MediaService
func NewMediaService(metadataClient *client.MetadataServiceClient) *MediaService {
	return &MediaService{
		metadataClient: metadataClient,
	}
}

// UploadMedia uploads a single file
func (s *MediaService) UploadMedia(userID string, fileData []byte, filename, contentType, folder string, tags []string) (*pb.MediaInfo, error) {
	// Validate request
	if err := ValidateUploadRequest(fileData, filename, contentType); err != nil {
		return nil, err
	}

	// Upload to metadata service
	result, err := s.metadataClient.UploadMedia(fileData, filename, contentType, folder, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to upload media: %w", err)
	}

	// Convert to proto response
	return MediaResponseToProto(result), nil
}

// BatchUploadMedia uploads multiple files
func (s *MediaService) BatchUploadMedia(userID string, files []*pb.FileData, folder string) ([]*pb.MediaInfo, error) {
	// Validate request
	if err := ValidateBatchUploadRequest(files); err != nil {
		return nil, err
	}

	// Extract file data
	fileData, filenames := ExtractFileData(files)

	// Upload to metadata service
	results, err := s.metadataClient.BatchUploadMedia(fileData, filenames, folder)
	if err != nil {
		return nil, fmt.Errorf("failed to batch upload media: %w", err)
	}

	// Convert to proto response
	return MediaResponsesToProto(results), nil
}
