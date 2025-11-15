package server

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/quangdang46/NFT-Marketplace/services/wallet-service/internal/models"
	"github.com/quangdang46/NFT-Marketplace/services/wallet-service/internal/repository"
	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WalletServer implements the gRPC WalletService
type WalletServer struct {
	pb.UnimplementedWalletServiceServer
	walletRepo repository.WalletRepository
}

// NewWalletServer creates a new wallet service gRPC server
func NewWalletServer(walletRepo repository.WalletRepository) *WalletServer {
	return &WalletServer{
		walletRepo: walletRepo,
	}
}

// UpsertLink creates or updates a wallet link (idempotent)
func (s *WalletServer) UpsertLink(ctx context.Context, req *pb.UpsertLinkRequest) (*pb.UpsertLinkResponse, error) {
	// Validate request
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}
	if req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "chain_id is required")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	// Check if wallet already exists
	existing, _ := s.walletRepo.GetByAddress(ctx, req.Address, req.ChainId)
	created := existing == nil

	// Get primary wallet before upsert to detect changes
	primaryBefore, _ := s.walletRepo.GetPrimaryWallet(ctx, userID)
	primaryChanged := false

	// Create wallet link model
	link := &models.WalletLink{
		UserID:     userID,
		AccountID:  req.AccountId,
		Address:    req.Address,
		ChainID:    req.ChainId,
		IsPrimary:  req.IsPrimary,
		Type:       models.WalletType(req.Type),
		VerifiedAt: time.Now(),
	}

	if req.Connector != "" {
		link.Connector = &req.Connector
	}
	if req.Label != "" {
		link.Label = &req.Label
	}

	// Upsert wallet link
	if err := s.walletRepo.UpsertLink(ctx, link); err != nil {
		if err == repository.ErrInvalidAddress {
			return nil, status.Error(codes.InvalidArgument, "invalid wallet address format")
		}
		return nil, status.Errorf(codes.Internal, "failed to upsert wallet link: %v", err)
	}

	// Check if primary wallet changed
	if req.IsPrimary {
		if primaryBefore == nil || primaryBefore.WalletID != link.WalletID {
			primaryChanged = true
		}
	}

	// Convert to proto
	pbLink := &pb.WalletLink{
		Id:         link.WalletID.String(),
		UserId:     link.UserID.String(),
		AccountId:  link.AccountID,
		Address:    link.Address,
		ChainId:    link.ChainID,
		IsPrimary:  link.IsPrimary,
		VerifiedAt: timestamppb.New(link.VerifiedAt),
		CreatedAt:  timestamppb.New(link.CreatedAt),
		UpdatedAt:  timestamppb.New(link.UpdatedAt),
	}

	return &pb.UpsertLinkResponse{
		Link:           pbLink,
		Created:        created,
		PrimaryChanged: primaryChanged,
	}, nil
}

// GetWallets retrieves all wallets for a user
func (s *WalletServer) GetWallets(ctx context.Context, req *pb.GetWalletsRequest) (*pb.GetWalletsResponse, error) {
	// Validate request
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	// Get wallets
	wallets, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get wallets: %v", err)
	}

	// Convert to proto
	pbWallets := make([]*pb.WalletLink, len(wallets))
	for i, wallet := range wallets {
		pbWallets[i] = &pb.WalletLink{
			Id:         wallet.WalletID.String(),
			UserId:     wallet.UserID.String(),
			AccountId:  wallet.AccountID,
			Address:    wallet.Address,
			ChainId:    wallet.ChainID,
			IsPrimary:  wallet.IsPrimary,
			VerifiedAt: timestamppb.New(wallet.VerifiedAt),
			CreatedAt:  timestamppb.New(wallet.CreatedAt),
			UpdatedAt:  timestamppb.New(wallet.UpdatedAt),
		}
	}

	return &pb.GetWalletsResponse{
		Wallets: pbWallets,
	}, nil
}
