package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/quangdang46/NFT-Marketplace/services/user-service/internal/models"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUsernameTaken = errors.New("username already taken")
	ErrInvalidStatus = errors.New("invalid user status")
)

// UserRepository handles user persistence operations
type UserRepository interface {
	// CreateUser creates a new user with profile
	CreateUser(ctx context.Context, user *models.User) error

	// GetByID retrieves a user by ID with optional profile preload
	GetByID(ctx context.Context, userID uuid.UUID, withProfile bool) (*models.User, error)

	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*models.User, error)

	// UpdateProfile updates user profile information
	UpdateProfile(ctx context.Context, profile *models.Profile) error

	// UpdateStatus updates user status
	UpdateStatus(ctx context.Context, userID uuid.UUID, status models.UserStatus) error

	// EnsureUser creates user if not exists, otherwise returns existing (idempotent)
	EnsureUser(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser creates a new user with default profile
// The database trigger will auto-create profile, preferences, and stats
func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID retrieves a user by ID with optional profile preload
func (r *userRepository) GetByID(ctx context.Context, userID uuid.UUID, withProfile bool) (*models.User, error) {
	var user models.User
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if withProfile {
		query = query.Preload("Profile")
	}

	err := query.First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Joins("JOIN profiles ON profiles.user_id = users.user_id").
		Where("profiles.username = ?", username).
		Preload("Profile").
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// UpdateProfile updates user profile information
func (r *userRepository) UpdateProfile(ctx context.Context, profile *models.Profile) error {
	result := r.db.WithContext(ctx).
		Model(&models.Profile{}).
		Where("user_id = ?", profile.UserID).
		Updates(profile)

	if result.Error != nil {
		// Check for unique constraint violation on username
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrUsernameTaken
		}
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateStatus updates user status
func (r *userRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status models.UserStatus) error {
	// Validate status
	validStatuses := map[models.UserStatus]bool{
		models.UserStatusActive:    true,
		models.UserStatusBanned:    true,
		models.UserStatusDeleted:   true,
		models.UserStatusSuspended: true,
	}
	if !validStatuses[status] {
		return ErrInvalidStatus
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("user_id = ?", userID).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// EnsureUser creates user if not exists, otherwise returns existing (idempotent)
// This is used during SIWE authentication to create user on first login
func (r *userRepository) EnsureUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	// Try to get existing user first
	user, err := r.GetByID(ctx, userID, true)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	// User doesn't exist, create new one
	user = &models.User{
		UserID: userID,
		Status: models.UserStatusActive,
	}

	if err := r.CreateUser(ctx, user); err != nil {
		// Handle race condition - another request might have created the user
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return r.GetByID(ctx, userID, true)
		}
		return nil, err
	}

	// Reload with profile (created by trigger)
	return r.GetByID(ctx, userID, true)
}
