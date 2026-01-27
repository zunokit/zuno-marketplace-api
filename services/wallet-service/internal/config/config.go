package config

import (
	"github.com/zunokit/zuno-marketplace-api/shared/env"
)

// Config holds all configuration for the wallet service
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Sentry   SentryConfig
}

// ServerConfig holds gRPC server configuration
type ServerConfig struct {
	GRPCPort string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// SentryConfig holds Sentry monitoring configuration
type SentryConfig struct {
	DSN         string
	Environment string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			GRPCPort: env.GetString("WALLET_GRPC_PORT", ":50053"),
		},
		Database: DatabaseConfig{
			Host:     env.GetString("POSTGRES_HOST", "localhost"),
			Port:     env.GetString("POSTGRES_PORT", "5432"),
			User:     env.GetString("POSTGRES_USER", "postgres"),
			Password: env.GetString("POSTGRES_PASSWORD", "postgres"),
			Database: env.GetString("POSTGRES_DATABASE", "nft_marketplace"),
			SSLMode:  env.GetString("POSTGRES_SSL_MODE", "disable"),
		},
		Sentry: SentryConfig{
			DSN:         env.GetString("SENTRY_DSN", ""),
			Environment: env.GetString("SENTRY_ENVIRONMENT", "development"),
		},
	}
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return "host=" + c.Host + " port=" + c.Port + " user=" + c.User +
		" password=" + c.Password + " dbname=" + c.Database + " sslmode=" + c.SSLMode
}
