package config

import (
	sharedConfig "github.com/quangdang46/NFT-Marketplace/shared/config"
	"github.com/quangdang46/NFT-Marketplace/shared/env"
)

// Config holds all configuration for the Media Service
type Config struct {
	Server          sharedConfig.ServerConfig
	MetadataService MetadataServiceConfig
}

// MetadataServiceConfig holds metadata service configuration
type MetadataServiceConfig struct {
	URL    string
	APIKey string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: sharedConfig.ServerConfig{
			GRPCPort: env.GetString("MEDIA_GRPC_PORT", ":50055"),
		},
		MetadataService: MetadataServiceConfig{
			URL:    env.GetString("METADATA_SERVICE_URL", ""),
			APIKey: env.GetString("METADATA_SERVICE_API_KEY", ""),
		},
	}
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.MetadataService.URL == "" {
		return ErrMissingMetadataServiceURL
	}
	if c.MetadataService.APIKey == "" {
		return ErrMissingMetadataServiceAPIKey
	}
	return nil
}

// Common configuration errors
var (
	ErrMissingMetadataServiceURL    = NewConfigError("METADATA_SERVICE_URL environment variable is required")
	ErrMissingMetadataServiceAPIKey = NewConfigError("METADATA_SERVICE_API_KEY environment variable is required")
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
