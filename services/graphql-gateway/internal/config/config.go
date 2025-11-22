package config

import (
	sharedConfig "github.com/quangdang46/NFT-Marketplace/shared/config"
	"github.com/quangdang46/NFT-Marketplace/shared/env"
)

// Config holds all configuration for the GraphQL gateway
type Config struct {
	Server   ServerConfig
	JWT      sharedConfig.JWTConfig
	Services sharedConfig.ServicesConfig
	Features FeatureConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	HTTPAddr string
}

// FeatureConfig holds feature flags
type FeatureConfig struct {
	PlaygroundEnabled bool
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			HTTPAddr: env.GetString("GATEWAY_HTTP_ADDR", ":8081"),
		},
		JWT: sharedConfig.JWTConfig{
			Secret: env.GetString("JWT_ACCESS_SECRET", ""),
		},
		Services: sharedConfig.ServicesConfig{
			AuthServiceURL:       env.GetString("AUTH_SERVICE_URL", "localhost:50051"),
			UserServiceURL:       env.GetString("USER_SERVICE_URL", "localhost:50052"),
			WalletServiceURL:     env.GetString("WALLET_SERVICE_URL", "localhost:50053"),
			CollectionServiceURL: env.GetString("COLLECTION_SERVICE_URL", "localhost:50054"),
			MediaServiceURL:      env.GetString("MEDIA_SERVICE_URL", "localhost:50055"),
		},
		Features: FeatureConfig{
			PlaygroundEnabled: env.GetBool("GRAPHQL_PLAYGROUND", true),
		},
	}
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return ErrMissingJWTSecret
	}
	return nil
}

// Common configuration errors
var (
	ErrMissingJWTSecret = NewConfigError("JWT_ACCESS_SECRET environment variable is required")
)

// ConfigError represents a configuration validation error
type ConfigError struct {
	message string
}

func NewConfigError(message string) error {
	return &ConfigError{message: message}
}

func (e *ConfigError) Error() string {
	return e.message
}
