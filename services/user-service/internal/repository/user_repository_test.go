package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zunokit/zuno-marketplace-api/services/user-service/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create the users table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create the profiles table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS profiles (
			user_id TEXT PRIMARY KEY,
			username TEXT,
			display_name TEXT,
			avatar_url TEXT,
			banner_url TEXT,
			bio TEXT,
			locale TEXT DEFAULT 'en',
			timezone TEXT DEFAULT 'UTC',
			socials_json TEXT,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create profiles table: %v", err)
	}

	// Create indexes
	db.Exec("CREATE INDEX idx_users_status ON users(status)")
	db.Exec("CREATE INDEX idx_users_created_at ON users(created_at)")
	db.Exec("CREATE UNIQUE INDEX uq_profiles_username ON profiles(username) WHERE username IS NOT NULL")
	db.Exec("CREATE INDEX idx_profiles_updated_at ON profiles(updated_at)")

	return db
}

func TestUserRepository_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	user := &models.User{
		UserID: userID,
		Status: models.UserStatusActive,
	}

	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify user was created
	retrieved, err := repo.GetByID(ctx, userID, false)
	if err != nil {
		t.Fatalf("Expected to retrieve user, got error: %v", err)
	}

	if retrieved.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, retrieved.UserID)
	}

	if retrieved.Status != models.UserStatusActive {
		t.Errorf("Expected status active, got %s", retrieved.Status)
	}
}

func TestUserRepository_GetByID_WithProfile(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	user := &models.User{
		UserID: userID,
		Status: models.UserStatusActive,
	}

	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create profile manually (in real DB, trigger does this)
	username := "testuser"
	profile := &models.Profile{
		UserID:   userID,
		Username: &username,
		Locale:   "en",
		Timezone: "UTC",
	}
	db.Create(profile)

	// Retrieve with profile
	retrieved, err := repo.GetByID(ctx, userID, true)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.Profile == nil {
		t.Fatal("Expected profile to be loaded")
	}

	if retrieved.Profile.Username == nil || *retrieved.Profile.Username != username {
		t.Errorf("Expected username %s, got %v", username, retrieved.Profile.Username)
	}
}

func TestUserRepository_GetByID_WithoutProfile(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	user := &models.User{
		UserID: userID,
		Status: models.UserStatusActive,
	}

	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Retrieve without profile
	retrieved, err := repo.GetByID(ctx, userID, false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.Profile != nil {
		t.Error("Expected profile to not be loaded")
	}
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Try to get non-existent user
	_, err := repo.GetByID(ctx, uuid.New(), false)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	username := "johndoe"

	user := &models.User{
		UserID: userID,
		Status: models.UserStatusActive,
	}

	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create profile
	profile := &models.Profile{
		UserID:   userID,
		Username: &username,
		Locale:   "en",
		Timezone: "UTC",
	}
	db.Create(profile)

	// Retrieve by username
	retrieved, err := repo.GetByUsername(ctx, username)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, retrieved.UserID)
	}

	if retrieved.Profile == nil {
		t.Fatal("Expected profile to be loaded")
	}

	if retrieved.Profile.Username == nil || *retrieved.Profile.Username != username {
		t.Errorf("Expected username %s, got %v", username, retrieved.Profile.Username)
	}
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Try to get non-existent username
	_, err := repo.GetByUsername(ctx, "nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_UpdateProfile(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	user := &models.User{
		UserID: userID,
		Status: models.UserStatusActive,
	}

	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create initial profile
	username := "oldusername"
	profile := &models.Profile{
		UserID:   userID,
		Username: &username,
		Locale:   "en",
		Timezone: "UTC",
	}
	db.Create(profile)

	// Update profile
	newUsername := "newusername"
	displayName := "John Doe"
	bio := "Hello World"

	updatedProfile := &models.Profile{
		UserID:      userID,
		Username:    &newUsername,
		DisplayName: &displayName,
		Bio:         &bio,
		Locale:      "es",
		Timezone:    "America/New_York",
	}

	err = repo.UpdateProfile(ctx, updatedProfile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify update
	retrieved, err := repo.GetByID(ctx, userID, true)
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}

	if retrieved.Profile.Username == nil || *retrieved.Profile.Username != newUsername {
		t.Errorf("Expected username %s, got %v", newUsername, retrieved.Profile.Username)
	}

	if retrieved.Profile.DisplayName == nil || *retrieved.Profile.DisplayName != displayName {
		t.Errorf("Expected display name %s, got %v", displayName, retrieved.Profile.DisplayName)
	}

	if retrieved.Profile.Bio == nil || *retrieved.Profile.Bio != bio {
		t.Errorf("Expected bio %s, got %v", bio, retrieved.Profile.Bio)
	}
}

func TestUserRepository_UpdateProfile_UsernameTaken(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create first user with username
	userID1 := uuid.New()
	user1 := &models.User{
		UserID: userID1,
		Status: models.UserStatusActive,
	}
	repo.CreateUser(ctx, user1)

	username1 := "takenusername"
	profile1 := &models.Profile{
		UserID:   userID1,
		Username: &username1,
		Locale:   "en",
		Timezone: "UTC",
	}
	db.Create(profile1)

	// Create second user
	userID2 := uuid.New()
	user2 := &models.User{
		UserID: userID2,
		Status: models.UserStatusActive,
	}
	repo.CreateUser(ctx, user2)

	username2 := "otherusername"
	profile2 := &models.Profile{
		UserID:   userID2,
		Username: &username2,
		Locale:   "en",
		Timezone: "UTC",
	}
	db.Create(profile2)

	// Try to update second user with first user's username
	updatedProfile := &models.Profile{
		UserID:   userID2,
		Username: &username1, // Taken username
		Locale:   "en",
		Timezone: "UTC",
	}

	err := repo.UpdateProfile(ctx, updatedProfile)
	if err != ErrUsernameTaken {
		t.Errorf("Expected ErrUsernameTaken, got %v", err)
	}
}

func TestUserRepository_UpdateProfile_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Try to update non-existent user's profile
	username := "someusername"
	profile := &models.Profile{
		UserID:   uuid.New(),
		Username: &username,
		Locale:   "en",
		Timezone: "UTC",
	}

	err := repo.UpdateProfile(ctx, profile)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	user := &models.User{
		UserID: userID,
		Status: models.UserStatusActive,
	}

	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Update status to banned
	err = repo.UpdateStatus(ctx, userID, models.UserStatusBanned)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify update
	retrieved, err := repo.GetByID(ctx, userID, false)
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}

	if retrieved.Status != models.UserStatusBanned {
		t.Errorf("Expected status banned, got %s", retrieved.Status)
	}
}

func TestUserRepository_UpdateStatus_InvalidStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	user := &models.User{
		UserID: userID,
		Status: models.UserStatusActive,
	}

	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Try to update with invalid status
	err = repo.UpdateStatus(ctx, userID, models.UserStatus("invalid"))
	if err != ErrInvalidStatus {
		t.Errorf("Expected ErrInvalidStatus, got %v", err)
	}
}

func TestUserRepository_UpdateStatus_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Try to update non-existent user
	err := repo.UpdateStatus(ctx, uuid.New(), models.UserStatusBanned)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_EnsureUser_CreateNew(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Ensure user (should create new)
	user, err := repo.EnsureUser(ctx, userID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, user.UserID)
	}

	if user.Status != models.UserStatusActive {
		t.Errorf("Expected status active, got %s", user.Status)
	}

	// Verify user exists in database
	retrieved, err := repo.GetByID(ctx, userID, false)
	if err != nil {
		t.Fatalf("Expected to retrieve user, got error: %v", err)
	}

	if retrieved.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, retrieved.UserID)
	}
}

func TestUserRepository_EnsureUser_ReturnExisting(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Create user first
	user := &models.User{
		UserID: userID,
		Status: models.UserStatusActive,
	}
	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Wait a moment to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Ensure user (should return existing)
	ensured, err := repo.EnsureUser(ctx, userID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ensured.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, ensured.UserID)
	}

	// CreatedAt should be older than UpdatedAt or very close
	// (not a new creation)
	if ensured.CreatedAt.After(ensured.UpdatedAt) {
		t.Error("Expected CreatedAt to be before or equal to UpdatedAt for existing user")
	}
}

func TestUserRepository_EnsureUser_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Call EnsureUser multiple times
	user1, err := repo.EnsureUser(ctx, userID)
	if err != nil {
		t.Fatalf("First EnsureUser failed: %v", err)
	}

	user2, err := repo.EnsureUser(ctx, userID)
	if err != nil {
		t.Fatalf("Second EnsureUser failed: %v", err)
	}

	user3, err := repo.EnsureUser(ctx, userID)
	if err != nil {
		t.Fatalf("Third EnsureUser failed: %v", err)
	}

	// All should return the same user ID
	if user1.UserID != userID || user2.UserID != userID || user3.UserID != userID {
		t.Error("Expected all EnsureUser calls to return the same user ID")
	}

	// Verify only one user exists
	var count int64
	db.Model(&models.User{}).Where("user_id = ?", userID).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 user in database, got %d", count)
	}
}

func TestUserRepository_AllValidStatuses(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	validStatuses := []models.UserStatus{
		models.UserStatusActive,
		models.UserStatusBanned,
		models.UserStatusSuspended,
		models.UserStatusDeleted,
	}

	for _, status := range validStatuses {
		userID := uuid.New()
		user := &models.User{
			UserID: userID,
			Status: models.UserStatusActive,
		}

		err := repo.CreateUser(ctx, user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		err = repo.UpdateStatus(ctx, userID, status)
		if err != nil {
			t.Errorf("Expected status %s to be valid, got error: %v", status, err)
		}

		// Verify status was updated
		retrieved, _ := repo.GetByID(ctx, userID, false)
		if retrieved.Status != status {
			t.Errorf("Expected status %s, got %s", status, retrieved.Status)
		}
	}
}
