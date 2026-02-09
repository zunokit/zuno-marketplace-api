package config

import (
	"strings"

	"github.com/zunokit/zuno-marketplace-api/shared/env"
)

// Config holds configuration for the collection service.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Sentry   SentryConfig
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	GRPCPort string
}

// DatabaseConfig holds database connection configuration.
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

// GetDSN returns the database connection string.
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

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Mode string // "docker" or "serverless"
	Host string
	Port string
	URL  string // Full URL for serverless (Upstash)
}

// GetAddr returns the Redis address.
func (c *RedisConfig) GetAddr() string {
	if c.Mode == "serverless" && c.URL != "" {
		return c.URL
	}
	return c.Host + ":" + c.Port
}

// SentryConfig holds Sentry monitoring configuration.
type SentryConfig struct {
	DSN         string
	Environment string
}

// Load loads configuration from environment variables.
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

	// Redis config
	redisConfig := RedisConfig{Mode: mode}
	if mode == "serverless" {
		redisConfig.URL = env.GetString("REDIS_URL", "")
	} else {
		redisConfig.Host = env.GetString("REDIS_HOST", "localhost")
		redisConfig.Port = env.GetString("REDIS_PORT", "6379")
	}

	return &Config{
		Server: ServerConfig{
			GRPCPort: env.GetString("COLLECTION_GRPC_PORT", ":50054"),
		},
		Database: dbConfig,
		Redis:    redisConfig,
		Sentry: SentryConfig{
			DSN:         env.GetString("SENTRY_DSN", ""),
			Environment: env.GetString("SENTRY_ENVIRONMENT", "development"),
		},
	}
}
