package server

import (
	"testing"

	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidateUploadMediaRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      *pb.UploadMediaRequest
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "valid request",
			req: &pb.UploadMediaRequest{
				UserId:      "user_123",
				FileData:    []byte("test data"),
				Filename:    "test.png",
				ContentType: "image/png",
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			req: &pb.UploadMediaRequest{
				UserId:      "",
				FileData:    []byte("test data"),
				Filename:    "test.png",
				ContentType: "image/png",
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "missing file_data",
			req: &pb.UploadMediaRequest{
				UserId:      "user_123",
				FileData:    []byte{},
				Filename:    "test.png",
				ContentType: "image/png",
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "missing filename",
			req: &pb.UploadMediaRequest{
				UserId:      "user_123",
				FileData:    []byte("test data"),
				Filename:    "",
				ContentType: "image/png",
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "missing content_type",
			req: &pb.UploadMediaRequest{
				UserId:      "user_123",
				FileData:    []byte("test data"),
				Filename:    "test.png",
				ContentType: "",
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUploadMediaRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUploadMediaRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				st, ok := status.FromError(err)
				if !ok {
					t.Error("Expected gRPC status error")
					return
				}
				if st.Code() != tt.wantCode {
					t.Errorf("ValidateUploadMediaRequest() code = %v, want %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}

func TestValidateBatchUploadMediaRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      *pb.BatchUploadMediaRequest
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "valid request",
			req: &pb.BatchUploadMediaRequest{
				UserId: "user_123",
				Files: []*pb.FileData{
					{Data: []byte("test"), Filename: "test.png", ContentType: "image/png"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			req: &pb.BatchUploadMediaRequest{
				UserId: "",
				Files: []*pb.FileData{
					{Data: []byte("test"), Filename: "test.png", ContentType: "image/png"},
				},
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "missing files",
			req: &pb.BatchUploadMediaRequest{
				UserId: "user_123",
				Files:  []*pb.FileData{},
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBatchUploadMediaRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBatchUploadMediaRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				st, ok := status.FromError(err)
				if !ok {
					t.Error("Expected gRPC status error")
					return
				}
				if st.Code() != tt.wantCode {
					t.Errorf("ValidateBatchUploadMediaRequest() code = %v, want %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}

func TestValidateGetMediaRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      *pb.GetMediaRequest
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "valid request",
			req: &pb.GetMediaRequest{
				MediaId: "media_123",
			},
			wantErr: false,
		},
		{
			name: "missing media_id",
			req: &pb.GetMediaRequest{
				MediaId: "",
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGetMediaRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGetMediaRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				st, ok := status.FromError(err)
				if !ok {
					t.Error("Expected gRPC status error")
					return
				}
				if st.Code() != tt.wantCode {
					t.Errorf("ValidateGetMediaRequest() code = %v, want %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}
