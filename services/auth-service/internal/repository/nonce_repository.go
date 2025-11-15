package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/quangdang46/NFT-Marketplace/services/auth-service/internal/models"
	"gorm.io/gorm"
)

var (
	ErrNonceNotFound = errors.New("nonce not found")
	ErrNonceExpired  = errors.New("nonce has expired")
	ErrNonceUsed     = errors.New("nonce already used")
	ErrNonceInvalid  = errors.New("nonce is invalid")
)

// NonceRepository handles nonce persistence operations
type NonceRepository interface {
	// CreateNonce generates and stores a new nonce
	CreateNonce(ctx context.Context, accountID, domain, chainID string, ttl time.Duration) (*models.AuthNonce, error)

	// GetNonce retrieves a nonce by its value
	GetNonce(ctx context.Context, nonce string) (*models.AuthNonce, error)

	// ValidateAndConsumeNonce atomically validates and marks a nonce as used
	ValidateAndConsumeNonce(ctx context.Context, nonce, accountID, chainID, domain string) error

	// CleanupExpired removes expired nonces (older than retention period)
	CleanupExpired(ctx context.Context, retentionPeriod time.Duration) (int64, error)
}

type nonceRepository struct {
	db *gorm.DB
}

// NewNonceRepository creates a new nonce repository
func NewNonceRepository(db *gorm.DB) NonceRepository {
	return &nonceRepository{db: db}
}

// CreateNonce generates a cryptographically secure nonce and stores it
func (r *nonceRepository) CreateNonce(ctx context.Context, accountID, domain, chainID string, ttl time.Duration) (*models.AuthNonce, error) {
	// Generate cryptographically secure random nonce (32 bytes = 64 hex chars)
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, err
	}
	nonceValue := hex.EncodeToString(nonceBytes)

	now := time.Now()
	nonce := &models.AuthNonce{
		Nonce:     nonceValue,
		AccountID: accountID,
		Domain:    domain,
		ChainID:   chainID,
		IssuedAt:  now,
		ExpiresAt: now.Add(ttl),
		Used:      false,
		CreatedAt: now,
	}

	if err := r.db.WithContext(ctx).Create(nonce).Error; err != nil {
		return nil, err
	}

	return nonce, nil
}

// GetNonce retrieves a nonce by its value
func (r *nonceRepository) GetNonce(ctx context.Context, nonce string) (*models.AuthNonce, error) {
	var authNonce models.AuthNonce
	err := r.db.WithContext(ctx).
		Where("nonce = ?", nonce).
		First(&authNonce).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNonceNotFound
		}
		return nil, err
	}

	return &authNonce, nil
}

// ValidateAndConsumeNonce uses the database function try_use_nonce for atomic consume
func (r *nonceRepository) ValidateAndConsumeNonce(ctx context.Context, nonce, accountID, chainID, domain string) error {
	var success bool

	// Call the PostgreSQL function try_use_nonce
	err := r.db.WithContext(ctx).
		Raw("SELECT try_use_nonce(?, ?, ?, ?)", nonce, accountID, chainID, domain).
		Scan(&success).Error

	if err != nil {
		return err
	}

	if !success {
		// Nonce was not consumed - need to check why
		existingNonce, err := r.GetNonce(ctx, nonce)
		if err != nil {
			if errors.Is(err, ErrNonceNotFound) {
				return ErrNonceInvalid
			}
			return err
		}

		if existingNonce.Used {
			return ErrNonceUsed
		}

		if existingNonce.IsExpired() {
			return ErrNonceExpired
		}

		// Nonce exists but parameters don't match
		return ErrNonceInvalid
	}

	return nil
}

// CleanupExpired removes expired nonces using the database function
func (r *nonceRepository) CleanupExpired(ctx context.Context, retentionPeriod time.Duration) (int64, error) {
	var deletedCount int64

	err := r.db.WithContext(ctx).
		Raw("SELECT cleanup_expired_nonces()").
		Scan(&deletedCount).Error

	if err != nil {
		return 0, err
	}

	return deletedCount, nil
}
