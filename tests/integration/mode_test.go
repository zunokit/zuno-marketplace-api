// Package integration_test provides mode detection and validation tests
// for the hybrid serverless infrastructure.
package integration_test

import (
	"os"
	"testing"

	"github.com/zunokit/zuno-marketplace-api/shared/env"
)

// TestInfrastructureMode validates that INFRA_MODE is set correctly.
func TestInfrastructureMode(t *testing.T) {
	mode := env.GetString("INFRA_MODE", "docker")

	if mode != "docker" && mode != "serverless" {
		t.Fatalf("invalid INFRA_MODE: %s (must be 'docker' or 'serverless')", mode)
	}

	t.Logf("✅ Infrastructure mode: %s", mode)
}

// TestServerlessConnectionStrings validates required environment variables
// for serverless mode.
func TestServerlessConnectionStrings(t *testing.T) {
	mode := env.GetString("INFRA_MODE", "docker")

	if mode != "serverless" {
		t.Skip("only runs in serverless mode")
	}

	// Required for serverless mode
	required := []string{
		"DATABASE_URL",
		"REDIS_URL",
		"CLOUDAMQP_URL",
	}

	missing := []string{}
	for _, key := range required {
		val := os.Getenv(key)
		if val == "" {
			missing = append(missing, key)
		} else {
			t.Logf("✅ %s is set", key)
		}
	}

	if len(missing) > 0 {
		t.Errorf("missing required env vars for serverless mode: %v", missing)
	}
}

// TestDockerConnectionVars validates required environment variables
// for Docker mode.
func TestDockerConnectionVars(t *testing.T) {
	mode := env.GetString("INFRA_MODE", "docker")

	if mode != "docker" {
		t.Skip("only runs in docker mode")
	}

	// Required for Docker mode
	required := []string{
		"POSTGRES_HOST",
		"POSTGRES_PORT",
		"REDIS_HOST",
		"REDIS_PORT",
		"RABBITMQ_HOST",
		"RABBITMQ_PORT",
	}

	for _, key := range required {
		val := os.Getenv(key)
		if val == "" {
			// Use default values from env.go
			t.Logf("⚠️  %s not set (will use default)", key)
		} else {
			t.Logf("✅ %s is set", key)
		}
	}
}

// TestServerlessDatabaseURL validates DATABASE_URL format for serverless mode.
func TestServerlessDatabaseURL(t *testing.T) {
	mode := env.GetString("INFRA_MODE", "docker")

	if mode != "serverless" {
		t.Skip("only runs in serverless mode")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	// Check for postgres:// or postgresql:// prefix
	if len(dbURL) < 11 || (dbURL[:10] != "postgres://" && dbURL[:11] != "postgresql://") {
		t.Errorf("DATABASE_URL has invalid format (should start with postgres:// or postgresql://)")
	} else {
		t.Logf("✅ DATABASE_URL format valid")
	}
}

// TestServerlessRedisURL validates REDIS_URL format for serverless mode.
func TestServerlessRedisURL(t *testing.T) {
	mode := env.GetString("INFRA_MODE", "docker")

	if mode != "serverless" {
		t.Skip("only runs in serverless mode")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set")
	}

	// Check for redis:// prefix
	if len(redisURL) < 9 || redisURL[:9] != "redis://" && redisURL[:8] != "rediss://" {
		t.Errorf("REDIS_URL has invalid format (should start with redis:// or rediss://)")
	} else {
		t.Logf("✅ REDIS_URL format valid")
	}
}

// TestServerlessCloudAMQPURL validates CLOUDAMQP_URL format for serverless mode.
func TestServerlessCloudAMQPURL(t *testing.T) {
	mode := env.GetString("INFRA_MODE", "docker")

	if mode != "serverless" {
		t.Skip("only runs in serverless mode")
	}

	amqpURL := os.Getenv("CLOUDAMQP_URL")
	if amqpURL == "" {
		t.Skip("CLOUDAMQP_URL not set")
	}

	// Check for amqp:// or amqps:// prefix
	if len(amqpURL) < 7 || (amqpURL[:7] != "amqp://" && amqpURL[:8] != "amqps://") {
		t.Errorf("CLOUDAMQP_URL has invalid format (should start with amqp:// or amqps://)")
	} else {
		t.Logf("✅ CLOUDAMQP_URL format valid")
	}
}
