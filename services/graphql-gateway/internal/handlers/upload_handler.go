package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/zunokit/zuno-marketplace-api/services/graphql-gateway/internal/middleware"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
)

type UploadHandler struct {
	mediaClient pb.MediaServiceClient
	logger      *log.Logger
}

func NewUploadHandler(mediaClient pb.MediaServiceClient, logger *log.Logger) *UploadHandler {
	return &UploadHandler{
		mediaClient: mediaClient,
		logger:      logger,
	}
}

// UploadMedia handles single file uploads
func (h *UploadHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	// Only allow POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Auth required - get user ID from context
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		h.logger.Printf("Upload attempt without authentication")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		h.logger.Printf("Failed to parse multipart form: %v", err)
		http.Error(w, "File too large or invalid request", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Printf("No file provided: %v", err)
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		h.logger.Printf("Failed to read file: %v", err)
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}

	// Get content type
	contentType := header.Header.Get("Content-Type")

	// Get optional folder from form
	folder := r.FormValue("folder")
	if folder == "" {
		folder = "uploads"
	}

	// Log upload attempt
	h.logger.Printf("User %s uploading file: %s (%d bytes, %s)", userID, header.Filename, len(fileData), contentType)

	// Call Media Service via gRPC
	resp, err := h.mediaClient.UploadMedia(context.Background(), &pb.UploadMediaRequest{
		UserId:      userID,
		FileData:    fileData,
		Filename:    header.Filename,
		ContentType: contentType,
		Folder:      folder,
	})
	if err != nil {
		h.logger.Printf("Media service upload failed: %v", err)
		http.Error(w, "Upload failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	// Log success
	h.logger.Printf("Upload successful: %s -> %s", header.Filename, resp.Media.Url)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    resp.Media,
	})
}

// BatchUploadMedia handles multiple file uploads
func (h *UploadHandler) BatchUploadMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form (max 50MB for batch)
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		http.Error(w, "Request too large or invalid", http.StatusBadRequest)
		return
	}

	// Get all files
	multipartFiles := r.MultipartForm.File["files"]
	if len(multipartFiles) == 0 {
		http.Error(w, "No files provided", http.StatusBadRequest)
		return
	}

	if len(multipartFiles) > 20 {
		http.Error(w, "Maximum 20 files per batch", http.StatusBadRequest)
		return
	}

	// Read all files
	var files []*pb.FileData
	for _, fileHeader := range multipartFiles {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to open file: "+fileHeader.Filename, http.StatusBadRequest)
			return
		}

		fileData, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			http.Error(w, "Failed to read file: "+fileHeader.Filename, http.StatusBadRequest)
			return
		}

		files = append(files, &pb.FileData{
			Data:        fileData,
			Filename:    fileHeader.Filename,
			ContentType: fileHeader.Header.Get("Content-Type"),
		})
	}

	// Get optional folder from form
	folder := r.FormValue("folder")
	if folder == "" {
		folder = "uploads"
	}

	h.logger.Printf("User %s uploading %d files", userID, len(files))

	// Call Media Service via gRPC
	resp, err := h.mediaClient.BatchUploadMedia(context.Background(), &pb.BatchUploadMediaRequest{
		UserId: userID,
		Files:  files,
		Folder: folder,
	})
	if err != nil {
		h.logger.Printf("Batch upload failed: %v", err)
		http.Error(w, "Batch upload failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	h.logger.Printf("Batch upload successful: %d files uploaded", len(resp.Media))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    resp.Media,
	})
}
