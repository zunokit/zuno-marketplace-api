package sentry

import (
	"testing"

	"github.com/getsentry/sentry-go"
)

func TestScrubString_EthereumAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid eth address",
			input:    "Wallet: 0x1234567890123456789012345678901234567890",
			expected: "Wallet: [FILTERED:ETH_ADDRESS]",
		},
		{
			name:     "multiple eth addresses",
			input:    "From: 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa To: 0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			expected: "From: [FILTERED:ETH_ADDRESS] To: [FILTERED:ETH_ADDRESS]",
		},
		{
			name:     "invalid eth address - too short",
			input:    "0x12345",
			expected: "0x12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scrubString(tt.input)
			if result != tt.expected {
				t.Errorf("scrubString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestScrubString_JWT(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid jwt token",
			input:    "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			expected: "Bearer [FILTERED:JWT]",
		},
		{
			name:     "jwt without signature",
			input:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0",
			expected: "[FILTERED:JWT]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scrubString(tt.input)
			if result != tt.expected {
				t.Errorf("scrubString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestScrubString_Email(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid email",
			input:    "Contact: user@example.com",
			expected: "Contact: [FILTERED:EMAIL]",
		},
		{
			name:     "email with subdomain",
			input:    "admin@mail.example.org",
			expected: "[FILTERED:EMAIL]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scrubString(tt.input)
			if result != tt.expected {
				t.Errorf("scrubString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestScrubString_PrivateKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "64 char hex private key",
			input:    "Private key: ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
			expected: "Private key: [FILTERED:PRIVATE_KEY]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scrubString(tt.input)
			if result != tt.expected {
				t.Errorf("scrubString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestScrubString_CAIP10(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "eip155 account id",
			input:    "Account: eip155:1:0xab16a96d359ec26a11e2c2b3d8f8b8942d5bfcdb",
			expected: "Account: [FILTERED:CAIP10]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scrubString(tt.input)
			if result != tt.expected {
				t.Errorf("scrubString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestScrubMap_Nested(t *testing.T) {
	data := map[string]interface{}{
		"wallet": "0x1234567890123456789012345678901234567890",
		"user": map[string]interface{}{
			"email": "user@example.com",
			"meta": map[string]interface{}{
				"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0",
			},
		},
	}

	result := scrubMap(data)

	if result["wallet"] != "[FILTERED:ETH_ADDRESS]" {
		t.Errorf("wallet not scrubbed: %v", result["wallet"])
	}

	user, ok := result["user"].(map[string]interface{})
	if !ok {
		t.Fatal("user map not found")
	}

	if user["email"] != "[FILTERED:EMAIL]" {
		t.Errorf("email not scrubbed: %v", user["email"])
	}

	meta, ok := user["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("meta map not found")
	}

	if meta["token"] != "[FILTERED:JWT]" {
		t.Errorf("token not scrubbed: %v", meta["token"])
	}
}

func TestScrubSlice(t *testing.T) {
	data := []interface{}{
		"0x1234567890123456789012345678901234567890",
		"user@example.com",
		"normal text",
	}

	result := scrubSlice(data)

	if result[0] != "[FILTERED:ETH_ADDRESS]" {
		t.Errorf("element 0 not scrubbed: %v", result[0])
	}

	if result[1] != "[FILTERED:EMAIL]" {
		t.Errorf("element 1 not scrubbed: %v", result[1])
	}

	if result[2] != "normal text" {
		t.Errorf("element 2 should not be scrubbed: %v", result[2])
	}
}

func TestScrubHeaders(t *testing.T) {
	headers := map[string]string{
		"Content-Type":   "application/json",
		"Authorization":  "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		"X-API-Key":      "secret-key-123",
		"Cookie":         "session=abc123",
		"User-Agent":     "Mozilla/5.0",
		"Wallet-Address": "0x1234567890123456789012345678901234567890",
	}

	result := scrubHeaders(headers)

	if result["Authorization"] != "[FILTERED]" {
		t.Errorf("Authorization should be filtered: %v", result["Authorization"])
	}

	if result["X-API-Key"] != "[FILTERED]" {
		t.Errorf("X-API-Key should be filtered: %v", result["X-API-Key"])
	}

	if result["Cookie"] != "[FILTERED]" {
		t.Errorf("Cookie should be filtered: %v", result["Cookie"])
	}

	if result["Content-Type"] != "application/json" {
		t.Errorf("Content-Type should not be filtered: %v", result["Content-Type"])
	}

	if result["User-Agent"] != "Mozilla/5.0" {
		t.Errorf("User-Agent should not be filtered: %v", result["User-Agent"])
	}

	if result["Wallet-Address"] != "[FILTERED:ETH_ADDRESS]" {
		t.Errorf("Wallet-Address should be scrubbed: %v", result["Wallet-Address"])
	}
}

func TestScrubUser(t *testing.T) {
	user := &sentry.User{
		ID:        "user-123",
		Email:     "user@example.com",
		IPAddress: "192.168.1.1",
		Data: map[string]string{
			"wallet": "0x1234567890123456789012345678901234567890",
		},
	}

	result := scrubUser(user)

	if result.Email != "[FILTERED]" {
		t.Errorf("Email should be filtered: %v", result.Email)
	}

	if result.IPAddress != "[FILTERED]" {
		t.Errorf("IPAddress should be filtered: %v", result.IPAddress)
	}

	if result.ID != "user-123" {
		t.Errorf("ID should not be filtered: %v", result.ID)
	}

	if result.Data["wallet"] != "[FILTERED:ETH_ADDRESS]" {
		t.Errorf("wallet in Data should be scrubbed: %v", result.Data["wallet"])
	}
}

func TestScrubEvent(t *testing.T) {
	event := &sentry.Event{
		Message: "Test error",
		Request: &sentry.Request{
			Headers: map[string]string{
				"Authorization": "Bearer token123",
			},
			Data: `{"password": "secret123", "wallet": "0x1234567890123456789012345678901234567890"}`,
		},
		Extra: map[string]interface{}{
			"wallet": "0x1234567890123456789012345678901234567890",
		},
	}

	result := scrubEvent(event, nil)

	if result.Request.Headers["Authorization"] != "[FILTERED]" {
		t.Errorf("Authorization header should be filtered")
	}

	if !contains(result.Request.Data, "[FILTERED:ETH_ADDRESS]") {
		t.Errorf("Request data should be scrubbed")
	}

	if result.Extra["wallet"] != "[FILTERED:ETH_ADDRESS]" {
		t.Errorf("Extra wallet should be scrubbed")
	}
}

func TestMultiplePatterns(t *testing.T) {
	input := "User: user@example.com logged in from 0x1234567890123456789012345678901234567890 with token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
	result := scrubString(input)

	// Check that all patterns are scrubbed
	if !contains(result, "[FILTERED:EMAIL]") {
		t.Error("Email should be filtered")
	}
	if !contains(result, "[FILTERED:ETH_ADDRESS]") {
		t.Error("ETH address should be filtered")
	}
	if !contains(result, "[FILTERED:JWT]") {
		t.Error("JWT should be filtered")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
