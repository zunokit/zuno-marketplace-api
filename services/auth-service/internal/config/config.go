package config

import (
	"time"

	"github.com/quangdang46/NFT-Marketplace/shared/env"
)

// Config holds all configuration for the auth service
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Services ServicesConfig
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

// JWTConfig holds JWT token configuration
type JWTConfig struct {
	Secret            string
	RefreshSecret     string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
}

// ServicesConfig holds URLs for dependent services
type ServicesConfig struct {
	UserServiceURL   string
	WalletServiceURL string
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
			GRPCPort: env.GetString("AUTH_GRPC_PORT", ":50051"),
		},
		Database: DatabaseConfig{
			Host:     env.GetString("POSTGRES_HOST", "localhost"),
			Port:     env.GetString("POSTGRES_PORT", "5432"),
			User:     env.GetString("POSTGRES_USER", "postgres"),
			Password: env.GetString("POSTGRES_PASSWORD", "postgres"),
			Database: env.GetString("POSTGRES_DATABASE", "nft_marketplace"),
			SSLMode:  env.GetString("POSTGRES_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:            env.GetString("JWT_SECRET", "your-jwt-secret-key-change-in-production"),
			RefreshSecret:     env.GetString("REFRESH_SECRET", "your-refresh-secret-key-change-in-production"),
			AccessExpiration:  time.Duration(env.GetInt("JWT_ACCESS_EXPIRATION_HOURS", 1)) * time.Hour,
			RefreshExpiration: time.Duration(env.GetInt("JWT_REFRESH_EXPIRATION_DAYS", 7)) * 24 * time.Hour,
		},
		Services: ServicesConfig{
			UserServiceURL:   env.GetString("USER_SERVICE_URL", "localhost:50052"),
			WalletServiceURL: env.GetString("WALLET_SERVICE_URL", "localhost:50053"),
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
