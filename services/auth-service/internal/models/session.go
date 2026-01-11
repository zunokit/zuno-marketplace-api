package models

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a user session after successful authentication
type Session struct {
	SessionID               uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"session_id"`
	UserID                  uuid.UUID              `gorm:"type:uuid;not null;index:idx_user_id,idx_active_by_user" json:"user_id"`
	DeviceID                *uuid.UUID             `gorm:"type:uuid" json:"device_id,omitempty"`
	RefreshHash             string                 `gorm:"size:128;not null;uniqueIndex:uq_refresh_hash" json:"-"`
	PreviousRefreshHash     *string                `gorm:"size:128;index:idx_prev_refresh" json:"-"`
	TokenFamilyID           uuid.UUID              `gorm:"type:uuid;not null;default:gen_random_uuid();index:idx_token_family" json:"token_family_id"`
	TokenGeneration         int                    `gorm:"not null;default:1" json:"token_generation"`
	IPAddress               *string                `gorm:"type:inet" json:"ip_address,omitempty"`
	UserAgent               *string                `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt               time.Time              `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	ExpiresAt               time.Time              `gorm:"not null;index:idx_expires_at" json:"expires_at"`
	RevokedAt               *time.Time             `json:"revoked_at,omitempty"`
	LastUsedAt              time.Time              `gorm:"default:CURRENT_TIMESTAMP" json:"last_used_at"`
	RevokedReason           *string                `gorm:"size:255" json:"revoked_reason,omitempty"`
	CollectionIntentContext map[string]interface{} `gorm:"type:jsonb" json:"collection_intent_context,omitempty"`
}

// TableName specifies the table name for GORM
func (Session) TableName() string {
	return "sessions"
}

// IsActive checks if the session is active (not revoked and not expired)
func (s *Session) IsActive() bool {
	return s.RevokedAt == nil && time.Now().Before(s.ExpiresAt)
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Revoke revokes the session with a reason
func (s *Session) Revoke(reason string) {
	now := time.Now()
	s.RevokedAt = &now
	s.RevokedReason = &reason
}

// UpdateLastUsed updates the last used timestamp
func (s *Session) UpdateLastUsed() {
	s.LastUsedAt = time.Now()
}
