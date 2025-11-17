package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/quangdang46/NFT-Marketplace/services/auth-service/internal/models"
	"gorm.io/gorm"
)

// LoginEventRepository handles login event audit logging operations
type LoginEventRepository interface {
	// CreateLoginEvent creates a new login event audit log
	CreateLoginEvent(ctx context.Context, event *models.LoginEvent) error

	// GetLoginEventsByUserID retrieves login events for a user
	GetLoginEventsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*models.LoginEvent, error)

	// GetLoginEventsByAccountID retrieves login events for an account
	GetLoginEventsByAccountID(ctx context.Context, accountID string, limit int) ([]*models.LoginEvent, error)

	// GetLoginEventsByIPAddress retrieves login events for an IP address
	GetLoginEventsByIPAddress(ctx context.Context, ipAddress string, limit int) ([]*models.LoginEvent, error)

	// GetRecentFailedAttempts retrieves recent failed login attempts
	GetRecentFailedAttempts(ctx context.Context, accountID string, since time.Time) ([]*models.LoginEvent, error)

	// CountFailedAttempts counts failed login attempts within a time window
	CountFailedAttempts(ctx context.Context, accountID string, ipAddress *string, since time.Time) (int64, error)

	// CleanupOldEvents removes login events older than retention period (uses DB function)
	CleanupOldEvents(ctx context.Context, retentionDays int) (int64, error)
}

type loginEventRepository struct {
	db *gorm.DB
}

// NewLoginEventRepository creates a new login event repository
func NewLoginEventRepository(db *gorm.DB) LoginEventRepository {
	return &loginEventRepository{db: db}
}

// CreateLoginEvent creates a new login event audit log
func (r *loginEventRepository) CreateLoginEvent(ctx context.Context, event *models.LoginEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// GetLoginEventsByUserID retrieves login events for a user
func (r *loginEventRepository) GetLoginEventsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*models.LoginEvent, error) {
	var events []*models.LoginEvent
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&events).Error

	if err != nil {
		return nil, err
	}

	return events, nil
}

// GetLoginEventsByAccountID retrieves login events for an account
func (r *loginEventRepository) GetLoginEventsByAccountID(ctx context.Context, accountID string, limit int) ([]*models.LoginEvent, error) {
	var events []*models.LoginEvent
	err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&events).Error

	if err != nil {
		return nil, err
	}

	return events, nil
}

// GetLoginEventsByIPAddress retrieves login events for an IP address
func (r *loginEventRepository) GetLoginEventsByIPAddress(ctx context.Context, ipAddress string, limit int) ([]*models.LoginEvent, error) {
	var events []*models.LoginEvent
	err := r.db.WithContext(ctx).
		Where("ip_address = ?", ipAddress).
		Order("timestamp DESC").
		Limit(limit).
		Find(&events).Error

	if err != nil {
		return nil, err
	}

	return events, nil
}

// GetRecentFailedAttempts retrieves recent failed login attempts
func (r *loginEventRepository) GetRecentFailedAttempts(ctx context.Context, accountID string, since time.Time) ([]*models.LoginEvent, error) {
	var events []*models.LoginEvent
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND result != ? AND timestamp >= ?", accountID, models.LoginResultSuccess, since).
		Order("timestamp DESC").
		Find(&events).Error

	if err != nil {
		return nil, err
	}

	return events, nil
}

// CountFailedAttempts counts failed login attempts within a time window
func (r *loginEventRepository) CountFailedAttempts(ctx context.Context, accountID string, ipAddress *string, since time.Time) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&models.LoginEvent{}).
		Where("result != ? AND timestamp >= ?", models.LoginResultSuccess, since)

	if accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}

	if ipAddress != nil && *ipAddress != "" {
		query = query.Where("ip_address = ?", *ipAddress)
	}

	err := query.Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CleanupOldEvents removes login events older than retention period using DB function
func (r *loginEventRepository) CleanupOldEvents(ctx context.Context, retentionDays int) (int64, error) {
	var deletedCount int64

	err := r.db.WithContext(ctx).
		Raw("SELECT cleanup_old_login_events(?)", retentionDays).
		Scan(&deletedCount).Error

	if err != nil {
		return 0, err
	}

	return deletedCount, nil
}
