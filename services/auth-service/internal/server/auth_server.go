package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/quangdang46/NFT-Marketplace/services/auth-service/internal/client"
	"github.com/quangdang46/NFT-Marketplace/services/auth-service/internal/models"
	"github.com/quangdang46/NFT-Marketplace/services/auth-service/internal/repository"
	"github.com/quangdang46/NFT-Marketplace/services/auth-service/internal/service"
	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	nonceRepo   repository.NonceRepository
	sessionRepo repository.SessionRepository
	siweService *service.SIWEService
	jwtService  *service.JWTService
	clients     *client.ServiceClients
}

func NewAuthServer(
	nonceRepo repository.NonceRepository,
	sessionRepo repository.SessionRepository,
	siweService *service.SIWEService,
	jwtService *service.JWTService,
	clients *client.ServiceClients,
) *AuthServer {
	return &AuthServer{
		nonceRepo:   nonceRepo,
		sessionRepo: sessionRepo,
		siweService: siweService,
		jwtService:  jwtService,
		clients:     clients,
	}
}

func (s *AuthServer) GetNonce(ctx context.Context, req *pb.GetNonceRequest) (*pb.GetNonceResponse, error) {
	if req.AccountId == "" || req.ChainId == "" || req.Domain == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id, chain_id, and domain are required")
	}

	nonce, err := s.nonceRepo.CreateNonce(ctx, strings.ToLower(req.AccountId), req.Domain, req.ChainId, 5*time.Minute)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create nonce: %v", err)
	}

	return &pb.GetNonceResponse{
		Nonce:     nonce.Nonce,
		ExpiresAt: nonce.ExpiresAt.Format(time.RFC3339),
	}, nil
}

func (s *AuthServer) VerifySiwe(ctx context.Context, req *pb.VerifySiweRequest) (*pb.VerifySiweResponse, error) {
	if req.AccountId == "" || req.Message == "" || req.Signature == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id, message and signature are required")
	}

	// Parse SIWE message
	siweMsg, err := s.siweService.ParseMessage(req.Message)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid SIWE message format")
	}

	address := siweMsg.GetAddress().Hex()
	chainIDStr := fmt.Sprintf("%d", siweMsg.GetChainID())
	nonce := siweMsg.GetNonce()
	domain := siweMsg.GetDomain()

	// Verify signature
	if err := s.siweService.VerifySignature(req.Message, req.Signature, address); err != nil {
		return nil, status.Error(codes.Unauthenticated, "signature verification failed")
	}

	// Validate and consume nonce using accountId from request
	if err := s.nonceRepo.ValidateAndConsumeNonce(ctx, nonce, strings.ToLower(req.AccountId), chainIDStr, domain); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired nonce")
	}

	// Ensure user exists (call user-service)
	userResp, err := s.clients.UserClient.EnsureUser(ctx, &pb.EnsureUserRequest{
		AccountId: strings.ToLower(req.AccountId),
		Address:   strings.ToLower(address),
		ChainId:   chainIDStr,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to ensure user: %v", err)
	}

	userID, _ := uuid.Parse(userResp.UserId)

	// Link wallet (call wallet-service)
	_, err = s.clients.WalletClient.UpsertLink(ctx, &pb.UpsertLinkRequest{
		UserId:    userID.String(),
		AccountId: strings.ToLower(req.AccountId),
		Address:   strings.ToLower(address),
		ChainId:   chainIDStr,
		IsPrimary: true,
		Type:      "eoa",
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to link wallet: %v", err)
	}

	// Create session
	session := &models.Session{
		UserID:          userID,
		TokenFamilyID:   uuid.New(),
		TokenGeneration: 1,
		CreatedAt:       time.Now(),
		ExpiresAt:       time.Now().Add(7 * 24 * time.Hour), // 7 days
		LastUsedAt:      time.Now(),
	}

	// Generate token pair
	tokens, err := s.jwtService.GenerateTokenPair(userID, session.SessionID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	// Hash and store refresh token
	session.RefreshHash = repository.HashRefreshToken(tokens.RefreshToken)

	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %v", err)
	}

	return &pb.VerifySiweResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		UserId:       userID.String(),
		Address:      strings.ToLower(address),
		ChainId:      chainIDStr,
		ExpiresAt:    tokens.ExpiresAt.Format(time.RFC3339),
	}, nil
}

func (s *AuthServer) RefreshSession(ctx context.Context, req *pb.RefreshSessionRequest) (*pb.RefreshSessionResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	refreshHash := repository.HashRefreshToken(req.RefreshToken)

	// Get session by refresh token
	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshHash)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	// Check if session is active
	if !session.IsActive() {
		return nil, status.Error(codes.Unauthenticated, "session expired or revoked")
	}

	// Generate new token pair
	tokens, err := s.jwtService.GenerateTokenPair(session.UserID, session.SessionID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	// Update session with new refresh token (token rotation)
	oldRefreshHash := session.RefreshHash
	session.RefreshHash = repository.HashRefreshToken(tokens.RefreshToken)
	session.PreviousRefreshHash = &oldRefreshHash
	session.TokenGeneration++
	session.LastUsedAt = time.Now()

	if err := s.sessionRepo.UpdateSession(ctx, session); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update session: %v", err)
	}

	return &pb.RefreshSessionResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt.Format(time.RFC3339),
	}, nil
}
