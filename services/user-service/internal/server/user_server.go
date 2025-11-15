package server

import (
	"context"

	"github.com/google/uuid"
	"github.com/quangdang46/NFT-Marketplace/services/user-service/internal/models"
	"github.com/quangdang46/NFT-Marketplace/services/user-service/internal/repository"
	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserServer implements the gRPC UserService
type UserServer struct {
	pb.UnimplementedUserServiceServer
	userRepo repository.UserRepository
}

// NewUserServer creates a new user service gRPC server
func NewUserServer(userRepo repository.UserRepository) *UserServer {
	return &UserServer{
		userRepo: userRepo,
	}
}

// EnsureUser creates user if not exists, otherwise returns existing (idempotent)
func (s *UserServer) EnsureUser(ctx context.Context, req *pb.EnsureUserRequest) (*pb.EnsureUserResponse, error) {
	// Validate request
	if req.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	// Generate deterministic user ID from account_id
	// Using UUID v5 (SHA-1 hash) for deterministic ID generation
	namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8") // DNS namespace
	userID := uuid.NewSHA1(namespace, []byte(req.AccountId))

	// Try to ensure user exists
	user, err := s.userRepo.EnsureUser(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to ensure user: %v", err)
	}

	// Check if this was a new creation or existing user
	created := user.CreatedAt.After(user.UpdatedAt.Add(-1 * 1000000000)) // Within 1 second = new

	return &pb.EnsureUserResponse{
		UserId:  user.UserID.String(),
		Created: created,
	}, nil
}

// GetUser retrieves user by ID with profile
func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	// Validate request
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	// Get user with profile
	user, err := s.userRepo.GetByID(ctx, userID, true)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	// Convert to proto
	pbUser := &pb.User{
		Id:        user.UserID.String(),
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	var pbProfile *pb.Profile
	if user.Profile != nil {
		pbProfile = &pb.Profile{
			UserId:    user.Profile.UserID.String(),
			Locale:    user.Profile.Locale,
			Timezone:  user.Profile.Timezone,
			UpdatedAt: user.Profile.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if user.Profile.Username != nil {
			pbProfile.Username = *user.Profile.Username
		}
		if user.Profile.DisplayName != nil {
			pbProfile.DisplayName = *user.Profile.DisplayName
		}
		if user.Profile.AvatarURL != nil {
			pbProfile.AvatarUrl = *user.Profile.AvatarURL
		}
		if user.Profile.BannerURL != nil {
			pbProfile.BannerUrl = *user.Profile.BannerURL
		}
		if user.Profile.Bio != nil {
			pbProfile.Bio = *user.Profile.Bio
		}
		if user.Profile.SocialsJSON != nil {
			pbProfile.SocialsJson = *user.Profile.SocialsJSON
		}
	}

	return &pb.GetUserResponse{
		User:    pbUser,
		Profile: pbProfile,
	}, nil
}

// UpsertProfile updates user profile
func (s *UserServer) UpsertProfile(ctx context.Context, req *pb.UpsertProfileRequest) (*pb.UpsertProfileResponse, error) {
	// Validate request
	if req.Profile == nil || req.Profile.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "profile with user_id is required")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.Profile.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	// Convert proto to model
	profile := &models.Profile{
		UserID:   userID,
		Locale:   req.Profile.Locale,
		Timezone: req.Profile.Timezone,
	}

	if req.Profile.Username != "" {
		profile.Username = &req.Profile.Username
	}
	if req.Profile.DisplayName != "" {
		profile.DisplayName = &req.Profile.DisplayName
	}
	if req.Profile.AvatarUrl != "" {
		profile.AvatarURL = &req.Profile.AvatarUrl
	}
	if req.Profile.BannerUrl != "" {
		profile.BannerURL = &req.Profile.BannerUrl
	}
	if req.Profile.Bio != "" {
		profile.Bio = &req.Profile.Bio
	}
	if req.Profile.SocialsJson != "" {
		profile.SocialsJSON = &req.Profile.SocialsJson
	}

	// Update profile
	if err := s.userRepo.UpdateProfile(ctx, profile); err != nil {
		if err == repository.ErrUsernameTaken {
			return nil, status.Error(codes.AlreadyExists, "username already taken")
		}
		if err == repository.ErrUserNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update profile: %v", err)
	}

	// Return updated profile
	return &pb.UpsertProfileResponse{
		Profile: req.Profile,
	}, nil
}
