package tracing

import (
	"testing"
)

func TestGetTracesSampleRate(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    float64
	}{
		{
			name:        "production returns 5%",
			environment: "production",
			expected:    0.05,
		},
		{
			name:        "staging returns 20%",
			environment: "staging",
			expected:    0.20,
		},
		{
			name:        "development returns 100%",
			environment: "development",
			expected:    1.0,
		},
		{
			name:        "unknown environment returns 100%",
			environment: "unknown",
			expected:    1.0,
		},
		{
			name:        "empty environment returns 100%",
			environment: "",
			expected:    1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTracesSampleRate(tt.environment)
			if result != tt.expected {
				t.Errorf("GetTracesSampleRate(%q) = %v, want %v", tt.environment, result, tt.expected)
			}
		})
	}
}

func TestIsAuthOperation(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    bool
	}{
		{
			name:        "VerifySIWE is auth operation",
			description: "/proto.AuthService/VerifySIWE",
			expected:    true,
		},
		{
			name:        "RefreshToken is auth operation",
			description: "RefreshToken",
			expected:    true,
		},
		{
			name:        "Login is auth operation",
			description: "UserLogin",
			expected:    true,
		},
		{
			name:        "Logout is auth operation",
			description: "UserLogout",
			expected:    true,
		},
		{
			name:        "Authenticate is auth operation",
			description: "AuthenticateUser",
			expected:    true,
		},
		{
			name:        "signIn is auth operation",
			description: "signInWithWallet",
			expected:    true,
		},
		{
			name:        "signOut is auth operation",
			description: "signOutUser",
			expected:    true,
		},
		{
			name:        "GetUser is not auth operation",
			description: "/proto.UserService/GetUser",
			expected:    false,
		},
		{
			name:        "GetWallet is not auth operation",
			description: "GetWallet",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAuthOperation(tt.description)
			if result != tt.expected {
				t.Errorf("isAuthOperation(%q) = %v, want %v", tt.description, result, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "substring at start",
			s:        "VerifySIWE",
			substr:   "Verify",
			expected: true,
		},
		{
			name:     "substring at end",
			s:        "UserLogin",
			substr:   "Login",
			expected: true,
		},
		{
			name:     "substring in middle",
			s:        "UserLoginMethod",
			substr:   "Login",
			expected: true,
		},
		{
			name:     "exact match",
			s:        "Login",
			substr:   "Login",
			expected: true,
		},
		{
			name:     "no match",
			s:        "GetUser",
			substr:   "Login",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "GetUser",
			substr:   "",
			expected: true,
		},
		{
			name:     "substring longer than string",
			s:        "Get",
			substr:   "GetUser",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}
