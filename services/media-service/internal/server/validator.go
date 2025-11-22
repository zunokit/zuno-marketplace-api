package server

import (
	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidateUploadMediaRequest validates the UploadMediaRequest
func ValidateUploadMediaRequest(req *pb.UploadMediaRequest) error {
	if req.UserId == "" {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}

	if len(req.FileData) == 0 {
		return status.Error(codes.InvalidArgument, "file_data is required")
	}

	if req.Filename == "" {
		return status.Error(codes.InvalidArgument, "filename is required")
	}

	if req.ContentType == "" {
		return status.Error(codes.InvalidArgument, "content_type is required")
	}

	return nil
}

// ValidateBatchUploadMediaRequest validates the BatchUploadMediaRequest
func ValidateBatchUploadMediaRequest(req *pb.BatchUploadMediaRequest) error {
	if req.UserId == "" {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}

	if len(req.Files) == 0 {
		return status.Error(codes.InvalidArgument, "files are required")
	}

	return nil
}

// ValidateGetMediaRequest validates the GetMediaRequest
func ValidateGetMediaRequest(req *pb.GetMediaRequest) error {
	if req.MediaId == "" {
		return status.Error(codes.InvalidArgument, "media_id is required")
	}

	return nil
}
