package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/quangdang46/NFT-Marketplace/services/auth-service/internal/models"
	"gorm.io/gorm"
)

var (
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrWalletAlreadyLinked = errors.New("wallet already linked to this user")
	ErrInvalidAddress      = errors.New("invalid wallet address format")
)

// WalletRepository handles wallet link persistence operations
type WalletRepository interface {
	// UpsertLink creates or updates a wallet link (idempotent)
	UpsertLink(ctx context.Context, link *models.WalletLink) error

	// GetByUserID retrieves all wallets for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.WalletLink, error)

	// GetByAddress retrieves a wallet by address and chain
	GetByAddress(ctx context.Context, address, chainID string) (*models.WalletLink, error)

	// GetPrimaryWallet retrieves the primary wallet for a user
	GetPrimaryWallet(ctx context.Context, userID uuid.UUID) (*models.WalletLink, error)

	// SetPrimaryWallet sets a wallet as the primary wallet for a user
	SetPrimaryWallet(ctx context.Context, walletID uuid.UUID) error

	// UnlinkWallet removes a wallet link
	UnlinkWallet(ctx context.Context, walletID uuid.UUID) error
}

type walletRepository struct {
	db *gorm.DB
}

// NewWalletRepository creates a new wallet repository
func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db: db}
}

// ValidateAddress validates Ethereum address format (lowercase 0x + 40 hex chars)
func ValidateAddress(address string) error {
	if len(address) != 42 {
		return ErrInvalidAddress
	}
	if !strings.HasPrefix(address, "0x") {
		return ErrInvalidAddress
	}
	if address != strings.ToLower(address) {
		return ErrInvalidAddress
	}
	// Check if remaining 40 chars are hex
	for _, c := range address[2:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return ErrInvalidAddress
		}
	}
	return nil
}

// UpsertLink creates or updates a wallet link (idempotent)
// If wallet already exists for this user/address/chain, it updates the record
func (r *walletRepository) UpsertLink(ctx context.Context, link *models.WalletLink) error {
	// Validate address format
	if err := ValidateAddress(link.Address); err != nil {
		return err
	}

	// Normalize address to lowercase
	link.Address = strings.ToLower(link.Address)

	// Use INSERT ... ON CONFLICT UPDATE for true upsert behavior
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Try to find existing wallet link
		var existing models.WalletLink
		err := tx.Where("user_id = ? AND address = ? AND chain_id = ?",
			link.UserID, link.Address, link.ChainID).
			First(&existing).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new wallet link
			return tx.Create(link).Error
		}

		// Update existing wallet link
		link.WalletID = existing.WalletID
		return tx.Model(&models.WalletLink{}).
			Where("wallet_id = ?", existing.WalletID).
			Updates(link).Error
	})

	return err
}

// GetByUserID retrieves all wallets for a user
func (r *walletRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.WalletLink, error) {
	var wallets []*models.WalletLink
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_primary DESC, created_at DESC").
		Find(&wallets).Error

	if err != nil {
		return nil, err
	}

	return wallets, nil
}

// GetByAddress retrieves a wallet by address and chain
func (r *walletRepository) GetByAddress(ctx context.Context, address, chainID string) (*models.WalletLink, error) {
	// Normalize address
	address = strings.ToLower(address)

	var wallet models.WalletLink
	err := r.db.WithContext(ctx).
		Where("address = ? AND chain_id = ?", address, chainID).
		First(&wallet).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}

	return &wallet, nil
}

// GetPrimaryWallet retrieves the primary wallet for a user
func (r *walletRepository) GetPrimaryWallet(ctx context.Context, userID uuid.UUID) (*models.WalletLink, error) {
	var wallet models.WalletLink
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_primary = ?", userID, true).
		First(&wallet).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}

	return &wallet, nil
}

// SetPrimaryWallet sets a wallet as the primary wallet for a user
// This will automatically unset other wallets via database trigger
func (r *walletRepository) SetPrimaryWallet(ctx context.Context, walletID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.WalletLink{}).
		Where("wallet_id = ?", walletID).
		Update("is_primary", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrWalletNotFound
	}

	return nil
}

// UnlinkWallet removes a wallet link
func (r *walletRepository) UnlinkWallet(ctx context.Context, walletID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Delete(&models.WalletLink{}, "wallet_id = ?", walletID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrWalletNotFound
	}

	return nil
}
