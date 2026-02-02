package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zunokit/zuno-marketplace-api/services/user-service/internal/models"
	"github.com/zunokit/zuno-marketplace-api/services/user-service/internal/repository"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mock UserRepository for testing
type mockUserRepository struct {
	createUserFunc    func(ctx context.Context, user *models.User) error
	getByIDFunc       func(ctx context.Context, userID uuid.UUID, withProfile bool) (*models.User, error)
	getByUsernameFunc func(ctx context.Context, username string) (*models.User, error)
	updateProfileFunc func(ctx context.Context, profile *models.Profile) error
	updateStatusFunc  func(ctx context.Context, userID uuid.UUID, status models.UserStatus) error
	ensureUserFunc    func(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

func (m *mockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, userID uuid.UUID, withProfile bool) (*models.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, userID, withProfile)
	}
	return nil, nil
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if m.getByUsernameFunc != nil {
		return m.getByUsernameFunc(ctx, username)
	}
	return nil, nil
}

func (m *mockUserRepository) UpdateProfile(ctx context.Context, profile *models.Profile) error {
	if m.updateProfileFunc != nil {
		return m.updateProfileFunc(ctx, profile)
	}
	return nil
}

func (m *mockUserRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status models.UserStatus) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, userID, status)
	}
	return nil
}

func (m *mockUserRepository) EnsureUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	if m.ensureUserFunc != nil {
		return m.ensureUserFunc(ctx, userID)
	}
	return nil, nil
}

func TestUserServer_EnsureUser_CreateNew(t *testing.T) {
	accountID := "eip155:1:0x1234567890123456789012345678901234567890"
	now := time.Now()

	mockRepo := &mockUserRepository{
		ensureUserFunc: func(ctx context.Context, userID uuid.UUID) (*models.User, error) {
			// Simulate new user creation
			return &models.User{
				UserID:    userID,
				Status:    models.UserStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	server := NewUserServer(mockRepo)

	req := &pb.EnsureUserRequest{
		AccountId: accountID,
	}

	resp, err := server.EnsureUser(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.UserId == "" {
		t.Error("Expected user ID to be set")
	}

	if !resp.Created {
		t.Error("Expected Created to be true for new user")
	}
}

func TestUserServer_EnsureUser_ReturnExisting(t *testing.T) {
	accountID := "eip155:1:0x1234567890123456789012345678901234567890"
	now := time.Now()
	oldTime := now.Add(-1 * time.Hour)

	mockRepo := &mockUserRepository{
		ensureUserFunc: func(ctx context.Context, userID uuid.UUID) (*models.User, error) {
			// Simulate existing user (CreatedAt is much older than UpdatedAt)
			return &models.User{
				UserID:    userID,
				Status:    models.UserStatusActive,
				CreatedAt: oldTime,
				UpdatedAt: now,
			}, nil
		},
	}

	server := NewUserServer(mockRepo)

	req := &pb.EnsureUserRequest{
		AccountId: accountID,
	}

	resp, err := server.EnsureUser(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.UserId == "" {
		t.Error("Expected user ID to be set")
	}

	if resp.Created {
		t.Error("Expected Created to be false for existing user")
	}
}

func TestUserServer_EnsureUser_ValidationError(t *testing.T) {
	mockRepo := &mockUserRepository{}
	server := NewUserServer(mockRepo)

	req := &pb.EnsureUserRequest{
		AccountId: "", // Missing account ID
	}

	_, err := server.EnsureUser(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for missing account_id")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected InvalidArgument error code, got %v", st.Code())
	}
}

func TestUserServer_EnsureUser_DeterministicID(t *testing.T) {
	accountID := "eip155:1:0x1234567890123456789012345678901234567890"
	now := time.Now()

	var capturedUserID uuid.UUID

	mockRepo := &mockUserRepository{
		ensureUserFunc: func(ctx context.Context, userID uuid.UUID) (*models.User, error) {
			capturedUserID = userID
			return &models.User{
				UserID:    userID,
				Status:    models.UserStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	server := NewUserServer(mockRepo)

	req := &pb.EnsureUserRequest{
		AccountId: accountID,
	}

	// Call multiple times with same account ID
	resp1, _ := server.EnsureUser(context.Background(), req)
	userID1 := capturedUserID

	resp2, _ := server.EnsureUser(context.Background(), req)
	userID2 := capturedUserID

	// Should generate the same UUID for the same account ID
	if resp1.UserId != resp2.UserId {
		t.Error("Expected same user ID for same account ID")
	}

	if userID1 != userID2 {
		t.Error("Expected deterministic UUID generation")
	}
}

func TestUserServer_GetUser_Success(t *testing.T) {
	userID := uuid.New()
	username := "johndoe"
	displayName := "John Doe"
	bio := "Hello World"
	now := time.Now()

	mockRepo := &mockUserRepository{
		getByIDFunc: func(ctx context.Context, uid uuid.UUID, withProfile bool) (*models.User, error) {
			if uid != userID {
				return nil, repository.ErrUserNotFound
			}

			user := &models.User{
				UserID:    userID,
				Status:    models.UserStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}

			if withProfile {
				user.Profile = &models.Profile{
					UserID:      userID,
					Username:    &username,
					DisplayName: &displayName,
					Bio:         &bio,
					Locale:      "en",
					Timezone:    "UTC",
					UpdatedAt:   now,
				}
			}

			return user, nil
		},
	}

	server := NewUserServer(mockRepo)

	req := &pb.GetUserRequest{
		UserId: userID.String(),
	}

	resp, err := server.GetUser(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.User == nil {
		t.Fatal("Expected user in response")
	}

	if resp.User.Id != userID.String() {
		t.Errorf("Expected user ID %s, got %s", userID.String(), resp.User.Id)
	}

	if resp.Profile == nil {
		t.Fatal("Expected profile in response")
	}

	if resp.Profile.Username != username {
		t.Errorf("Expected username %s, got %s", username, resp.Profile.Username)
	}
}

func TestUserServer_GetUser_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		req     *pb.GetUserRequest
		wantErr codes.Code
	}{
		{
			name:    "missing user_id",
			req:     &pb.GetUserRequest{},
			wantErr: codes.InvalidArgument,
		},
		{
			name: "invalid user_id format",
			req: &pb.GetUserRequest{
				UserId: "invalid-uuid",
			},
			wantErr: codes.InvalidArgument,
		},
	}

	mockRepo := &mockUserRepository{}
	server := NewUserServer(mockRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := server.GetUser(context.Background(), tt.req)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatal("Expected gRPC status error")
			}

			if st.Code() != tt.wantErr {
				t.Errorf("Expected error code %v, got %v", tt.wantErr, st.Code())
			}
		})
	}
}

func TestUserServer_GetUser_NotFound(t *testing.T) {
	mockRepo := &mockUserRepository{
		getByIDFunc: func(ctx context.Context, userID uuid.UUID, withProfile bool) (*models.User, error) {
			return nil, repository.ErrUserNotFound
		},
	}

	server := NewUserServer(mockRepo)

	req := &pb.GetUserRequest{
		UserId: uuid.New().String(),
	}

	_, err := server.GetUser(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.NotFound {
		t.Errorf("Expected NotFound error code, got %v", st.Code())
	}
}

func TestUserServer_UpsertProfile_Success(t *testing.T) {
	userID := uuid.New()
	username := "johndoe"
	displayName := "John Doe"

	mockRepo := &mockUserRepository{
		updateProfileFunc: func(ctx context.Context, profile *models.Profile) error {
			// Verify fields are set correctly
			if profile.UserID != userID {
				t.Errorf("Expected user ID %s, got %s", userID, profile.UserID)
			}
			if profile.Username == nil || *profile.Username != username {
				t.Errorf("Expected username %s, got %v", username, profile.Username)
			}
			if profile.DisplayName == nil || *profile.DisplayName != displayName {
				t.Errorf("Expected display name %s, got %v", displayName, profile.DisplayName)
			}
			return nil
		},
	}

	server := NewUserServer(mockRepo)

	req := &pb.UpsertProfileRequest{
		Profile: &pb.Profile{
			UserId:      userID.String(),
			Username:    username,
			DisplayName: displayName,
			Locale:      "en",
			Timezone:    "UTC",
		},
	}

	resp, err := server.UpsertProfile(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Profile == nil {
		t.Fatal("Expected profile in response")
	}

	if resp.Profile.Username != username {
		t.Errorf("Expected username %s, got %s", username, resp.Profile.Username)
	}
}

func TestUserServer_UpsertProfile_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		req     *pb.UpsertProfileRequest
		wantErr codes.Code
	}{
		{
			name:    "missing profile",
			req:     &pb.UpsertProfileRequest{},
			wantErr: codes.InvalidArgument,
		},
		{
			name: "missing user_id",
			req: &pb.UpsertProfileRequest{
				Profile: &pb.Profile{},
			},
			wantErr: codes.InvalidArgument,
		},
		{
			name: "invalid user_id format",
			req: &pb.UpsertProfileRequest{
				Profile: &pb.Profile{
					UserId: "invalid-uuid",
				},
			},
			wantErr: codes.InvalidArgument,
		},
	}

	mockRepo := &mockUserRepository{}
	server := NewUserServer(mockRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := server.UpsertProfile(context.Background(), tt.req)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatal("Expected gRPC status error")
			}

			if st.Code() != tt.wantErr {
				t.Errorf("Expected error code %v, got %v", tt.wantErr, st.Code())
			}
		})
	}
}

func TestUserServer_UpsertProfile_UsernameTaken(t *testing.T) {
	mockRepo := &mockUserRepository{
		updateProfileFunc: func(ctx context.Context, profile *models.Profile) error {
			return repository.ErrUsernameTaken
		},
	}

	server := NewUserServer(mockRepo)

	req := &pb.UpsertProfileRequest{
		Profile: &pb.Profile{
			UserId:   uuid.New().String(),
			Username: "takenusername",
			Locale:   "en",
			Timezone: "UTC",
		},
	}

	_, err := server.UpsertProfile(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for taken username")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.AlreadyExists {
		t.Errorf("Expected AlreadyExists error code, got %v", st.Code())
	}
}

func TestUserServer_UpsertProfile_UserNotFound(t *testing.T) {
	mockRepo := &mockUserRepository{
		updateProfileFunc: func(ctx context.Context, profile *models.Profile) error {
			return repository.ErrUserNotFound
		},
	}

	server := NewUserServer(mockRepo)

	req := &pb.UpsertProfileRequest{
		Profile: &pb.Profile{
			UserId:   uuid.New().String(),
			Username: "newusername",
			Locale:   "en",
			Timezone: "UTC",
		},
	}

	_, err := server.UpsertProfile(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.NotFound {
		t.Errorf("Expected NotFound error code, got %v", st.Code())
	}
}

func TestUserServer_UpsertProfile_OptionalFields(t *testing.T) {
	userID := uuid.New()
	username := "johndoe"
	avatarURL := "https://example.com/avatar.jpg"
	bio := "Developer"

	mockRepo := &mockUserRepository{
		updateProfileFunc: func(ctx context.Context, profile *models.Profile) error {
			// Verify optional fields
			if profile.AvatarURL != nil && *profile.AvatarURL != avatarURL {
				t.Errorf("Expected avatar URL %s, got %v", avatarURL, *profile.AvatarURL)
			}
			if profile.Bio != nil && *profile.Bio != bio {
				t.Errorf("Expected bio %s, got %v", bio, *profile.Bio)
			}
			return nil
		},
	}

	server := NewUserServer(mockRepo)

	req := &pb.UpsertProfileRequest{
		Profile: &pb.Profile{
			UserId:    userID.String(),
			Username:  username,
			AvatarUrl: avatarURL,
			Bio:       bio,
			Locale:    "en",
			Timezone:  "UTC",
		},
	}

	_, err := server.UpsertProfile(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestUserServer_UpsertProfile_EmptyOptionalFields(t *testing.T) {
	userID := uuid.New()
	username := "johndoe"

	mockRepo := &mockUserRepository{
		updateProfileFunc: func(ctx context.Context, profile *models.Profile) error {
			// Verify empty strings don't set optional fields
			if profile.DisplayName != nil {
				t.Error("Expected DisplayName to be nil for empty string")
			}
			if profile.AvatarURL != nil {
				t.Error("Expected AvatarURL to be nil for empty string")
			}
			if profile.Bio != nil {
				t.Error("Expected Bio to be nil for empty string")
			}
			return nil
		},
	}

	server := NewUserServer(mockRepo)

	req := &pb.UpsertProfileRequest{
		Profile: &pb.Profile{
			UserId:      userID.String(),
			Username:    username,
			DisplayName: "", // Empty - should not be set
			AvatarUrl:   "", // Empty - should not be set
			Bio:         "", // Empty - should not be set
			Locale:      "en",
			Timezone:    "UTC",
		},
	}

	_, err := server.UpsertProfile(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestUserServer_EnsureUser_RepositoryError(t *testing.T) {
	mockRepo := &mockUserRepository{
		ensureUserFunc: func(ctx context.Context, userID uuid.UUID) (*models.User, error) {
			return nil, errors.New("database error")
		},
	}

	server := NewUserServer(mockRepo)

	req := &pb.EnsureUserRequest{
		AccountId: "eip155:1:0x1234567890123456789012345678901234567890",
	}

	_, err := server.EnsureUser(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.Internal {
		t.Errorf("Expected Internal error code, got %v", st.Code())
	}
}
