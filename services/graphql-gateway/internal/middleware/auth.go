package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserClaimsKey contextKey = "user_claims"
)

// UserClaims represents the JWT claims for an authenticated user
type UserClaims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT tokens from the Authorization header
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				// Expected format: "Bearer <token>"
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenStr := parts[1]

					// Validate the token
					claims, err := validateAccessToken(tokenStr, jwtSecret)
					if err == nil {
						// Add claims to context
						ctx = context.WithValue(ctx, UserClaimsKey, claims)
					}
					// Note: We don't return an error here because not all GraphQL queries require auth
					// Individual resolvers will check if auth is required
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateAccessToken validates a JWT access token and returns the claims
func validateAccessToken(tokenStr string, jwtSecret string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetUserClaims retrieves the user claims from context
func GetUserClaims(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*UserClaims)
	return claims, ok
}

// RequireAuth returns an error if the user is not authenticated
func RequireAuth(ctx context.Context) (*UserClaims, error) {
	claims, ok := GetUserClaims(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}
	return claims, nil
}
