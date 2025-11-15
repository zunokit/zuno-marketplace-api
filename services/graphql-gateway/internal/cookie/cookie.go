package cookie

import (
	"net/http"
	"time"
)

const (
	RefreshTokenCookie = "refresh_token"
	MaxAge             = 30 * 24 * 60 * 60 // 30 days in seconds
)

// SetRefreshTokenCookie sets an HTTP-only secure cookie for the refresh token
func SetRefreshTokenCookie(w http.ResponseWriter, token string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,  // Prevents JavaScript access
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
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
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete immediately
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	})
}
