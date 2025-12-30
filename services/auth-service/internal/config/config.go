package config

import (
	"time"

	sharedConfig "github.com/zunokit/zuno-marketplace-api/shared/config"
	"github.com/zunokit/zuno-marketplace-api/shared/env"
)

// Config holds all configuration for the auth service
type Config struct {
	Server   sharedConfig.ServerConfig
	Database sharedConfig.DatabaseConfig
	JWT      sharedConfig.JWTConfig
	Services sharedConfig.ServicesConfig
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: sharedConfig.ServerConfig{
			GRPCPort: env.GetString("AUTH_GRPC_PORT", ":50051"),
		},
		Database: sharedConfig.DatabaseConfig{
			Host:     env.GetString("POSTGRES_HOST", "localhost"),
			Port:     env.GetString("POSTGRES_PORT", "5432"),
			User:     env.GetString("POSTGRES_USER", "postgres"),
			Password: env.GetString("POSTGRES_PASSWORD", "postgres"),
			Database: env.GetString("POSTGRES_DATABASE", "nft_marketplace"),
			SSLMode:  env.GetString("POSTGRES_SSL_MODE", "disable"),
		},
		JWT: sharedConfig.JWTConfig{
			Secret:            env.GetString("JWT_SECRET", "your-jwt-secret-key-change-in-production"),
			RefreshSecret:     env.GetString("REFRESH_SECRET", "your-refresh-secret-key-change-in-production"),
			AccessExpiration:  time.Duration(env.GetInt("JWT_ACCESS_EXPIRATION_HOURS", 1)) * time.Hour,
			RefreshExpiration: time.Duration(env.GetInt("JWT_REFRESH_EXPIRATION_DAYS", 7)) * 24 * time.Hour,
		},
		Services: sharedConfig.ServicesConfig{
			UserServiceURL:   env.GetString("USER_SERVICE_URL", "localhost:50052"),
			WalletServiceURL: env.GetString("WALLET_SERVICE_URL", "localhost:50053"),
		},
	}
}
