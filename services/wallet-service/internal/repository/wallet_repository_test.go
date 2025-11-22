package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/quangdang46/NFT-Marketplace/services/wallet-service/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create the wallet_links table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS wallet_links (
			wallet_id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
			user_id TEXT NOT NULL,
			account_id TEXT NOT NULL,
			address TEXT NOT NULL,
			chain_id TEXT NOT NULL,
			is_primary INTEGER NOT NULL DEFAULT 0,
			type TEXT DEFAULT 'eoa',
			connector TEXT,
			label TEXT,
			verified_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Create indexes
	db.Exec("CREATE INDEX idx_wallet_links_user_id ON wallet_links(user_id)")
	db.Exec("CREATE INDEX idx_wallet_links_address ON wallet_links(address)")
	db.Exec("CREATE INDEX idx_wallet_links_chain_id ON wallet_links(chain_id)")
	db.Exec("CREATE UNIQUE INDEX idx_wallet_user_address_chain ON wallet_links(user_id, address, chain_id)")

	return db
}

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "valid lowercase address",
			address: "0x1234567890123456789012345678901234567890",
			wantErr: false,
		},
		{
			name:    "invalid - uppercase",
			address: "0X1234567890123456789012345678901234567890",
			wantErr: true,
		},
		{
			name:    "invalid - too short",
			address: "0x12345",
			wantErr: true,
		},
		{
			name:    "invalid - too long",
			address: "0x12345678901234567890123456789012345678901",
			wantErr: true,
		},
		{
			name:    "invalid - missing 0x prefix",
			address: "1234567890123456789012345678901234567890",
			wantErr: true,
		},
		{
			name:    "invalid - non-hex characters",
			address: "0x123456789012345678901234567890123456789g",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != ErrInvalidAddress {
				t.Errorf("Expected ErrInvalidAddress, got %v", err)
			}
		})
	}
}

func TestWalletRepository_UpsertLink_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	address := "0x1234567890123456789012345678901234567890"
	chainID := "eip155:1"
	accountID := "eip155:1:0x1234567890123456789012345678901234567890"

	link := &models.WalletLink{
		UserID:     userID,
		AccountID:  accountID,
		Address:    address,
		ChainID:    chainID,
		IsPrimary:  true,
		Type:       models.WalletTypeEOA,
		VerifiedAt: time.Now(),
	}

	err := repo.UpsertLink(ctx, link)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify wallet was created
	retrieved, err := repo.GetByAddress(ctx, address, chainID)
	if err != nil {
		t.Fatalf("Expected to retrieve wallet, got error: %v", err)
	}

	if retrieved.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, retrieved.UserID)
	}

	if retrieved.Address != address {
		t.Errorf("Expected address %s, got %s", address, retrieved.Address)
	}

	if !retrieved.IsPrimary {
		t.Error("Expected wallet to be primary")
	}
}

func TestWalletRepository_UpsertLink_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	address := "0x1234567890123456789012345678901234567890"
	chainID := "eip155:1"
	accountID := "eip155:1:0x1234567890123456789012345678901234567890"

	// Create initial link
	link := &models.WalletLink{
		UserID:     userID,
		AccountID:  accountID,
		Address:    address,
		ChainID:    chainID,
		IsPrimary:  false,
		Type:       models.WalletTypeEOA,
		VerifiedAt: time.Now(),
	}

	err := repo.UpsertLink(ctx, link)
	if err != nil {
		t.Fatalf("Failed to create initial link: %v", err)
	}

	// Update the same link
	link.IsPrimary = true
	label := "My Primary Wallet"
	link.Label = &label

	err = repo.UpsertLink(ctx, link)
	if err != nil {
		t.Fatalf("Expected no error on update, got %v", err)
	}

	// Verify update
	retrieved, err := repo.GetByAddress(ctx, address, chainID)
	if err != nil {
		t.Fatalf("Expected to retrieve wallet, got error: %v", err)
	}

	if !retrieved.IsPrimary {
		t.Error("Expected wallet to be primary after update")
	}

	if retrieved.Label == nil || *retrieved.Label != label {
		t.Errorf("Expected label %s, got %v", label, retrieved.Label)
	}
}

func TestWalletRepository_UpsertLink_InvalidAddress(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	link := &models.WalletLink{
		UserID:     userID,
		AccountID:  "eip155:1:0xINVALID",
		Address:    "0xINVALID",
		ChainID:    "eip155:1",
		IsPrimary:  false,
		Type:       models.WalletTypeEOA,
		VerifiedAt: time.Now(),
	}

	err := repo.UpsertLink(ctx, link)
	if err != ErrInvalidAddress {
		t.Errorf("Expected ErrInvalidAddress, got %v", err)
	}
}

func TestWalletRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Create multiple wallets for the user
	addresses := []string{
		"0x1111111111111111111111111111111111111111",
		"0x2222222222222222222222222222222222222222",
		"0x3333333333333333333333333333333333333333",
	}

	for i, addr := range addresses {
		link := &models.WalletLink{
			UserID:     userID,
			AccountID:  "eip155:1:" + addr,
			Address:    addr,
			ChainID:    "eip155:1",
			IsPrimary:  i == 0, // First one is primary
			Type:       models.WalletTypeEOA,
			VerifiedAt: time.Now(),
		}
		err := repo.UpsertLink(ctx, link)
		if err != nil {
			t.Fatalf("Failed to create wallet: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Retrieve all wallets
	wallets, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(wallets) != 3 {
		t.Errorf("Expected 3 wallets, got %d", len(wallets))
	}

	// Verify primary wallet is first
	if len(wallets) > 0 && !wallets[0].IsPrimary {
		t.Error("Expected first wallet to be primary")
	}
}

func TestWalletRepository_GetByAddress(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	address := "0xabcdef1234567890abcdef1234567890abcdef12"
	chainID := "eip155:1"
	accountID := "eip155:1:0xabcdef1234567890abcdef1234567890abcdef12"

	// Create wallet
	link := &models.WalletLink{
		UserID:     userID,
		AccountID:  accountID,
		Address:    address,
		ChainID:    chainID,
		IsPrimary:  true,
		Type:       models.WalletTypeEOA,
		VerifiedAt: time.Now(),
	}

	err := repo.UpsertLink(ctx, link)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Retrieve by address
	retrieved, err := repo.GetByAddress(ctx, address, chainID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.Address != address {
		t.Errorf("Expected address %s, got %s", address, retrieved.Address)
	}

	if retrieved.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, retrieved.UserID)
	}
}

func TestWalletRepository_GetByAddress_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	address := "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	chainID := "eip155:1"

	// Try to retrieve non-existent wallet
	_, err := repo.GetByAddress(ctx, address, chainID)
	if err != ErrWalletNotFound {
		t.Errorf("Expected ErrWalletNotFound, got %v", err)
	}
}

func TestWalletRepository_GetPrimaryWallet(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	primaryAddr := "0x1111111111111111111111111111111111111111"
	secondaryAddr := "0x2222222222222222222222222222222222222222"

	// Create primary wallet
	primaryLink := &models.WalletLink{
		UserID:     userID,
		AccountID:  "eip155:1:" + primaryAddr,
		Address:    primaryAddr,
		ChainID:    "eip155:1",
		IsPrimary:  true,
		Type:       models.WalletTypeEOA,
		VerifiedAt: time.Now(),
	}
	err := repo.UpsertLink(ctx, primaryLink)
	if err != nil {
		t.Fatalf("Failed to create primary wallet: %v", err)
	}

	// Create secondary wallet
	secondaryLink := &models.WalletLink{
		UserID:     userID,
		AccountID:  "eip155:1:" + secondaryAddr,
		Address:    secondaryAddr,
		ChainID:    "eip155:1",
		IsPrimary:  false,
		Type:       models.WalletTypeEOA,
		VerifiedAt: time.Now(),
	}
	err = repo.UpsertLink(ctx, secondaryLink)
	if err != nil {
		t.Fatalf("Failed to create secondary wallet: %v", err)
	}

	// Get primary wallet
	primary, err := repo.GetPrimaryWallet(ctx, userID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if primary.Address != primaryAddr {
		t.Errorf("Expected primary address %s, got %s", primaryAddr, primary.Address)
	}

	if !primary.IsPrimary {
		t.Error("Expected wallet to be marked as primary")
	}
}

func TestWalletRepository_GetPrimaryWallet_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Try to get primary wallet for user with no wallets
	_, err := repo.GetPrimaryWallet(ctx, userID)
	if err != ErrWalletNotFound {
		t.Errorf("Expected ErrWalletNotFound, got %v", err)
	}
}

func TestWalletRepository_SetPrimaryWallet(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	address := "0x1234567890123456789012345678901234567890"
	chainID := "eip155:1"
	accountID := "eip155:1:0x1234567890123456789012345678901234567890"

	// Create wallet as non-primary
	link := &models.WalletLink{
		UserID:     userID,
		AccountID:  accountID,
		Address:    address,
		ChainID:    chainID,
		IsPrimary:  false,
		Type:       models.WalletTypeEOA,
		VerifiedAt: time.Now(),
	}

	err := repo.UpsertLink(ctx, link)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Get the wallet ID
	retrieved, _ := repo.GetByAddress(ctx, address, chainID)

	// Set as primary
	err = repo.SetPrimaryWallet(ctx, retrieved.WalletID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify it's now primary
	updated, err := repo.GetByAddress(ctx, address, chainID)
	if err != nil {
		t.Fatalf("Failed to retrieve wallet: %v", err)
	}

	if !updated.IsPrimary {
		t.Error("Expected wallet to be primary")
	}
}

func TestWalletRepository_SetPrimaryWallet_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	// Try to set non-existent wallet as primary
	err := repo.SetPrimaryWallet(ctx, uuid.New())
	if err != ErrWalletNotFound {
		t.Errorf("Expected ErrWalletNotFound, got %v", err)
	}
}

func TestWalletRepository_UnlinkWallet(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	address := "0x1234567890123456789012345678901234567890"
	chainID := "eip155:1"
	accountID := "eip155:1:0x1234567890123456789012345678901234567890"

	// Create wallet
	link := &models.WalletLink{
		UserID:     userID,
		AccountID:  accountID,
		Address:    address,
		ChainID:    chainID,
		IsPrimary:  false,
		Type:       models.WalletTypeEOA,
		VerifiedAt: time.Now(),
	}

	err := repo.UpsertLink(ctx, link)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Get the wallet ID
	retrieved, _ := repo.GetByAddress(ctx, address, chainID)

	// Unlink wallet
	err = repo.UnlinkWallet(ctx, retrieved.WalletID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify wallet is deleted
	_, err = repo.GetByAddress(ctx, address, chainID)
	if err != ErrWalletNotFound {
		t.Errorf("Expected ErrWalletNotFound after unlink, got %v", err)
	}
}

func TestWalletRepository_UnlinkWallet_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	// Try to unlink non-existent wallet
	err := repo.UnlinkWallet(ctx, uuid.New())
	if err != ErrWalletNotFound {
		t.Errorf("Expected ErrWalletNotFound, got %v", err)
	}
}

func TestWalletRepository_AddressNormalization(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	addressMixed := "0xAbCdEf1234567890abcdef1234567890abcdef12"
	addressLower := "0xabcdef1234567890abcdef1234567890abcdef12"
	chainID := "eip155:1"
	accountID := "eip155:1:0xabcdef1234567890abcdef1234567890abcdef12"

	// Create wallet with mixed case (should fail validation)
	link := &models.WalletLink{
		UserID:     userID,
		AccountID:  accountID,
		Address:    addressMixed,
		ChainID:    chainID,
		IsPrimary:  false,
		Type:       models.WalletTypeEOA,
		VerifiedAt: time.Now(),
	}

	err := repo.UpsertLink(ctx, link)
	if err != ErrInvalidAddress {
		t.Errorf("Expected ErrInvalidAddress for mixed case, got %v", err)
	}

	// Create with lowercase (should succeed)
	link.Address = addressLower
	err = repo.UpsertLink(ctx, link)
	if err != nil {
		t.Fatalf("Expected no error for lowercase, got %v", err)
	}

	// Verify retrieval works
	retrieved, err := repo.GetByAddress(ctx, addressLower, chainID)
	if err != nil {
		t.Fatalf("Failed to retrieve wallet: %v", err)
	}

	if retrieved.Address != addressLower {
		t.Errorf("Expected address %s, got %s", addressLower, retrieved.Address)
	}
}
