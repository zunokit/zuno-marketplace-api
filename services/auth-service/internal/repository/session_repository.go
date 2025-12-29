package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/models"
	"gorm.io/gorm"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionRevoked  = errors.New("session has been revoked")
	ErrSessionExpired  = errors.New("session has expired")
)

// SessionRepository handles session persistence operations
type SessionRepository interface {
	// CreateSession creates a new session
	CreateSession(ctx context.Context, session *models.Session) error

	// GetByRefreshToken retrieves a session by refresh token hash
	GetByRefreshToken(ctx context.Context, refreshTokenHash string) (*models.Session, error)

	// GetByID retrieves a session by ID
	GetByID(ctx context.Context, sessionID uuid.UUID) (*models.Session, error)

	// UpdateSession updates a session
	UpdateSession(ctx context.Context, session *models.Session) error

	// RevokeSession revokes a session
	RevokeSession(ctx context.Context, sessionID uuid.UUID, reason string) error

	// RevokeByRefreshToken revokes a session by refresh token
	RevokeByRefreshToken(ctx context.Context, refreshTokenHash, reason string) error

	// RevokeTokenFamily revokes all sessions in a token family (for replay attack detection)
	RevokeTokenFamily(ctx context.Context, familyID uuid.UUID, reason string) error

	// GetActiveSessions retrieves all active sessions for a user
	GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)

	// CleanupExpired marks expired sessions as revoked
	CleanupExpired(ctx context.Context) (int64, error)
}

type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

// HashRefreshToken creates a SHA-256 hash of the refresh token
func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CreateSession creates a new session
func (r *sessionRepository) CreateSession(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetByRefreshToken retrieves a session by refresh token hash
func (r *sessionRepository) GetByRefreshToken(ctx context.Context, refreshTokenHash string) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).
		Where("refresh_hash = ?", refreshTokenHash).
		First(&session).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	return &session, nil
}

// GetByID retrieves a session by ID
func (r *sessionRepository) GetByID(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		First(&session).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	return &session, nil
}

// UpdateSession updates a session
func (r *sessionRepository) UpdateSession(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// RevokeSession revokes a session by ID
func (r *sessionRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID, reason string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("session_id = ? AND revoked_at IS NULL", sessionID).
		Updates(map[string]interface{}{
			"revoked_at":     now,
			"revoked_reason": reason,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// RevokeByRefreshToken revokes a session by refresh token hash
func (r *sessionRepository) RevokeByRefreshToken(ctx context.Context, refreshTokenHash, reason string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("refresh_hash = ? AND revoked_at IS NULL", refreshTokenHash).
		Updates(map[string]interface{}{
			"revoked_at":     now,
			"revoked_reason": reason,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// RevokeTokenFamily revokes all sessions in a token family
func (r *sessionRepository) RevokeTokenFamily(ctx context.Context, familyID uuid.UUID, reason string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("token_family_id = ? AND revoked_at IS NULL", familyID).
		Updates(map[string]interface{}{
			"revoked_at":     now,
			"revoked_reason": reason,
		}).Error
}

// GetActiveSessions retrieves all active sessions for a user
func (r *sessionRepository) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	var sessions []*models.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&sessions).Error

	if err != nil {
		return nil, err
	}

	return sessions, nil
}

// CleanupExpired calls the database function to mark expired sessions as revoked
func (r *sessionRepository) CleanupExpired(ctx context.Context) (int64, error) {
	var updatedCount int64

	err := r.db.WithContext(ctx).
		Raw("SELECT cleanup_expired_sessions()").
		Scan(&updatedCount).Error

	if err != nil {
		return 0, err
	}

	return updatedCount, nil
}
