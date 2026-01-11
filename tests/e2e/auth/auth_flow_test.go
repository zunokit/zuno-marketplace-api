package auth_test

import (
	"context"
	"testing"
	"time"

	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	authServiceAddr = "localhost:50051"
	testAddress     = "0x1234567890123456789012345678901234567890"
	testChainID     = "eip155:1"
	testDomain      = "localhost"
)

func TestAuthFlow_GetNonce(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	conn, err := grpc.Dial(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to auth service: %v", err)
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.GetNonceRequest{
		AccountId: testAddress,
		ChainId:   testChainID,
		Domain:    testDomain,
	}

	resp, err := client.GetNonce(ctx, req)
	if err != nil {
		t.Fatalf("GetNonce failed: %v", err)
	}

	if resp.Nonce == "" {
		t.Error("Expected nonce to be generated")
	}

	if len(resp.Nonce) != 64 {
		t.Errorf("Expected nonce length 64, got %d", len(resp.Nonce))
	}

	if resp.ExpiresAt == "" {
		t.Error("Expected expiration time to be set")
	}
}

func TestAuthFlow_GetNonce_InvalidInput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	conn, err := grpc.Dial(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to auth service: %v", err)
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test with empty account ID
	req := &pb.GetNonceRequest{
		AccountId: "",
		ChainId:   testChainID,
		Domain:    testDomain,
	}

	_, err = client.GetNonce(ctx, req)
	if err == nil {
		t.Error("Expected error for empty account ID")
	}
}

// Note: Full VerifySiwe test would require actual signature generation with a real wallet
// This is a placeholder demonstrating the test structure
func TestAuthFlow_VerifySiwe_RequiresRealSignature(t *testing.T) {
	t.Skip("VerifySiwe test requires real wallet signature - run manually with real signatures")

	// Example structure:
	// 1. Generate nonce via GetNonce
	// 2. Create SIWE message
	// 3. Sign with real wallet (metamask, hardhat, etc.)
	// 4. Call VerifySiwe
	// 5. Verify JWT tokens returned
}
