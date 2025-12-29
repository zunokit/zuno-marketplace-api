package config

import (
	"github.com/zunokit/zuno-marketplace-api/shared/env"
)

// Config holds all configuration for the wallet service
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

// ServerConfig holds gRPC server configuration
type ServerConfig struct {
	GRPCPort string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Mode     string // "docker" or "serverless"
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
	URL      string // Full connection URL for serverless mode
}

// Load loads configuration from environment variables
func Load() *Config {
	mode := env.GetString("INFRA_MODE", "docker")

	// Database config
	dbConfig := DatabaseConfig{Mode: mode}
	if mode == "serverless" {
		dbConfig.URL = env.GetString("DATABASE_URL", "")
	} else {
		dbConfig.Host = env.GetString("POSTGRES_HOST", "localhost")
		dbConfig.Port = env.GetString("POSTGRES_PORT", "5432")
		dbConfig.User = env.GetString("POSTGRES_USER", "postgres")
		dbConfig.Password = env.GetString("POSTGRES_PASSWORD", "postgres")
		dbConfig.Database = env.GetString("POSTGRES_DATABASE", "nft_marketplace")
		dbConfig.SSLMode = env.GetString("POSTGRES_SSL_MODE", "disable")
	}

	return &Config{
		Server: ServerConfig{
			GRPCPort: env.GetString("WALLET_GRPC_PORT", ":50053"),
		},
		Database: dbConfig,
	}
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	if c.Mode == "serverless" && c.URL != "" {
		return c.URL
	}
	return "host=" + c.Host + " port=" + c.Port + " user=" + c.User +
		" password=" + c.Password + " dbname=" + c.Database + " sslmode=" + c.SSLMode
}
