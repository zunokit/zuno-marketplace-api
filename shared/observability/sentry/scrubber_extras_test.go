package sentry

import (
	"testing"

	sentrygo "github.com/getsentry/sentry-go"
)

func TestToString(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want string
	}{
		{"string passthrough", "hello", "hello"},
		{"empty string", "", ""},
		{"int returns empty", 42, ""},
		{"bool returns empty", true, ""},
		{"nil returns empty", nil, ""},
		{"map returns empty", map[string]string{"a": "b"}, ""},
		{"slice returns empty", []int{1, 2, 3}, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := toString(tc.in); got != tc.want {
				t.Errorf("toString(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestScrubTransaction_DelegatesToScrubEvent(t *testing.T) {
	tx := &sentrygo.Event{
		Type: "transaction",
		Extra: map[string]interface{}{
			"wallet": "0x1234567890123456789012345678901234567890",
		},
	}

	got := scrubTransaction(tx, nil)
	if got == nil {
		t.Fatal("scrubTransaction returned nil")
	}
	if wallet, ok := got.Extra["wallet"].(string); !ok || wallet == "0x1234567890123456789012345678901234567890" {
		t.Errorf("scrubTransaction should have redacted wallet, got %v", got.Extra["wallet"])
	}
}

func TestScrubMap_PreservesNonStringTypes(t *testing.T) {
	in := map[string]interface{}{
		"count":   42,
		"enabled": true,
		"email":   "alice@example.com",
		"nested": map[string]interface{}{
			"jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIn0.signature",
		},
		"slice": []interface{}{
			"0x1234567890123456789012345678901234567890",
			123,
		},
	}

	out := scrubMap(in)

	if out["count"] != 42 {
		t.Errorf("count should be preserved, got %v", out["count"])
	}
	if out["enabled"] != true {
		t.Errorf("enabled should be preserved, got %v", out["enabled"])
	}
	if email, _ := out["email"].(string); email == "alice@example.com" {
		t.Error("email should be redacted")
	}

	nested, ok := out["nested"].(map[string]interface{})
	if !ok {
		t.Fatalf("nested should remain a map, got %T", out["nested"])
	}
	if jwt, _ := nested["jwt"].(string); jwt == "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIn0.signature" {
		t.Error("nested jwt should be redacted")
	}

	slice, ok := out["slice"].([]interface{})
	if !ok {
		t.Fatalf("slice should remain a slice, got %T", out["slice"])
	}
	if len(slice) != 2 {
		t.Fatalf("slice length changed: got %d, want 2", len(slice))
	}
	if addr, _ := slice[0].(string); addr == "0x1234567890123456789012345678901234567890" {
		t.Error("slice eth address should be redacted")
	}
	if slice[1] != 123 {
		t.Errorf("slice numeric element should be preserved, got %v", slice[1])
	}
}
