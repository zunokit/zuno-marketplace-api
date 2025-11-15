package service

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spruceid/siwe-go"
)

type SIWEService struct{}

func NewSIWEService() *SIWEService {
	return &SIWEService{}
}

// ParseMessage parses a SIWE message string
func (s *SIWEService) ParseMessage(messageStr string) (*siwe.Message, error) {
	message, err := siwe.ParseMessage(messageStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SIWE message: %w", err)
	}
	return message, nil
}

// VerifySignature verifies an EIP-191 signature
func (s *SIWEService) VerifySignature(message, signature, address string) error {
	// Parse SIWE message
	siweMsg, err := siwe.ParseMessage(message)
	if err != nil {
		return fmt.Errorf("invalid SIWE message: %w", err)
	}

	// Verify address matches
	if !strings.EqualFold(siweMsg.GetAddress().Hex(), address) {
		return fmt.Errorf("address mismatch: expected %s, got %s", address, siweMsg.GetAddress().Hex())
	}

	// Decode signature
	sigBytes, err := hexutil.Decode(signature)
	if err != nil {
		return fmt.Errorf("invalid signature format: %w", err)
	}

	// Ethereum signature format: v should be 27 or 28, convert to 0 or 1
	if len(sigBytes) == 65 {
		if sigBytes[64] >= 27 {
			sigBytes[64] -= 27
		}
	}

	// Prepare the message for signing (EIP-191 format)
	messageHash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))

	// Recover public key from signature
	pubKey, err := crypto.SigToPub(messageHash.Bytes(), sigBytes)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	// Get address from public key
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	// Compare addresses (case-insensitive)
	expectedAddr := common.HexToAddress(address)
	if !strings.EqualFold(recoveredAddr.Hex(), expectedAddr.Hex()) {
		return fmt.Errorf("signature verification failed: recovered address %s != expected %s",
			recoveredAddr.Hex(), expectedAddr.Hex())
	}

	return nil
}

// ValidateSIWEMessage validates SIWE message fields
func (s *SIWEService) ValidateSIWEMessage(message *siwe.Message, expectedDomain, expectedChainID, expectedNonce string) error {
	// Validate domain
	if message.GetDomain() != expectedDomain {
		return fmt.Errorf("domain mismatch: expected %s, got %s", expectedDomain, message.GetDomain())
	}

	// Validate chain ID
	if fmt.Sprintf("eip155:%d", message.GetChainID()) != expectedChainID {
		return fmt.Errorf("chain ID mismatch: expected %s, got eip155:%d", expectedChainID, message.GetChainID())
	}

	// Validate nonce
	if message.GetNonce() != expectedNonce {
		return fmt.Errorf("nonce mismatch")
	}

	// Validate expiration
	if message.GetExpirationTime() != nil && message.GetExpirationTime().Before(message.GetIssuedAt()) {
		return fmt.Errorf("message expired")
	}

	return nil
}
