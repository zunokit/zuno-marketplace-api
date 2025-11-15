package models

import "time"

// AuthNonce represents a one-time nonce for SIWE authentication
type AuthNonce struct {
	Nonce     string     `gorm:"primaryKey;size:64" json:"nonce"`
	AccountID string     `gorm:"size:42;not null;index:idx_account_id" json:"account_id"`
	Domain    string     `gorm:"size:255;not null" json:"domain"`
	ChainID   string     `gorm:"size:32;not null" json:"chain_id"`
	IssuedAt  time.Time  `gorm:"not null" json:"issued_at"`
	ExpiresAt time.Time  `gorm:"not null;index:idx_expires_at" json:"expires_at"`
	Used      bool       `gorm:"default:false;not null;index:idx_used" json:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP;not null" json:"created_at"`
}

// TableName specifies the table name for GORM
func (AuthNonce) TableName() string {
	return "auth_nonces"
}

// IsExpired checks if the nonce has expired
func (n *AuthNonce) IsExpired() bool {
	return time.Now().After(n.ExpiresAt)
}

// IsValid checks if the nonce is valid and can be used
func (n *AuthNonce) IsValid() bool {
	return !n.Used && !n.IsExpired()
}

// MarkAsUsed marks the nonce as used
func (n *AuthNonce) MarkAsUsed() {
	now := time.Now()
	n.Used = true
	n.UsedAt = &now
}
