package models

import (
	"time"

	"github.com/google/uuid"
)

// LoginResult represents the result of a login attempt
type LoginResult string

const (
	LoginResultSuccess          LoginResult = "success"
	LoginResultFailed           LoginResult = "failed"
	LoginResultInvalidSignature LoginResult = "invalid_signature"
	LoginResultInvalidNonce     LoginResult = "invalid_nonce"
	LoginResultExpiredNonce     LoginResult = "expired_nonce"
	LoginResultInvalidMessage   LoginResult = "invalid_message"
	LoginResultRateLimited      LoginResult = "rate_limited"
)

// LoginEvent represents an audit log entry for authentication attempts
type LoginEvent struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       *uuid.UUID  `gorm:"type:uuid;index:idx_user_id" json:"user_id,omitempty"`
	AccountID    string      `gorm:"size:42;not null;index:idx_account_id" json:"account_id"`
	IPAddress    *string     `gorm:"type:inet;index:idx_ip_address" json:"ip_address,omitempty"`
	UserAgent    *string     `gorm:"type:text" json:"user_agent,omitempty"`
	Result       LoginResult `gorm:"size:32;not null;index:idx_result" json:"result"`
	ErrorMessage *string     `gorm:"type:text" json:"error_message,omitempty"`
	ChainID      *string     `gorm:"size:32" json:"chain_id,omitempty"`
	Domain       *string     `gorm:"size:255" json:"domain,omitempty"`
	Timestamp    time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_timestamp" json:"timestamp"`
}

// TableName specifies the table name for GORM
func (LoginEvent) TableName() string {
	return "login_events"
}

// IsSuccess checks if the login was successful
func (e *LoginEvent) IsSuccess() bool {
	return e.Result == LoginResultSuccess
}
