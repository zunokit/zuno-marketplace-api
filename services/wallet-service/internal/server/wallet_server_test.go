package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zunokit/zuno-marketplace-api/services/wallet-service/internal/models"
	"github.com/zunokit/zuno-marketplace-api/services/wallet-service/internal/repository"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mock WalletRepository for testing
type mockWalletRepository struct {
	upsertLinkFunc       func(ctx context.Context, link *models.WalletLink) error
	getByUserIDFunc      func(ctx context.Context, userID uuid.UUID) ([]*models.WalletLink, error)
	getByAddressFunc     func(ctx context.Context, address, chainID string) (*models.WalletLink, error)
	getPrimaryWalletFunc func(ctx context.Context, userID uuid.UUID) (*models.WalletLink, error)
	setPrimaryWalletFunc func(ctx context.Context, walletID uuid.UUID) error
	unlinkWalletFunc     func(ctx context.Context, walletID uuid.UUID) error
}

func (m *mockWalletRepository) UpsertLink(ctx context.Context, link *models.WalletLink) error {
	if m.upsertLinkFunc != nil {
		return m.upsertLinkFunc(ctx, link)
	}
	return nil
}

func (m *mockWalletRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.WalletLink, error) {
	if m.getByUserIDFunc != nil {
		return m.getByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockWalletRepository) GetByAddress(ctx context.Context, address, chainID string) (*models.WalletLink, error) {
	if m.getByAddressFunc != nil {
		return m.getByAddressFunc(ctx, address, chainID)
	}
	return nil, nil
}

func (m *mockWalletRepository) GetPrimaryWallet(ctx context.Context, userID uuid.UUID) (*models.WalletLink, error) {
	if m.getPrimaryWalletFunc != nil {
		return m.getPrimaryWalletFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockWalletRepository) SetPrimaryWallet(ctx context.Context, walletID uuid.UUID) error {
	if m.setPrimaryWalletFunc != nil {
		return m.setPrimaryWalletFunc(ctx, walletID)
	}
	return nil
}

func (m *mockWalletRepository) UnlinkWallet(ctx context.Context, walletID uuid.UUID) error {
	if m.unlinkWalletFunc != nil {
		return m.unlinkWalletFunc(ctx, walletID)
	}
	return nil
}

func TestWalletServer_UpsertLink_Success(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()
	address := "0x1234567890123456789012345678901234567890"
	chainID := "eip155:1"
	accountID := "eip155:1:0x1234567890123456789012345678901234567890"

	mockRepo := &mockWalletRepository{
		getByAddressFunc: func(ctx context.Context, addr, chain string) (*models.WalletLink, error) {
			// Return not found to simulate new wallet
			return nil, repository.ErrWalletNotFound
		},
		getPrimaryWalletFunc: func(ctx context.Context, uid uuid.UUID) (*models.WalletLink, error) {
			// No existing primary wallet
			return nil, repository.ErrWalletNotFound
		},
		upsertLinkFunc: func(ctx context.Context, link *models.WalletLink) error {
			// Simulate successful upsert
			link.WalletID = walletID
			link.CreatedAt = time.Now()
			link.UpdatedAt = time.Now()
			return nil
		},
	}

	server := NewWalletServer(mockRepo)

	req := &pb.UpsertLinkRequest{
		UserId:    userID.String(),
		AccountId: accountID,
		Address:   address,
		ChainId:   chainID,
		IsPrimary: true,
		Type:      "eoa",
	}

	resp, err := server.UpsertLink(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Link == nil {
		t.Fatal("Expected link in response")
	}

	if resp.Link.Address != address {
		t.Errorf("Expected address %s, got %s", address, resp.Link.Address)
	}

	if !resp.Created {
		t.Error("Expected Created to be true for new wallet")
	}

	if !resp.PrimaryChanged {
		t.Error("Expected PrimaryChanged to be true when setting first primary")
	}
}

func TestWalletServer_UpsertLink_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		req     *pb.UpsertLinkRequest
		wantErr codes.Code
	}{
		{
			name: "missing user_id",
			req: &pb.UpsertLinkRequest{
				Address: "0x1234567890123456789012345678901234567890",
				ChainId: "eip155:1",
			},
			wantErr: codes.InvalidArgument,
		},
		{
			name: "missing address",
			req: &pb.UpsertLinkRequest{
				UserId:  uuid.New().String(),
				ChainId: "eip155:1",
			},
			wantErr: codes.InvalidArgument,
		},
		{
			name: "missing chain_id",
			req: &pb.UpsertLinkRequest{
				UserId:  uuid.New().String(),
				Address: "0x1234567890123456789012345678901234567890",
			},
			wantErr: codes.InvalidArgument,
		},
		{
			name: "invalid user_id format",
			req: &pb.UpsertLinkRequest{
				UserId:  "invalid-uuid",
				Address: "0x1234567890123456789012345678901234567890",
				ChainId: "eip155:1",
			},
			wantErr: codes.InvalidArgument,
		},
	}

	mockRepo := &mockWalletRepository{}
	server := NewWalletServer(mockRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := server.UpsertLink(context.Background(), tt.req)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatal("Expected gRPC status error")
			}

			if st.Code() != tt.wantErr {
				t.Errorf("Expected error code %v, got %v", tt.wantErr, st.Code())
			}
		})
	}
}

func TestWalletServer_UpsertLink_InvalidAddress(t *testing.T) {
	mockRepo := &mockWalletRepository{
		upsertLinkFunc: func(ctx context.Context, link *models.WalletLink) error {
			return repository.ErrInvalidAddress
		},
	}

	server := NewWalletServer(mockRepo)

	req := &pb.UpsertLinkRequest{
		UserId:    uuid.New().String(),
		AccountId: "eip155:1:0xINVALID",
		Address:   "0xINVALID",
		ChainId:   "eip155:1",
		IsPrimary: false,
		Type:      "eoa",
	}

	_, err := server.UpsertLink(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for invalid address")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected InvalidArgument error code, got %v", st.Code())
	}
}

func TestWalletServer_UpsertLink_Update(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()
	address := "0x1234567890123456789012345678901234567890"
	chainID := "eip155:1"
	accountID := "eip155:1:0x1234567890123456789012345678901234567890"

	existingWallet := &models.WalletLink{
		WalletID:   walletID,
		UserID:     userID,
		AccountID:  accountID,
		Address:    address,
		ChainID:    chainID,
		IsPrimary:  false,
		Type:       models.WalletTypeEOA,
		VerifiedAt: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockRepo := &mockWalletRepository{
		getByAddressFunc: func(ctx context.Context, addr, chain string) (*models.WalletLink, error) {
			// Return existing wallet
			return existingWallet, nil
		},
		getPrimaryWalletFunc: func(ctx context.Context, uid uuid.UUID) (*models.WalletLink, error) {
			return nil, repository.ErrWalletNotFound
		},
		upsertLinkFunc: func(ctx context.Context, link *models.WalletLink) error {
			link.WalletID = walletID
			return nil
		},
	}

	server := NewWalletServer(mockRepo)

	req := &pb.UpsertLinkRequest{
		UserId:    userID.String(),
		AccountId: accountID,
		Address:   address,
		ChainId:   chainID,
		IsPrimary: false,
		Type:      "eoa",
		Label:     "Updated Label",
	}

	resp, err := server.UpsertLink(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Created {
		t.Error("Expected Created to be false for existing wallet")
	}
}

func TestWalletServer_GetWallets_Success(t *testing.T) {
	userID := uuid.New()
	walletID1 := uuid.New()
	walletID2 := uuid.New()
	now := time.Now()

	wallets := []*models.WalletLink{
		{
			WalletID:   walletID1,
			UserID:     userID,
			AccountID:  "eip155:1:0x1111111111111111111111111111111111111111",
			Address:    "0x1111111111111111111111111111111111111111",
			ChainID:    "eip155:1",
			IsPrimary:  true,
			Type:       models.WalletTypeEOA,
			VerifiedAt: now,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			WalletID:   walletID2,
			UserID:     userID,
			AccountID:  "eip155:137:0x2222222222222222222222222222222222222222",
			Address:    "0x2222222222222222222222222222222222222222",
			ChainID:    "eip155:137",
			IsPrimary:  false,
			Type:       models.WalletTypeEOA,
			VerifiedAt: now,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	mockRepo := &mockWalletRepository{
		getByUserIDFunc: func(ctx context.Context, uid uuid.UUID) ([]*models.WalletLink, error) {
			if uid != userID {
				return nil, nil
			}
			return wallets, nil
		},
	}

	server := NewWalletServer(mockRepo)

	req := &pb.GetWalletsRequest{
		UserId: userID.String(),
	}

	resp, err := server.GetWallets(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resp.Wallets) != 2 {
		t.Errorf("Expected 2 wallets, got %d", len(resp.Wallets))
	}

	// Verify first wallet is primary
	if len(resp.Wallets) > 0 && !resp.Wallets[0].IsPrimary {
		t.Error("Expected first wallet to be primary")
	}
}

func TestWalletServer_GetWallets_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		req     *pb.GetWalletsRequest
		wantErr codes.Code
	}{
		{
			name:    "missing user_id",
			req:     &pb.GetWalletsRequest{},
			wantErr: codes.InvalidArgument,
		},
		{
			name: "invalid user_id format",
			req: &pb.GetWalletsRequest{
				UserId: "invalid-uuid",
			},
			wantErr: codes.InvalidArgument,
		},
	}

	mockRepo := &mockWalletRepository{}
	server := NewWalletServer(mockRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := server.GetWallets(context.Background(), tt.req)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatal("Expected gRPC status error")
			}

			if st.Code() != tt.wantErr {
				t.Errorf("Expected error code %v, got %v", tt.wantErr, st.Code())
			}
		})
	}
}

func TestWalletServer_GetWallets_EmptyResult(t *testing.T) {
	userID := uuid.New()

	mockRepo := &mockWalletRepository{
		getByUserIDFunc: func(ctx context.Context, uid uuid.UUID) ([]*models.WalletLink, error) {
			return []*models.WalletLink{}, nil
		},
	}

	server := NewWalletServer(mockRepo)

	req := &pb.GetWalletsRequest{
		UserId: userID.String(),
	}

	resp, err := server.GetWallets(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resp.Wallets) != 0 {
		t.Errorf("Expected 0 wallets, got %d", len(resp.Wallets))
	}
}

func TestWalletServer_GetWallets_RepositoryError(t *testing.T) {
	userID := uuid.New()

	mockRepo := &mockWalletRepository{
		getByUserIDFunc: func(ctx context.Context, uid uuid.UUID) ([]*models.WalletLink, error) {
			return nil, errors.New("database error")
		},
	}

	server := NewWalletServer(mockRepo)

	req := &pb.GetWalletsRequest{
		UserId: userID.String(),
	}

	_, err := server.GetWallets(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.Internal {
		t.Errorf("Expected Internal error code, got %v", st.Code())
	}
}

func TestWalletServer_UpsertLink_WithOptionalFields(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()
	address := "0x1234567890123456789012345678901234567890"
	chainID := "eip155:1"
	accountID := "eip155:1:0x1234567890123456789012345678901234567890"
	connector := "metamask"
	label := "My Wallet"

	mockRepo := &mockWalletRepository{
		getByAddressFunc: func(ctx context.Context, addr, chain string) (*models.WalletLink, error) {
			return nil, repository.ErrWalletNotFound
		},
		getPrimaryWalletFunc: func(ctx context.Context, uid uuid.UUID) (*models.WalletLink, error) {
			return nil, repository.ErrWalletNotFound
		},
		upsertLinkFunc: func(ctx context.Context, link *models.WalletLink) error {
			link.WalletID = walletID

			// Verify optional fields are set
			if link.Connector == nil || *link.Connector != connector {
				t.Errorf("Expected connector %s, got %v", connector, link.Connector)
			}
			if link.Label == nil || *link.Label != label {
				t.Errorf("Expected label %s, got %v", label, link.Label)
			}

			return nil
		},
	}

	server := NewWalletServer(mockRepo)

	req := &pb.UpsertLinkRequest{
		UserId:    userID.String(),
		AccountId: accountID,
		Address:   address,
		ChainId:   chainID,
		IsPrimary: false,
		Type:      "eoa",
		Connector: connector,
		Label:     label,
	}

	_, err := server.UpsertLink(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestWalletServer_UpsertLink_PrimaryChanged(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()
	oldPrimaryWalletID := uuid.New()
	address := "0x1234567890123456789012345678901234567890"
	chainID := "eip155:1"
	accountID := "eip155:1:0x1234567890123456789012345678901234567890"

	oldPrimaryWallet := &models.WalletLink{
		WalletID:  oldPrimaryWalletID,
		UserID:    userID,
		Address:   "0x9999999999999999999999999999999999999999",
		ChainID:   chainID,
		IsPrimary: true,
	}

	mockRepo := &mockWalletRepository{
		getByAddressFunc: func(ctx context.Context, addr, chain string) (*models.WalletLink, error) {
			return nil, repository.ErrWalletNotFound
		},
		getPrimaryWalletFunc: func(ctx context.Context, uid uuid.UUID) (*models.WalletLink, error) {
			return oldPrimaryWallet, nil
		},
		upsertLinkFunc: func(ctx context.Context, link *models.WalletLink) error {
			link.WalletID = walletID
			return nil
		},
	}

	server := NewWalletServer(mockRepo)

	req := &pb.UpsertLinkRequest{
		UserId:    userID.String(),
		AccountId: accountID,
		Address:   address,
		ChainId:   chainID,
		IsPrimary: true,
		Type:      "eoa",
	}

	resp, err := server.UpsertLink(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.PrimaryChanged {
		t.Error("Expected PrimaryChanged to be true when primary changes")
	}
}
