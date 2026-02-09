package config

import (
	"strings"

	sharedConfig "github.com/zunokit/zuno-marketplace-api/shared/config"
	"github.com/zunokit/zuno-marketplace-api/shared/env"
)

// Config holds configuration for the collection service.
type Config struct {
	Server      sharedConfig.ServerConfig
	Database    sharedConfig.DatabaseConfig
	DatabaseDSN string
	Sentry      SentryConfig
}

// SentryConfig holds Sentry monitoring configuration.
type SentryConfig struct {
	DSN         string
	Environment string
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

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Mode string // "docker" or "serverless"
	Host string
	Port string
	URL  string // Full URL for serverless (Upstash)
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

// GetAddr returns the Redis address
func (c *RedisConfig) GetAddr() string {
	if c.Mode == "serverless" && c.URL != "" {
		return c.URL
	}
	return c.Host + ":" + c.Port
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

	return &Config{
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
		Sentry: SentryConfig{
			DSN:         env.GetString("SENTRY_DSN", ""),
			Environment: env.GetString("SENTRY_ENVIRONMENT", "development"),
		},
	}
}

// GetRedisConfig returns Redis configuration (call this in main.go)
func GetRedisConfig() RedisConfig {
	mode := env.GetString("INFRA_MODE", "docker")
	redisConfig := RedisConfig{Mode: mode}
	if mode == "serverless" {
		redisConfig.URL = env.GetString("REDIS_URL", "")
	} else {
		redisConfig.Host = env.GetString("REDIS_HOST", "localhost")
		redisConfig.Port = env.GetString("REDIS_PORT", "6379")
	}
	return redisConfig
}
