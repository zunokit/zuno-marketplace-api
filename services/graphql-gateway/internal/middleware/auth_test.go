package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testJWTSecret = "test-secret-key-min-32-characters-required-here"

func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Generate a valid token
	userID := uuid.New().String()
	sessionID := uuid.New().String()
	token := generateTestToken(t, userID, sessionID, time.Now().Add(1*time.Hour), testJWTSecret)

	// Create middleware
	middleware := AuthMiddleware(testJWTSecret)

	// Create test handler
	var capturedClaims *UserClaims
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetUserClaims(r.Context())
		if ok {
			capturedClaims = claims
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap handler with middleware
	wrappedHandler := middleware(handler)

	// Create request with Authorization header
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w := httptest.NewRecorder()

	// Execute request
	wrappedHandler.ServeHTTP(w, req)

	// Verify claims were added to context
	if capturedClaims == nil {
		t.Fatal("Expected claims to be added to context")
	}

	if capturedClaims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, capturedClaims.UserID)
	}

	if capturedClaims.SessionID != sessionID {
		t.Errorf("Expected SessionID %s, got %s", sessionID, capturedClaims.SessionID)
	}
}

func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	middleware := AuthMiddleware(testJWTSecret)

	var capturedClaims *UserClaims
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetUserClaims(r.Context())
		if ok {
			capturedClaims = claims
		}
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	// Request without Authorization header
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Should not have claims in context
	if capturedClaims != nil {
		t.Error("Expected no claims in context when Authorization header is missing")
	}

	// Should still return 200 (middleware doesn't reject, resolvers do)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
	}{
		{
			name:       "malformed token",
			authHeader: "Bearer invalid.token.here",
		},
		{
			name:       "not a JWT",
			authHeader: "Bearer notajwt",
		},
		{
			name:       "missing Bearer prefix",
			authHeader: "sometoken",
		},
		{
			name:       "wrong prefix",
			authHeader: "Basic sometoken",
		},
		{
			name:       "empty bearer",
			authHeader: "Bearer ",
		},
		{
			name:       "bearer with whitespace only",
			authHeader: "Bearer    ",
		},
		{
			name:       "bearer with tab",
			authHeader: "Bearer\t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := AuthMiddleware(testJWTSecret)

			var capturedClaims *UserClaims
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims, ok := GetUserClaims(r.Context())
				if ok {
					capturedClaims = claims
				}
				w.WriteHeader(http.StatusOK)
			})

			wrappedHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tt.authHeader)
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			// Should not have claims in context for invalid tokens
			if capturedClaims != nil {
				t.Error("Expected no claims in context for invalid token")
			}
		})
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	userID := uuid.New().String()
	sessionID := uuid.New().String()

	// Generate expired token (expired 1 hour ago)
	token := generateTestToken(t, userID, sessionID, time.Now().Add(-1*time.Hour), testJWTSecret)

	middleware := AuthMiddleware(testJWTSecret)

	var capturedClaims *UserClaims
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetUserClaims(r.Context())
		if ok {
			capturedClaims = claims
		}
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Should not have claims for expired token
	if capturedClaims != nil {
		t.Error("Expected no claims in context for expired token")
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	userID := uuid.New().String()
	sessionID := uuid.New().String()

	// Generate token with different secret
	token := generateTestToken(t, userID, sessionID, time.Now().Add(1*time.Hour), "wrong-secret-key")

	middleware := AuthMiddleware(testJWTSecret)

	var capturedClaims *UserClaims
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetUserClaims(r.Context())
		if ok {
			capturedClaims = claims
		}
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Should not have claims when signed with wrong secret
	if capturedClaims != nil {
		t.Error("Expected no claims in context for token signed with wrong secret")
	}
}

func TestRequireAuth_WithValidClaims(t *testing.T) {
	userID := uuid.New().String()
	sessionID := uuid.New().String()

	claims := &UserClaims{
		UserID:    userID,
		SessionID: sessionID,
	}

	ctx := context.WithValue(context.Background(), UserClaimsKey, claims)

	result, err := RequireAuth(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, result.UserID)
	}
}

func TestRequireAuth_WithoutClaims(t *testing.T) {
	ctx := context.Background()

	_, err := RequireAuth(ctx)
	if err == nil {
		t.Error("Expected error when no claims in context")
	}

	expectedErr := "authentication required"
	if err.Error() != expectedErr {
		t.Errorf("Expected error message '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestGetUserClaims_WithValidClaims(t *testing.T) {
	userID := uuid.New().String()
	claims := &UserClaims{UserID: userID}

	ctx := context.WithValue(context.Background(), UserClaimsKey, claims)

	result, ok := GetUserClaims(ctx)
	if !ok {
		t.Fatal("Expected claims to be found")
	}

	if result.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, result.UserID)
	}
}

func TestGetUserClaims_WithoutClaims(t *testing.T) {
	ctx := context.Background()

	_, ok := GetUserClaims(ctx)
	if ok {
		t.Error("Expected no claims to be found")
	}
}

func TestAuthMiddleware_CaseInsensitiveBearer(t *testing.T) {
	userID := uuid.New().String()
	sessionID := uuid.New().String()
	token := generateTestToken(t, userID, sessionID, time.Now().Add(1*time.Hour), testJWTSecret)

	tests := []struct {
		name       string
		authHeader string
		shouldWork bool
	}{
		{
			name:       "lowercase bearer",
			authHeader: fmt.Sprintf("bearer %s", token),
			shouldWork: true,
		},
		{
			name:       "uppercase BEARER",
			authHeader: fmt.Sprintf("BEARER %s", token),
			shouldWork: true,
		},
		{
			name:       "mixed case Bearer",
			authHeader: fmt.Sprintf("Bearer %s", token),
			shouldWork: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := AuthMiddleware(testJWTSecret)

			var capturedClaims *UserClaims
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims, ok := GetUserClaims(r.Context())
				if ok {
					capturedClaims = claims
				}
				w.WriteHeader(http.StatusOK)
			})

			wrappedHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tt.authHeader)
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			if tt.shouldWork && capturedClaims == nil {
				t.Error("Expected claims to be added to context")
			}
			if tt.shouldWork && capturedClaims != nil && capturedClaims.UserID != userID {
				t.Errorf("Expected UserID %s, got %s", userID, capturedClaims.UserID)
			}
		})
	}
}

// Helper function to generate test JWT tokens
func generateTestToken(t *testing.T, userID, sessionID string, expiresAt time.Time, secret string) string {
	claims := &UserClaims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	return tokenStr
}
