package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/quangdang46/NFT-Marketplace/services/graphql-gateway/internal/middleware"
	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc"
)

// MockMediaServiceClient is a mock implementation of MediaServiceClient
type MockMediaServiceClient struct {
	UploadMediaFunc      func(ctx context.Context, req *pb.UploadMediaRequest, opts ...grpc.CallOption) (*pb.UploadMediaResponse, error)
	BatchUploadMediaFunc func(ctx context.Context, req *pb.BatchUploadMediaRequest, opts ...grpc.CallOption) (*pb.BatchUploadMediaResponse, error)
	GetMediaFunc         func(ctx context.Context, req *pb.GetMediaRequest, opts ...grpc.CallOption) (*pb.GetMediaResponse, error)
}

func (m *MockMediaServiceClient) UploadMedia(ctx context.Context, req *pb.UploadMediaRequest, opts ...grpc.CallOption) (*pb.UploadMediaResponse, error) {
	if m.UploadMediaFunc != nil {
		return m.UploadMediaFunc(ctx, req, opts...)
	}
	return &pb.UploadMediaResponse{
		Media: &pb.MediaInfo{
			Id:        "media_123",
			FileName:  req.Filename,
			Url:       "https://ik.imagekit.io/test/" + req.Filename,
			IsPinned:  false,
			CreatedAt: "2025-11-22T10:00:00Z",
		},
	}, nil
}

func (m *MockMediaServiceClient) BatchUploadMedia(ctx context.Context, req *pb.BatchUploadMediaRequest, opts ...grpc.CallOption) (*pb.BatchUploadMediaResponse, error) {
	if m.BatchUploadMediaFunc != nil {
		return m.BatchUploadMediaFunc(ctx, req, opts...)
	}
	var media []*pb.MediaInfo
	for i, f := range req.Files {
		media = append(media, &pb.MediaInfo{
			Id:        "media_" + string(rune('1'+i)),
			FileName:  f.Filename,
			Url:       "https://ik.imagekit.io/test/" + f.Filename,
			IsPinned:  false,
			CreatedAt: "2025-11-22T10:00:00Z",
		})
	}
	return &pb.BatchUploadMediaResponse{Media: media}, nil
}

func (m *MockMediaServiceClient) GetMedia(ctx context.Context, req *pb.GetMediaRequest, opts ...grpc.CallOption) (*pb.GetMediaResponse, error) {
	if m.GetMediaFunc != nil {
		return m.GetMediaFunc(ctx, req, opts...)
	}
	return nil, nil
}

func TestUploadMedia_Success(t *testing.T) {
	mockClient := &MockMediaServiceClient{}
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	handler := NewUploadHandler(mockClient, logger)

	// Create multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	h := make(map[string][]string)
	h["Content-Disposition"] = []string{`form-data; name="file"; filename="test.png"`}
	h["Content-Type"] = []string{"image/png"}
	part, _ := writer.CreatePart(h)
	part.Write([]byte("fake image content"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload/media", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(req.Context(), middleware.UserClaimsKey, &middleware.UserClaims{
		UserID: "user_123",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UploadMedia(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var response map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&response)
	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}
}

func TestUploadMedia_Unauthorized(t *testing.T) {
	mockClient := &MockMediaServiceClient{}
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	handler := NewUploadHandler(mockClient, logger)

	req := httptest.NewRequest(http.MethodPost, "/api/upload/media", nil)
	rr := httptest.NewRecorder()
	handler.UploadMedia(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestUploadMedia_MethodNotAllowed(t *testing.T) {
	mockClient := &MockMediaServiceClient{}
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	handler := NewUploadHandler(mockClient, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/upload/media", nil)
	ctx := context.WithValue(req.Context(), middleware.UserClaimsKey, &middleware.UserClaims{
		UserID: "user_123",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UploadMedia(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestBatchUploadMedia_Success(t *testing.T) {
	mockClient := &MockMediaServiceClient{}
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	handler := NewUploadHandler(mockClient, logger)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part1, _ := writer.CreateFormFile("files", "test1.png")
	part1.Write([]byte("fake image 1"))
	part2, _ := writer.CreateFormFile("files", "test2.jpg")
	part2.Write([]byte("fake image 2"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload/batch", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(req.Context(), middleware.UserClaimsKey, &middleware.UserClaims{
		UserID: "user_123",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.BatchUploadMedia(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}
}

func TestBatchUploadMedia_TooManyFiles(t *testing.T) {
	mockClient := &MockMediaServiceClient{}
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	handler := NewUploadHandler(mockClient, logger)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for i := 0; i < 21; i++ {
		part, _ := writer.CreateFormFile("files", "test.png")
		part.Write([]byte("fake image"))
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload/batch", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(req.Context(), middleware.UserClaimsKey, &middleware.UserClaims{
		UserID: "user_123",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.BatchUploadMedia(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "20") {
		t.Error("Expected error message about max 20 files")
	}
}
