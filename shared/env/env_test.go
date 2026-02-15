package env

import (
	"os"
	"testing"
)

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		fallback string
		want     string
		setup    bool
	}{
		{
			name:     "returns environment variable when set",
			key:      "TEST_STRING_VAR",
			value:    "test_value",
			fallback: "fallback_value",
			want:     "test_value",
			setup:    true,
		},
		{
			name:     "returns fallback when environment variable not set",
			key:      "NONEXISTENT_STRING_VAR",
			fallback: "fallback_value",
			want:     "fallback_value",
			setup:    false,
		},
		{
			name:     "returns empty string when set to empty",
			key:      "EMPTY_STRING_VAR",
			value:    "",
			fallback: "fallback_value",
			want:     "",
			setup:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setup {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			// Execute
			got := GetString(tt.key, tt.fallback)

			// Assert
			if got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		fallback int
		want     int
		setup    bool
	}{
		{
			name:     "returns integer when valid",
			key:      "TEST_INT_VAR",
			value:    "42",
			fallback: 10,
			want:     42,
			setup:    true,
		},
		{
			name:     "returns fallback when not set",
			key:      "NONEXISTENT_INT_VAR",
			fallback: 10,
			want:     10,
			setup:    false,
		},
		{
			name:     "returns fallback when invalid integer",
			key:      "INVALID_INT_VAR",
			value:    "not_a_number",
			fallback: 10,
			want:     10,
			setup:    true,
		},
		{
			name:     "returns negative integer when valid",
			key:      "NEGATIVE_INT_VAR",
			value:    "-100",
			fallback: 10,
			want:     -100,
			setup:    true,
		},
		{
			name:     "returns zero when set to zero",
			key:      "ZERO_INT_VAR",
			value:    "0",
			fallback: 10,
			want:     0,
			setup:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setup {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			// Execute
			got := GetInt(tt.key, tt.fallback)

			// Assert
			if got != tt.want {
				t.Errorf("GetInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		fallback bool
		want     bool
		setup    bool
	}{
		{
			name:     "returns true when set to 'true'",
			key:      "TEST_BOOL_VAR",
			value:    "true",
			fallback: false,
			want:     true,
			setup:    true,
		},
		{
			name:     "returns false when set to 'false'",
			key:      "TEST_BOOL_VAR",
			value:    "false",
			fallback: true,
			want:     false,
			setup:    true,
		},
		{
			name:     "returns true when set to '1'",
			key:      "TEST_BOOL_VAR",
			value:    "1",
			fallback: false,
			want:     true,
			setup:    true,
		},
		{
			name:     "returns false when set to '0'",
			key:      "TEST_BOOL_VAR",
			value:    "0",
			fallback: true,
			want:     false,
			setup:    true,
		},
		{
			name:     "returns true when set to 't'",
			key:      "TEST_BOOL_VAR",
			value:    "t",
			fallback: false,
			want:     true,
			setup:    true,
		},
		{
			name:     "returns false when set to 'f'",
			key:      "TEST_BOOL_VAR",
			value:    "f",
			fallback: true,
			want:     false,
			setup:    true,
		},
		{
			name:     "returns fallback when not set",
			key:      "NONEXISTENT_BOOL_VAR",
			fallback: true,
			want:     true,
			setup:    false,
		},
		{
			name:     "returns fallback when invalid boolean",
			key:      "INVALID_BOOL_VAR",
			value:    "not_a_bool",
			fallback: true,
			want:     true,
			setup:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setup {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			// Execute
			got := GetBool(tt.key, tt.fallback)

			// Assert
			if got != tt.want {
				t.Errorf("GetBool() = %v, want %v", got, tt.want)
			}
		})
	}
}
