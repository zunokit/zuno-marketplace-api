package cookie

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestSetRefreshTokenCookie_Development(t *testing.T) {
	// Set environment to development
	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")

	w := httptest.NewRecorder()
	token := "test-refresh-token-123"
	maxAge := 86400 // 1 day

	SetRefreshTokenCookie(w, token, maxAge)

	// Get the Set-Cookie header
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]

	// Verify cookie properties
	if cookie.Name != RefreshTokenCookie {
		t.Errorf("Expected cookie name %s, got %s", RefreshTokenCookie, cookie.Name)
	}

	if cookie.Value != token {
		t.Errorf("Expected cookie value %s, got %s", token, cookie.Value)
	}

	if cookie.Path != "/" {
		t.Errorf("Expected path /, got %s", cookie.Path)
	}

	if cookie.MaxAge != maxAge {
		t.Errorf("Expected MaxAge %d, got %d", maxAge, cookie.MaxAge)
	}

	if !cookie.HttpOnly {
		t.Error("Expected HttpOnly to be true")
	}

	// In development, Secure should be false
	if cookie.Secure {
		t.Error("Expected Secure to be false in development")
	}

	if cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("Expected SameSite Strict, got %v", cookie.SameSite)
	}
}

func TestSetRefreshTokenCookie_Production(t *testing.T) {
	// Set environment to production
	os.Setenv("ENVIRONMENT", "production")
	defer os.Unsetenv("ENVIRONMENT")

	w := httptest.NewRecorder()
	token := "test-refresh-token-456"
	maxAge := 604800 // 7 days

	SetRefreshTokenCookie(w, token, maxAge)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]

	// In production, Secure should be true
	if !cookie.Secure {
		t.Error("Expected Secure to be true in production")
	}

	if !cookie.HttpOnly {
		t.Error("Expected HttpOnly to be true")
	}

	if cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("Expected SameSite Strict, got %v", cookie.SameSite)
	}
}

func TestSetRefreshTokenCookie_DefaultMaxAge(t *testing.T) {
	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")

	w := httptest.NewRecorder()
	token := "test-token"
	maxAge := MaxAge // Use the constant

	SetRefreshTokenCookie(w, token, maxAge)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]

	expectedMaxAge := 30 * 24 * 60 * 60 // 30 days in seconds
	if cookie.MaxAge != expectedMaxAge {
		t.Errorf("Expected MaxAge %d (30 days), got %d", expectedMaxAge, cookie.MaxAge)
	}
}

func TestGetRefreshTokenFromCookie_Success(t *testing.T) {
	expectedToken := "my-refresh-token"

	// Create a request with the cookie
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.AddCookie(&http.Cookie{
		Name:  RefreshTokenCookie,
		Value: expectedToken,
	})

	// Get the token
	token, err := GetRefreshTokenFromCookie(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token != expectedToken {
		t.Errorf("Expected token %s, got %s", expectedToken, token)
	}
}

func TestGetRefreshTokenFromCookie_NotFound(t *testing.T) {
	// Create a request without the cookie
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	// Try to get the token
	_, err := GetRefreshTokenFromCookie(req)
	if err == nil {
		t.Error("Expected error when cookie not found")
	}

	if err != http.ErrNoCookie {
		t.Errorf("Expected ErrNoCookie, got %v", err)
	}
}

func TestGetRefreshTokenFromCookie_EmptyValue(t *testing.T) {
	// Create a request with empty cookie value
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.AddCookie(&http.Cookie{
		Name:  RefreshTokenCookie,
		Value: "",
	})

	// Get the token (should succeed but return empty string)
	token, err := GetRefreshTokenFromCookie(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token != "" {
		t.Errorf("Expected empty token, got %s", token)
	}
}

func TestClearRefreshTokenCookie_Development(t *testing.T) {
	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")

	w := httptest.NewRecorder()
	ClearRefreshTokenCookie(w)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]

	// Verify cookie is being deleted
	if cookie.Name != RefreshTokenCookie {
		t.Errorf("Expected cookie name %s, got %s", RefreshTokenCookie, cookie.Name)
	}

	if cookie.Value != "" {
		t.Errorf("Expected empty value, got %s", cookie.Value)
	}

	if cookie.MaxAge != -1 {
		t.Errorf("Expected MaxAge -1 (delete), got %d", cookie.MaxAge)
	}

	if cookie.Path != "/" {
		t.Errorf("Expected path /, got %s", cookie.Path)
	}

	if !cookie.HttpOnly {
		t.Error("Expected HttpOnly to be true")
	}

	// In development, Secure should be false
	if cookie.Secure {
		t.Error("Expected Secure to be false in development")
	}

	if cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("Expected SameSite Strict, got %v", cookie.SameSite)
	}

	// Verify Expires is set to Unix epoch
	if !cookie.Expires.IsZero() && cookie.Expires.Unix() != 0 {
		t.Errorf("Expected Expires to be Unix epoch, got %v", cookie.Expires)
	}
}

func TestClearRefreshTokenCookie_Production(t *testing.T) {
	os.Setenv("ENVIRONMENT", "production")
	defer os.Unsetenv("ENVIRONMENT")

	w := httptest.NewRecorder()
	ClearRefreshTokenCookie(w)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]

	// In production, Secure should be true even when clearing
	if !cookie.Secure {
		t.Error("Expected Secure to be true in production")
	}

	if cookie.MaxAge != -1 {
		t.Errorf("Expected MaxAge -1 (delete), got %d", cookie.MaxAge)
	}
}

func TestCookieConstants(t *testing.T) {
	expectedCookieName := "refresh_token"
	if RefreshTokenCookie != expectedCookieName {
		t.Errorf("Expected RefreshTokenCookie to be %s, got %s", expectedCookieName, RefreshTokenCookie)
	}

	expectedMaxAge := 30 * 24 * 60 * 60 // 30 days
	if MaxAge != expectedMaxAge {
		t.Errorf("Expected MaxAge to be %d (30 days), got %d", expectedMaxAge, MaxAge)
	}
}

func TestSetAndGetRefreshToken_RoundTrip(t *testing.T) {
	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")

	expectedToken := "round-trip-token-789"

	// Set cookie
	w := httptest.NewRecorder()
	SetRefreshTokenCookie(w, expectedToken, MaxAge)

	// Create a request with the cookie from the response
	cookies := w.Result().Cookies()
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// Get token back
	token, err := GetRefreshTokenFromCookie(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token != expectedToken {
		t.Errorf("Round trip failed: expected %s, got %s", expectedToken, token)
	}
}

func TestMultipleCookies_GetRefreshToken(t *testing.T) {
	expectedToken := "correct-token"

	// Create a request with multiple cookies
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.AddCookie(&http.Cookie{
		Name:  "other_cookie",
		Value: "other-value",
	})
	req.AddCookie(&http.Cookie{
		Name:  RefreshTokenCookie,
		Value: expectedToken,
	})
	req.AddCookie(&http.Cookie{
		Name:  "another_cookie",
		Value: "another-value",
	})

	// Get the refresh token
	token, err := GetRefreshTokenFromCookie(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token != expectedToken {
		t.Errorf("Expected token %s, got %s", expectedToken, token)
	}
}

func TestSecurityAttributes(t *testing.T) {
	os.Setenv("ENVIRONMENT", "production")
	defer os.Unsetenv("ENVIRONMENT")

	w := httptest.NewRecorder()
	SetRefreshTokenCookie(w, "secure-token", MaxAge)

	cookies := w.Result().Cookies()
	cookie := cookies[0]

	// Verify all security attributes
	securityChecks := []struct {
		name      string
		condition bool
		message   string
	}{
		{"HttpOnly", cookie.HttpOnly, "HttpOnly must be true to prevent XSS"},
		{"Secure", cookie.Secure, "Secure must be true in production for HTTPS"},
		{"SameSite", cookie.SameSite == http.SameSiteStrictMode, "SameSite must be Strict for CSRF protection"},
		{"Path", cookie.Path == "/", "Path should be / for site-wide access"},
	}

	for _, check := range securityChecks {
		if !check.condition {
			t.Errorf("Security check failed for %s: %s", check.name, check.message)
		}
	}
}

func TestSetRefreshTokenCookie_SpecialCharacters(t *testing.T) {
	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")

	// Test with token containing special characters
	token := "token.with-special_chars123"

	w := httptest.NewRecorder()
	SetRefreshTokenCookie(w, token, MaxAge)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	if cookies[0].Value != token {
		t.Errorf("Expected token %s, got %s", token, cookies[0].Value)
	}
}

func TestClearRefreshTokenCookie_Idempotent(t *testing.T) {
	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")

	// Clear cookie multiple times
	w1 := httptest.NewRecorder()
	ClearRefreshTokenCookie(w1)

	w2 := httptest.NewRecorder()
	ClearRefreshTokenCookie(w2)

	// Both should produce the same result
	cookies1 := w1.Result().Cookies()
	cookies2 := w2.Result().Cookies()

	if len(cookies1) != 1 || len(cookies2) != 1 {
		t.Fatal("Expected 1 cookie in each response")
	}

	if cookies1[0].MaxAge != cookies2[0].MaxAge {
		t.Error("Expected idempotent behavior for ClearRefreshTokenCookie")
	}
}
