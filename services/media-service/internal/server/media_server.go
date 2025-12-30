package server

import (
	"context"

	"github.com/zunokit/zuno-marketplace-api/services/media-service/internal/service"
	"github.com/zunokit/zuno-marketplace-api/shared/logger"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MediaServer implements the gRPC MediaService
type MediaServer struct {
	pb.UnimplementedMediaServiceServer
	mediaService *service.MediaService
	logger       *logger.Logger
}

// NewMediaServer creates a new MediaServer
func NewMediaServer(mediaService *service.MediaService, log *logger.Logger) *MediaServer {
	return &MediaServer{
		mediaService: mediaService,
		logger:       log,
	}
}

// UploadMedia handles single file upload
func (s *MediaServer) UploadMedia(ctx context.Context, req *pb.UploadMediaRequest) (*pb.UploadMediaResponse, error) {
	// Validate request
	if err := ValidateUploadMediaRequest(req); err != nil {
		return nil, err
	}

	s.logger.Infof("User %s uploading file: %s (%d bytes, %s)", req.UserId, req.Filename, len(req.FileData), req.ContentType)

	// Upload media
	mediaInfo, err := s.mediaService.UploadMedia(
		req.UserId,
		req.FileData,
		req.Filename,
		req.ContentType,
		req.Folder,
		req.Tags,
	)
	if err != nil {
		s.logger.Errorf("Upload failed for user %s: %v", req.UserId, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.logger.Infof("Upload successful: %s -> %s", req.Filename, mediaInfo.Url)

	return &pb.UploadMediaResponse{
		Media: mediaInfo,
	}, nil
}

// BatchUploadMedia handles multiple file upload
func (s *MediaServer) BatchUploadMedia(ctx context.Context, req *pb.BatchUploadMediaRequest) (*pb.BatchUploadMediaResponse, error) {
	// Validate request
	if err := ValidateBatchUploadMediaRequest(req); err != nil {
		return nil, err
	}

	s.logger.Infof("User %s batch uploading %d files", req.UserId, len(req.Files))

	// Upload media
	mediaInfos, err := s.mediaService.BatchUploadMedia(
		req.UserId,
		req.Files,
		req.Folder,
	)
	if err != nil {
		s.logger.Errorf("Batch upload failed for user %s: %v", req.UserId, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.logger.Infof("Batch upload successful: %d files uploaded", len(mediaInfos))

	return &pb.BatchUploadMediaResponse{
		Media: mediaInfos,
	}, nil
}

// GetMedia retrieves media info by ID
func (s *MediaServer) GetMedia(ctx context.Context, req *pb.GetMediaRequest) (*pb.GetMediaResponse, error) {
	// TODO: Implement when needed (requires Metadata Service API support)
	return nil, status.Error(codes.Unimplemented, "GetMedia not yet implemented")
}
