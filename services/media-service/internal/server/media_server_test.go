package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zunokit/zuno-marketplace-api/services/media-service/internal/client"
	"github.com/zunokit/zuno-marketplace-api/services/media-service/internal/service"
	"github.com/zunokit/zuno-marketplace-api/shared/logger"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setupTestServer(metadataURL string) *MediaServer {
	metadataClient := client.NewMetadataServiceClient(metadataURL, "test-api-key")
	mediaService := service.NewMediaService(metadataClient)
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		ServiceName: "media-service-test",
		Pretty:      false,
	})
	return NewMediaServer(mediaService, log)
}

func TestUploadMedia_Success(t *testing.T) {
	// Create mock metadata service
	metadataServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":        "media_123",
				"fileName":  "test.png",
				"fileSize":  1024,
				"mimeType":  "image/png",
				"mediaType": "image",
				"url":       "https://ik.imagekit.io/test/test.png",
				"isPinned":  false,
				"createdAt": "2025-11-22T10:00:00Z",
			},
		})
	}))
	defer metadataServer.Close()

	server := setupTestServer(metadataServer.URL)

	req := &pb.UploadMediaRequest{
		UserId:      "user_123",
		FileData:    []byte("fake image content"),
		Filename:    "test.png",
		ContentType: "image/png",
		Folder:      "collections",
		Tags:        []string{"test"},
	}

	resp, err := server.UploadMedia(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Media == nil {
		t.Fatal("Expected media in response")
	}

	if resp.Media.Id != "media_123" {
		t.Errorf("Expected ID media_123, got %s", resp.Media.Id)
	}

	if resp.Media.FileName != "test.png" {
		t.Errorf("Expected fileName test.png, got %s", resp.Media.FileName)
	}
}

func TestUploadMedia_MissingUserID(t *testing.T) {
	server := setupTestServer("http://localhost:3001")

	req := &pb.UploadMediaRequest{
		UserId:      "",
		FileData:    []byte("fake content"),
		Filename:    "test.png",
		ContentType: "image/png",
	}

	_, err := server.UploadMedia(context.Background(), req)
	if err == nil {
		t.Error("Expected error when user_id is missing")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected InvalidArgument code, got %v", st.Code())
	}
}

func TestUploadMedia_MissingFileData(t *testing.T) {
	server := setupTestServer("http://localhost:3001")

	req := &pb.UploadMediaRequest{
		UserId:      "user_123",
		FileData:    []byte{},
		Filename:    "test.png",
		ContentType: "image/png",
	}

	_, err := server.UploadMedia(context.Background(), req)
	if err == nil {
		t.Error("Expected error when file_data is missing")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected InvalidArgument code, got %v", st.Code())
	}
}

func TestUploadMedia_InvalidContentType(t *testing.T) {
	server := setupTestServer("http://localhost:3001")

	req := &pb.UploadMediaRequest{
		UserId:      "user_123",
		FileData:    []byte("fake content"),
		Filename:    "test.pdf",
		ContentType: "application/pdf",
	}

	_, err := server.UploadMedia(context.Background(), req)
	if err == nil {
		t.Error("Expected error for invalid content type")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Errorf("Expected Internal code, got %v", st.Code())
	}
}

func TestBatchUploadMedia_Success(t *testing.T) {
	// Create mock metadata service
	metadataServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": []map[string]interface{}{
				{
					"id":        "media_1",
					"fileName":  "test1.png",
					"url":       "https://ik.imagekit.io/test/test1.png",
					"isPinned":  false,
					"createdAt": "2025-11-22T10:00:00Z",
				},
				{
					"id":        "media_2",
					"fileName":  "test2.jpg",
					"url":       "https://ik.imagekit.io/test/test2.jpg",
					"isPinned":  false,
					"createdAt": "2025-11-22T10:00:01Z",
				},
			},
		})
	}))
	defer metadataServer.Close()

	server := setupTestServer(metadataServer.URL)

	req := &pb.BatchUploadMediaRequest{
		UserId: "user_123",
		Files: []*pb.FileData{
			{
				Data:        []byte("fake image 1"),
				Filename:    "test1.png",
				ContentType: "image/png",
			},
			{
				Data:        []byte("fake image 2"),
				Filename:    "test2.jpg",
				ContentType: "image/jpeg",
			},
		},
		Folder: "collections",
	}

	resp, err := server.BatchUploadMedia(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resp.Media) != 2 {
		t.Fatalf("Expected 2 media items, got %d", len(resp.Media))
	}

	if resp.Media[0].Id != "media_1" {
		t.Errorf("Expected first media ID media_1, got %s", resp.Media[0].Id)
	}
}

func TestBatchUploadMedia_MissingUserID(t *testing.T) {
	server := setupTestServer("http://localhost:3001")

	req := &pb.BatchUploadMediaRequest{
		UserId: "",
		Files: []*pb.FileData{
			{
				Data:        []byte("fake content"),
				Filename:    "test.png",
				ContentType: "image/png",
			},
		},
	}

	_, err := server.BatchUploadMedia(context.Background(), req)
	if err == nil {
		t.Error("Expected error when user_id is missing")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected InvalidArgument code, got %v", st.Code())
	}
}

func TestBatchUploadMedia_NoFiles(t *testing.T) {
	server := setupTestServer("http://localhost:3001")

	req := &pb.BatchUploadMediaRequest{
		UserId: "user_123",
		Files:  []*pb.FileData{},
	}

	_, err := server.BatchUploadMedia(context.Background(), req)
	if err == nil {
		t.Error("Expected error when no files provided")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected InvalidArgument code, got %v", st.Code())
	}
}

func TestGetMedia_Unimplemented(t *testing.T) {
	server := setupTestServer("http://localhost:3001")

	req := &pb.GetMediaRequest{
		MediaId: "media_123",
	}

	_, err := server.GetMedia(context.Background(), req)
	if err == nil {
		t.Error("Expected error for unimplemented method")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Unimplemented {
		t.Errorf("Expected Unimplemented code, got %v", st.Code())
	}
}
