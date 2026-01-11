package server

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/client"
	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/models"
	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/repository"
	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/service"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	nonceRepo      repository.NonceRepository
	sessionRepo    repository.SessionRepository
	loginEventRepo repository.LoginEventRepository
	siweService    *service.SIWEService
	jwtService     *service.JWTService
	clients        *client.ServiceClients
}

func NewAuthServer(
	nonceRepo repository.NonceRepository,
	sessionRepo repository.SessionRepository,
	loginEventRepo repository.LoginEventRepository,
	siweService *service.SIWEService,
	jwtService *service.JWTService,
	clients *client.ServiceClients,
) *AuthServer {
	return &AuthServer{
		nonceRepo:      nonceRepo,
		sessionRepo:    sessionRepo,
		loginEventRepo: loginEventRepo,
		siweService:    siweService,
		jwtService:     jwtService,
		clients:        clients,
	}
}

// Helper functions to extract metadata from gRPC context
func getIPAddress(ctx context.Context) *string {
	if p, ok := peer.FromContext(ctx); ok {
		addr := p.Addr.String()
		// Use net.SplitHostPort to properly handle both IPv4 and IPv6
		// (e.g., "10.1.2.145:33408" -> "10.1.2.145", "[::1]:33408" -> "::1")
		if host, _, err := net.SplitHostPort(addr); err == nil {
			return &host
		}
		// If SplitHostPort fails, return as-is (might be IP without port)
		return &addr
	}
	return nil
}

func getUserAgent(ctx context.Context) *string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ua := md.Get("user-agent"); len(ua) > 0 {
			return &ua[0]
		}
	}
	return nil
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
	// Extract metadata for audit logging
	ipAddress := getIPAddress(ctx)
	userAgent := getUserAgent(ctx)
	accountID := strings.ToLower(req.AccountId)

	// Helper function to log login events
	logLoginEvent := func(result models.LoginResult, userID *uuid.UUID, chainID, domain *string, errMsg *string) {
		event := &models.LoginEvent{
			UserID:       userID,
			AccountID:    accountID,
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
			Result:       result,
			ErrorMessage: errMsg,
			ChainID:      chainID,
			Domain:       domain,
			Timestamp:    time.Now(),
		}
		// Fire and forget - don't block on logging errors
		go func() {
			_ = s.loginEventRepo.CreateLoginEvent(context.Background(), event)
		}()
	}

	if req.AccountId == "" || req.Message == "" || req.Signature == "" {
		errMsg := "account_id, message and signature are required"
		logLoginEvent(models.LoginResultFailed, nil, nil, nil, &errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	// Parse SIWE message
	siweMsg, err := s.siweService.ParseMessage(req.Message)
	if err != nil {
		errMsg := fmt.Sprintf("invalid SIWE message format: %v", err)
		logLoginEvent(models.LoginResultInvalidMessage, nil, nil, nil, &errMsg)
		return nil, status.Error(codes.InvalidArgument, "invalid SIWE message format")
	}

	address := siweMsg.GetAddress().Hex()
	chainIDStr := fmt.Sprintf("eip155:%d", siweMsg.GetChainID())
	nonce := siweMsg.GetNonce()
	domain := siweMsg.GetDomain()

	// Verify signature
	if err := s.siweService.VerifySignature(req.Message, req.Signature, address); err != nil {
		errMsg := fmt.Sprintf("signature verification failed: %v", err)
		logLoginEvent(models.LoginResultInvalidSignature, nil, &chainIDStr, &domain, &errMsg)
		return nil, status.Error(codes.Unauthenticated, "signature verification failed")
	}

	// Validate and consume nonce using accountId from request
	if err := s.nonceRepo.ValidateAndConsumeNonce(ctx, nonce, accountID, chainIDStr, domain); err != nil {
		var result models.LoginResult
		errMsg := err.Error()

		// Determine specific error type
		switch {
		case strings.Contains(errMsg, "expired"):
			result = models.LoginResultExpiredNonce
		case strings.Contains(errMsg, "used"):
			result = models.LoginResultInvalidNonce
		default:
			result = models.LoginResultInvalidNonce
		}

		logLoginEvent(result, nil, &chainIDStr, &domain, &errMsg)
		return nil, status.Error(codes.Unauthenticated, "invalid or expired nonce")
	}

	// Ensure user exists (call user-service)
	userResp, err := s.clients.UserClient.EnsureUser(ctx, &pb.EnsureUserRequest{
		AccountId: accountID,
		Address:   strings.ToLower(address),
		ChainId:   chainIDStr,
	})
	if err != nil {
		errMsg := fmt.Sprintf("failed to ensure user: %v", err)
		logLoginEvent(models.LoginResultFailed, nil, &chainIDStr, &domain, &errMsg)
		return nil, status.Errorf(codes.Internal, "failed to ensure user: %v", err)
	}

	userID, _ := uuid.Parse(userResp.UserId)

	// Link wallet (call wallet-service)
	_, err = s.clients.WalletClient.UpsertLink(ctx, &pb.UpsertLinkRequest{
		UserId:    userID.String(),
		AccountId: accountID,
		Address:   strings.ToLower(address),
		ChainId:   chainIDStr,
		IsPrimary: true,
		Type:      "eoa",
	})
	if err != nil {
		errMsg := fmt.Sprintf("failed to link wallet: %v", err)
		logLoginEvent(models.LoginResultFailed, &userID, &chainIDStr, &domain, &errMsg)
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
		errMsg := fmt.Sprintf("failed to generate tokens: %v", err)
		logLoginEvent(models.LoginResultFailed, &userID, &chainIDStr, &domain, &errMsg)
		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	// Hash and store refresh token
	session.RefreshHash = repository.HashRefreshToken(tokens.RefreshToken)

	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		errMsg := fmt.Sprintf("failed to create session: %v", err)
		logLoginEvent(models.LoginResultFailed, &userID, &chainIDStr, &domain, &errMsg)
		return nil, status.Errorf(codes.Internal, "failed to create session: %v", err)
	}

	// Log successful login
	logLoginEvent(models.LoginResultSuccess, &userID, &chainIDStr, &domain, nil)

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
		UserId:       session.UserID.String(),
	}, nil
}

func (s *AuthServer) RevokeSession(ctx context.Context, req *pb.RevokeSessionRequest) (*pb.RevokeSessionResponse, error) {
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session_id format")
	}

	err = s.sessionRepo.RevokeSession(ctx, sessionID, "user_requested")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to revoke session: %v", err)
	}

	return &pb.RevokeSessionResponse{
		Success: true,
	}, nil
}

func (s *AuthServer) RevokeSessionByRefreshToken(ctx context.Context, req *pb.RevokeSessionByRefreshTokenRequest) (*pb.RevokeSessionByRefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	refreshHash := repository.HashRefreshToken(req.RefreshToken)

	err := s.sessionRepo.RevokeByRefreshToken(ctx, refreshHash, "user_logout")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to revoke session: %v", err)
	}

	return &pb.RevokeSessionByRefreshTokenResponse{
		Success: true,
	}, nil
}
