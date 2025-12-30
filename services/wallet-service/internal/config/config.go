package config

import (
	sharedConfig "github.com/zunokit/zuno-marketplace-api/shared/config"
	"github.com/zunokit/zuno-marketplace-api/shared/env"
)

// Load loads configuration from environment variables
func Load() *sharedConfig.Config {
	return &sharedConfig.Config{
		Server: sharedConfig.ServerConfig{
			GRPCPort: env.GetString("WALLET_GRPC_PORT", ":50053"),
		},
		Database: sharedConfig.DatabaseConfig{
			Host:     env.GetString("POSTGRES_HOST", "localhost"),
			Port:     env.GetString("POSTGRES_PORT", "5432"),
			User:     env.GetString("POSTGRES_USER", "postgres"),
			Password: env.GetString("POSTGRES_PASSWORD", "postgres"),
			Database: env.GetString("POSTGRES_DATABASE", "nft_marketplace"),
			SSLMode:  env.GetString("POSTGRES_SSL_MODE", "disable"),
		},
	}
}
