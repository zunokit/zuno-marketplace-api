package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTService_GenerateTokenPair(t *testing.T) {
	jwtService := NewJWTService("test-secret-32-chars-minimum!", "refresh-secret-32-chars-min!", 1*time.Hour, 24*time.Hour)

	userID := uuid.New()
	sessionID := uuid.New()

	tokens, err := jwtService.GenerateTokenPair(userID, sessionID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if tokens.AccessToken == "" {
		t.Error("Expected access token to be generated")
	}

	if tokens.RefreshToken == "" {
		t.Error("Expected refresh token to be generated")
	}

	if len(tokens.RefreshToken) != 64 {
		t.Errorf("Expected refresh token length 64, got %d", len(tokens.RefreshToken))
	}

	if tokens.ExpiresAt.Before(time.Now()) {
		t.Error("Expected expiration to be in the future")
	}
}

func TestJWTService_ValidateAccessToken(t *testing.T) {
	jwtService := NewJWTService("test-secret-32-chars-minimum!", "refresh-secret-32-chars-min!", 1*time.Hour, 24*time.Hour)

	userID := uuid.New()
	sessionID := uuid.New()

	tokens, err := jwtService.GenerateTokenPair(userID, sessionID)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	// Validate the access token
	claims, err := jwtService.ValidateAccessToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("Expected token to be valid, got error: %v", err)
	}

	if claims.UserID != userID.String() {
		t.Errorf("Expected user ID %s, got %s", userID.String(), claims.UserID)
	}

	if claims.SessionID != sessionID.String() {
		t.Errorf("Expected session ID %s, got %s", sessionID.String(), claims.SessionID)
	}

	// Test invalid token
	_, err = jwtService.ValidateAccessToken("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}
