package graph

import (
	"testing"
	"time"

	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStringPtrToValue(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "nil pointer",
			input:    nil,
			expected: "",
		},
		{
			name:     "non-nil pointer with value",
			input:    strPtr("hello"),
			expected: "hello",
		},
		{
			name:     "non-nil pointer with empty string",
			input:    strPtr(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringPtrToValue(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValueToStringPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "non-empty string",
			input:    "hello",
			expected: strPtr("hello"),
		},
		{
			name:     "whitespace string",
			input:    "   ",
			expected: strPtr("   "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valueToStringPtr(tt.input)

			// Check if both are nil
			if tt.expected == nil && result == nil {
				return
			}

			// Check if one is nil and the other isn't
			if (tt.expected == nil) != (result == nil) {
				t.Errorf("Expected nil=%v, got nil=%v", tt.expected == nil, result == nil)
				return
			}

			// Both are non-nil, compare values
			if *tt.expected != *result {
				t.Errorf("Expected %q, got %q", *tt.expected, *result)
			}
		})
	}
}

func TestProfileFromProto_Nil(t *testing.T) {
	result := profileFromProto(nil)
	if result != nil {
		t.Errorf("Expected nil result for nil input, got %v", result)
	}
}

func TestProfileFromProto_Complete(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	username := "johndoe"
	displayName := "John Doe"
	avatarURL := "https://example.com/avatar.jpg"
	bannerURL := "https://example.com/banner.jpg"
	bio := "Software Developer"
	locale := "en"
	timezone := "America/New_York"
	socialsJSON := `{"twitter":"@johndoe"}`
	updatedAt := "2024-01-15T10:30:00Z"

	pbProfile := &pb.Profile{
		UserId:      userID,
		Username:    username,
		DisplayName: displayName,
		AvatarUrl:   avatarURL,
		BannerUrl:   bannerURL,
		Bio:         bio,
		Locale:      locale,
		Timezone:    timezone,
		SocialsJson: socialsJSON,
		UpdatedAt:   updatedAt,
	}

	result := profileFromProto(pbProfile)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.UserID != userID {
		t.Errorf("Expected UserID %q, got %q", userID, result.UserID)
	}

	if result.Username == nil || *result.Username != username {
		t.Errorf("Expected Username %q, got %v", username, result.Username)
	}

	if result.DisplayName == nil || *result.DisplayName != displayName {
		t.Errorf("Expected DisplayName %q, got %v", displayName, result.DisplayName)
	}

	if result.AvatarURL == nil || *result.AvatarURL != avatarURL {
		t.Errorf("Expected AvatarURL %q, got %v", avatarURL, result.AvatarURL)
	}

	if result.BannerURL == nil || *result.BannerURL != bannerURL {
		t.Errorf("Expected BannerURL %q, got %v", bannerURL, result.BannerURL)
	}

	if result.Bio == nil || *result.Bio != bio {
		t.Errorf("Expected Bio %q, got %v", bio, result.Bio)
	}

	if result.Locale == nil || *result.Locale != locale {
		t.Errorf("Expected Locale %q, got %v", locale, result.Locale)
	}

	if result.Timezone == nil || *result.Timezone != timezone {
		t.Errorf("Expected Timezone %q, got %v", timezone, result.Timezone)
	}

	if result.SocialsJSON == nil || *result.SocialsJSON != socialsJSON {
		t.Errorf("Expected SocialsJSON %q, got %v", socialsJSON, result.SocialsJSON)
	}

	if result.UpdatedAt == nil || *result.UpdatedAt != updatedAt {
		t.Errorf("Expected UpdatedAt %q, got %v", updatedAt, result.UpdatedAt)
	}
}

func TestProfileFromProto_EmptyStrings(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"

	pbProfile := &pb.Profile{
		UserId:      userID,
		Username:    "", // Empty strings should become nil
		DisplayName: "",
		AvatarUrl:   "",
		BannerUrl:   "",
		Bio:         "",
		Locale:      "",
		Timezone:    "",
		SocialsJson: "",
		UpdatedAt:   "",
	}

	result := profileFromProto(pbProfile)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Username != nil {
		t.Errorf("Expected Username to be nil for empty string, got %v", result.Username)
	}

	if result.DisplayName != nil {
		t.Errorf("Expected DisplayName to be nil for empty string, got %v", result.DisplayName)
	}

	if result.AvatarURL != nil {
		t.Errorf("Expected AvatarURL to be nil for empty string, got %v", result.AvatarURL)
	}

	if result.BannerURL != nil {
		t.Errorf("Expected BannerURL to be nil for empty string, got %v", result.BannerURL)
	}

	if result.Bio != nil {
		t.Errorf("Expected Bio to be nil for empty string, got %v", result.Bio)
	}
}

func TestWalletLinkFromProto_Nil(t *testing.T) {
	result := walletLinkFromProto(nil)
	if result != nil {
		t.Errorf("Expected nil result for nil input, got %v", result)
	}
}

func TestWalletLinkFromProto_Complete(t *testing.T) {
	id := "wallet-id-123"
	userID := "user-id-456"
	accountID := "eip155:1:0x1234567890123456789012345678901234567890"
	address := "0x1234567890123456789012345678901234567890"
	chainID := "eip155:1"
	verifiedAt := time.Now()
	createdAt := time.Now()
	updatedAt := time.Now()

	pbWallet := &pb.WalletLink{
		Id:         id,
		UserId:     userID,
		AccountId:  accountID,
		Address:    address,
		ChainId:    chainID,
		IsPrimary:  true,
		VerifiedAt: timestamppb.New(verifiedAt),
		CreatedAt:  timestamppb.New(createdAt),
		UpdatedAt:  timestamppb.New(updatedAt),
	}

	result := walletLinkFromProto(pbWallet)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.ID != id {
		t.Errorf("Expected ID %q, got %q", id, result.ID)
	}

	if result.UserID != userID {
		t.Errorf("Expected UserID %q, got %q", userID, result.UserID)
	}

	if result.AccountID != accountID {
		t.Errorf("Expected AccountID %q, got %q", accountID, result.AccountID)
	}

	if result.Address != address {
		t.Errorf("Expected Address %q, got %q", address, result.Address)
	}

	if result.ChainID != chainID {
		t.Errorf("Expected ChainID %q, got %q", chainID, result.ChainID)
	}

	if !result.IsPrimary {
		t.Error("Expected IsPrimary to be true")
	}

	if result.VerifiedAt == nil {
		t.Error("Expected VerifiedAt to be set")
	}

	// CreatedAt and UpdatedAt should be non-empty strings
	if result.CreatedAt == "" {
		t.Error("Expected CreatedAt to be non-empty")
	}

	if result.UpdatedAt == "" {
		t.Error("Expected UpdatedAt to be non-empty")
	}
}

func TestWalletLinkFromProto_NotPrimary(t *testing.T) {
	pbWallet := &pb.WalletLink{
		Id:         "wallet-id",
		UserId:     "user-id",
		AccountId:  "eip155:137:0x9999999999999999999999999999999999999999",
		Address:    "0x9999999999999999999999999999999999999999",
		ChainId:    "eip155:137",
		IsPrimary:  false,
		VerifiedAt: timestamppb.New(time.Now()),
		CreatedAt:  timestamppb.New(time.Now()),
		UpdatedAt:  timestamppb.New(time.Now()),
	}

	result := walletLinkFromProto(pbWallet)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.IsPrimary {
		t.Error("Expected IsPrimary to be false")
	}
}

func TestConversionRoundTrip(t *testing.T) {
	// Test that converting back and forth maintains values
	original := "test value"

	// String -> Pointer -> String
	ptr := valueToStringPtr(original)
	backToString := stringPtrToValue(ptr)

	if backToString != original {
		t.Errorf("Round trip failed: expected %q, got %q", original, backToString)
	}

	// Empty string edge case
	empty := ""
	emptyPtr := valueToStringPtr(empty)
	if emptyPtr != nil {
		t.Error("Expected empty string to convert to nil pointer")
	}

	emptyBack := stringPtrToValue(emptyPtr)
	if emptyBack != "" {
		t.Errorf("Expected nil pointer to convert to empty string, got %q", emptyBack)
	}
}

func TestProfileFromProto_PartialData(t *testing.T) {
	// Test with only required fields and some optional fields
	userID := "user-123"
	username := "partial_user"

	pbProfile := &pb.Profile{
		UserId:   userID,
		Username: username,
		Locale:   "en",
		Timezone: "UTC",
		// Other fields left as empty strings
	}

	result := profileFromProto(pbProfile)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.UserID != userID {
		t.Errorf("Expected UserID %q, got %q", userID, result.UserID)
	}

	if result.Username == nil || *result.Username != username {
		t.Errorf("Expected Username %q, got %v", username, result.Username)
	}

	// Empty optional fields should be nil
	if result.DisplayName != nil {
		t.Errorf("Expected DisplayName to be nil, got %v", result.DisplayName)
	}

	if result.Bio != nil {
		t.Errorf("Expected Bio to be nil, got %v", result.Bio)
	}
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}
