package cookie

import (
	"net/http"
	"os"
	"time"
)

const (
	RefreshTokenCookie = "refresh_token"
	MaxAge             = 30 * 24 * 60 * 60 // 30 days in seconds
)

// SetRefreshTokenCookie sets an HTTP-only secure cookie for the refresh token
func SetRefreshTokenCookie(w http.ResponseWriter, token string, maxAge int) {
	// Use Secure cookies in production (HTTPS), allow HTTP in development
	secure := os.Getenv("ENVIRONMENT") != "development"

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,                    // Prevents JavaScript access - CRITICAL for XSS protection
		Secure:   secure,                  // true in production (HTTPS only), false in dev for localhost
		SameSite: http.SameSiteStrictMode, // Strict CSRF protection - only same-site requests
	})
}

// GetRefreshTokenFromCookie extracts the refresh token from the request cookie
func GetRefreshTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(RefreshTokenCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// ClearRefreshTokenCookie removes the refresh token cookie
func ClearRefreshTokenCookie(w http.ResponseWriter) {
	// Use same security settings as SetRefreshTokenCookie for consistency
	secure := os.Getenv("ENVIRONMENT") != "development"

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete immediately
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
	})
}
