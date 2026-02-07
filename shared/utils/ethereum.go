package utils

import (
	"hash/crc32"
	"regexp"
	"strings"
)

var ethereumAddressRegex = regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)

// NormalizeAddress converts an Ethereum address to lowercase for case-insensitive comparisons
// This ensures database queries match regardless of checksum case
func NormalizeAddress(address string) string {
	if address == "" {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(address))
}

// NormalizeAddressPtr normalizes a pointer to an Ethereum address
func NormalizeAddressPtr(address *string) *string {
	if address == nil || *address == "" {
		return address
	}
	normalized := NormalizeAddress(*address)
	return &normalized
}

// IsValidEthereumAddress checks if a string is a valid Ethereum address format
func IsValidEthereumAddress(address string) bool {
	return ethereumAddressRegex.MatchString(address)
}

// IsZeroAddress checks if an address is the zero address
func IsZeroAddress(address string) bool {
	normalized := NormalizeAddress(address)
	return normalized == "0x0000000000000000000000000000000000000000"
}

// GenerateAdvisoryLockKey creates a numeric PostgreSQL advisory lock key from a string identifier.
// CRC32 provides stable hashing with low collision probability for idempotency keys.
func GenerateAdvisoryLockKey(identifier string) int64 {
	checksum := crc32.ChecksumIEEE([]byte(identifier))
	return int64(checksum)
}
