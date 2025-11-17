package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/quangdang46/NFT-Marketplace/services/auth-service/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create the login_events table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS login_events (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			account_id TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			result TEXT NOT NULL,
			error_message TEXT,
			chain_id TEXT,
			domain TEXT,
			timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Create indexes
	db.Exec("CREATE INDEX idx_login_events_user_id ON login_events(user_id)")
	db.Exec("CREATE INDEX idx_login_events_account_id ON login_events(account_id)")
	db.Exec("CREATE INDEX idx_login_events_timestamp ON login_events(timestamp)")
	db.Exec("CREATE INDEX idx_login_events_result ON login_events(result)")
	db.Exec("CREATE INDEX idx_login_events_ip_address ON login_events(ip_address)")

	return db
}

func TestLoginEventRepository_CreateLoginEvent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLoginEventRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"
	chainID := "eip155:1"
	domain := "localhost"

	event := &models.LoginEvent{
		ID:        uuid.New(), // SQLite doesn't auto-generate UUIDs
		UserID:    &userID,
		AccountID: "0x1234567890123456789012345678901234567890",
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
		Result:    models.LoginResultSuccess,
		ChainID:   &chainID,
		Domain:    &domain,
		Timestamp: time.Now(),
	}

	err := repo.CreateLoginEvent(ctx, event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if event.ID == uuid.Nil {
		t.Error("Expected ID to be set")
	}
}

func TestLoginEventRepository_CreateLoginEvent_Failed(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLoginEventRepository(db)
	ctx := context.Background()

	ipAddress := "192.168.1.1"
	errorMsg := "signature verification failed"
	chainID := "eip155:1"
	domain := "localhost"

	event := &models.LoginEvent{
		ID:           uuid.New(), // SQLite doesn't auto-generate UUIDs
		AccountID:    "0x1234567890123456789012345678901234567890",
		IPAddress:    &ipAddress,
		Result:       models.LoginResultInvalidSignature,
		ErrorMessage: &errorMsg,
		ChainID:      &chainID,
		Domain:       &domain,
		Timestamp:    time.Now(),
	}

	err := repo.CreateLoginEvent(ctx, event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if event.ID == uuid.Nil {
		t.Error("Expected ID to be set")
	}
}

func TestLoginEventRepository_GetLoginEventsByUserID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLoginEventRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	ipAddress := "192.168.1.1"

	// Create multiple login events
	for i := 0; i < 5; i++ {
		event := &models.LoginEvent{
			ID:        uuid.New(),
			UserID:    &userID,
			AccountID: "0x1234567890123456789012345678901234567890",
			IPAddress: &ipAddress,
			Result:    models.LoginResultSuccess,
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
		err := repo.CreateLoginEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}

	// Retrieve events
	events, err := repo.GetLoginEventsByUserID(ctx, userID, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(events) != 5 {
		t.Errorf("Expected 5 events, got %d", len(events))
	}

	// Verify events are ordered by timestamp DESC
	for i := 0; i < len(events)-1; i++ {
		if events[i].Timestamp.Before(events[i+1].Timestamp) {
			t.Error("Expected events to be ordered by timestamp DESC")
		}
	}
}

func TestLoginEventRepository_GetLoginEventsByAccountID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLoginEventRepository(db)
	ctx := context.Background()

	accountID := "0x1234567890123456789012345678901234567890"
	ipAddress := "192.168.1.1"

	// Create events
	for i := 0; i < 3; i++ {
		event := &models.LoginEvent{
			ID:        uuid.New(),
			AccountID: accountID,
			IPAddress: &ipAddress,
			Result:    models.LoginResultSuccess,
			Timestamp: time.Now(),
		}
		err := repo.CreateLoginEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}

	// Retrieve events
	events, err := repo.GetLoginEventsByAccountID(ctx, accountID, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}

	// Verify all events have the same account ID
	for _, event := range events {
		if event.AccountID != accountID {
			t.Errorf("Expected account ID %s, got %s", accountID, event.AccountID)
		}
	}
}

func TestLoginEventRepository_GetLoginEventsByIPAddress(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLoginEventRepository(db)
	ctx := context.Background()

	ipAddress := "192.168.1.100"

	// Create events
	for i := 0; i < 2; i++ {
		event := &models.LoginEvent{
			ID:        uuid.New(),
			AccountID: "0x1234567890123456789012345678901234567890",
			IPAddress: &ipAddress,
			Result:    models.LoginResultSuccess,
			Timestamp: time.Now(),
		}
		err := repo.CreateLoginEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}

	// Retrieve events
	events, err := repo.GetLoginEventsByIPAddress(ctx, ipAddress, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}
}

func TestLoginEventRepository_GetRecentFailedAttempts(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLoginEventRepository(db)
	ctx := context.Background()

	accountID := "0x1234567890123456789012345678901234567890"
	ipAddress := "192.168.1.1"

	// Create successful login
	successEvent := &models.LoginEvent{
		ID:        uuid.New(),
		AccountID: accountID,
		IPAddress: &ipAddress,
		Result:    models.LoginResultSuccess,
		Timestamp: time.Now(),
	}
	repo.CreateLoginEvent(ctx, successEvent)

	// Create failed logins
	for i := 0; i < 3; i++ {
		errorMsg := "signature verification failed"
		event := &models.LoginEvent{
			ID:           uuid.New(),
			AccountID:    accountID,
			IPAddress:    &ipAddress,
			Result:       models.LoginResultInvalidSignature,
			ErrorMessage: &errorMsg,
			Timestamp:    time.Now().Add(time.Duration(i) * time.Second),
		}
		err := repo.CreateLoginEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}

	// Retrieve failed attempts from the last 5 minutes
	since := time.Now().Add(-5 * time.Minute)
	events, err := repo.GetRecentFailedAttempts(ctx, accountID, since)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should only get the 3 failed attempts, not the successful one
	if len(events) != 3 {
		t.Errorf("Expected 3 failed attempts, got %d", len(events))
	}

	// Verify all are failures
	for _, event := range events {
		if event.Result == models.LoginResultSuccess {
			t.Error("Expected only failed login attempts")
		}
	}
}

func TestLoginEventRepository_CountFailedAttempts(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLoginEventRepository(db)
	ctx := context.Background()

	accountID := "0x1234567890123456789012345678901234567890"
	ipAddress := "192.168.1.1"

	// Create failed logins
	for i := 0; i < 5; i++ {
		errorMsg := "invalid nonce"
		event := &models.LoginEvent{
			ID:           uuid.New(),
			AccountID:    accountID,
			IPAddress:    &ipAddress,
			Result:       models.LoginResultInvalidNonce,
			ErrorMessage: &errorMsg,
			Timestamp:    time.Now().Add(time.Duration(-i) * time.Minute),
		}
		err := repo.CreateLoginEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}

	// Count failed attempts in the last 10 minutes
	since := time.Now().Add(-10 * time.Minute)
	count, err := repo.CountFailedAttempts(ctx, accountID, &ipAddress, since)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if count != 5 {
		t.Errorf("Expected 5 failed attempts, got %d", count)
	}

	// Count with only account ID (no IP filter)
	count, err = repo.CountFailedAttempts(ctx, accountID, nil, since)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if count != 5 {
		t.Errorf("Expected 5 failed attempts, got %d", count)
	}
}

func TestLoginEventRepository_CountFailedAttempts_TimeWindow(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLoginEventRepository(db)
	ctx := context.Background()

	accountID := "0x1234567890123456789012345678901234567890"
	ipAddress := "192.168.1.1"

	// Create old failed login (outside time window)
	oldEvent := &models.LoginEvent{
		ID:        uuid.New(),
		AccountID: accountID,
		IPAddress: &ipAddress,
		Result:    models.LoginResultInvalidSignature,
		Timestamp: time.Now().Add(-20 * time.Minute),
	}
	repo.CreateLoginEvent(ctx, oldEvent)

	// Create recent failed login (inside time window)
	recentEvent := &models.LoginEvent{
		ID:        uuid.New(),
		AccountID: accountID,
		IPAddress: &ipAddress,
		Result:    models.LoginResultInvalidSignature,
		Timestamp: time.Now().Add(-2 * time.Minute),
	}
	repo.CreateLoginEvent(ctx, recentEvent)

	// Count only recent attempts (last 5 minutes)
	since := time.Now().Add(-5 * time.Minute)
	count, err := repo.CountFailedAttempts(ctx, accountID, &ipAddress, since)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should only count the recent one, not the old one
	if count != 1 {
		t.Errorf("Expected 1 failed attempt in time window, got %d", count)
	}
}
