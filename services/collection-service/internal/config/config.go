package config

import (
	"strings"

	sharedConfig "github.com/zunokit/zuno-marketplace-api/shared/config"
	"github.com/zunokit/zuno-marketplace-api/shared/env"
)

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

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	if c.Mode == "serverless" && c.URL != "" {
		// Neon requires SSL mode
		if !strings.Contains(c.URL, "sslmode=") {
			return c.URL + "&sslmode=require"
		}
		return c.URL
	}
	return "host=" + c.Host + " port=" + c.Port + " user=" + c.User +
		" password=" + c.Password + " dbname=" + c.Database + " sslmode=" + c.SSLMode
}

// Load loads configuration from environment variables
func Load() *sharedConfig.Config {
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

	return &sharedConfig.Config{
		Server: sharedConfig.ServerConfig{
			GRPCPort: env.GetString("COLLECTION_GRPC_PORT", ":50054"),
		},
		Database: sharedConfig.DatabaseConfig{
			Host:     dbConfig.Host,
			Port:     dbConfig.Port,
			User:     dbConfig.User,
			Password: dbConfig.Password,
			Database: dbConfig.Database,
			SSLMode:  dbConfig.SSLMode,
		},
		DatabaseDSN: dbConfig.GetDSN(),
	}
}
