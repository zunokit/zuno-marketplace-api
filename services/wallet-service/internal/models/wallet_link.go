package models

import (
	"time"

	"github.com/google/uuid"
)

// WalletType represents the type of wallet
type WalletType string

const (
	WalletTypeEOA          WalletType = "eoa"
	WalletTypeContract     WalletType = "contract"
	WalletTypeMultisig     WalletType = "multisig"
	WalletTypeSmartAccount WalletType = "smart_account"
)

// WalletLink represents a user's linked wallet
type WalletLink struct {
	WalletID   uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"wallet_id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_wallet_links_user_id,idx_wallet_user_address_chain" json:"user_id"`
	AccountID  string     `gorm:"size:255;not null" json:"account_id"`
	Address    string     `gorm:"size:42;not null;index:idx_wallet_links_address,idx_wallet_user_address_chain" json:"address"`
	ChainID    string     `gorm:"size:32;not null;index:idx_wallet_links_chain_id,idx_wallet_user_address_chain" json:"chain_id"`
	IsPrimary  bool       `gorm:"not null;default:false" json:"is_primary"`
	Type       WalletType `gorm:"size:20;default:'eoa'" json:"type"`
	Connector  *string    `gorm:"size:50" json:"connector,omitempty"`
	Label      *string    `gorm:"size:100" json:"label,omitempty"`
	VerifiedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"verified_at"`
	CreatedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_wallet_links_created_at" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (WalletLink) TableName() string {
	return "wallet_links"
}

// IsVerified checks if the wallet has been verified
func (w *WalletLink) IsVerified() bool {
	return !w.VerifiedAt.IsZero()
}

// SetAsPrimary marks this wallet as the primary wallet
func (w *WalletLink) SetAsPrimary() {
	w.IsPrimary = true
	w.UpdatedAt = time.Now()
}

// UnsetPrimary marks this wallet as non-primary
func (w *WalletLink) UnsetPrimary() {
	w.IsPrimary = false
	w.UpdatedAt = time.Now()
}
