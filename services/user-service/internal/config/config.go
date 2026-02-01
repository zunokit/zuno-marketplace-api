package config

import (
	"fmt"
	"strings"

	"github.com/zunokit/zuno-marketplace-api/shared/env"
)

// Config holds all configuration for the user service
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
	Sentry   SentryConfig
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

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Mode string // "docker" or "serverless"
	Host string
	Port string
	URL  string // Full URL for serverless
}

// RabbitMQConfig holds RabbitMQ connection configuration
type RabbitMQConfig struct {
	Mode     string // "docker" or "serverless"
	Host     string
	Port     string
	User     string
	Password string
	Exchange string
	URL      string // Full URL for serverless
}

// SentryConfig holds Sentry monitoring configuration
type SentryConfig struct {
	DSN         string
	Environment string
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

	// Redis config
	redisConfig := RedisConfig{Mode: mode}
	if mode == "serverless" {
		redisConfig.URL = env.GetString("REDIS_URL", "")
	} else {
		redisConfig.Host = env.GetString("REDIS_HOST", "localhost")
		redisConfig.Port = env.GetString("REDIS_PORT", "6379")
	}

	// RabbitMQ config
	rabbitConfig := RabbitMQConfig{Mode: mode}
	if mode == "serverless" {
		rabbitConfig.URL = env.GetString("CLOUDAMQP_URL", "")
	} else {
		rabbitConfig.Host = env.GetString("RABBITMQ_HOST", "localhost")
		rabbitConfig.Port = env.GetString("RABBITMQ_PORT", "5672")
		rabbitConfig.User = env.GetString("RABBITMQ_USER", "guest")
		rabbitConfig.Password = env.GetString("RABBITMQ_PASSWORD", "guest")
		rabbitConfig.Exchange = env.GetString("RABBITMQ_EXCHANGE", "nft_events")
	}

	return &Config{
		Server: ServerConfig{
			GRPCPort: env.GetString("USER_GRPC_PORT", ":50052"),
		},
		Database: dbConfig,
		Redis:    redisConfig,
		RabbitMQ: rabbitConfig,
		Sentry: SentryConfig{
			DSN:         env.GetString("SENTRY_DSN", ""),
			Environment: env.GetString("SENTRY_ENVIRONMENT", "development"),
		},
	}
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

// GetURL returns the RabbitMQ connection URL
func (c *RabbitMQConfig) GetURL() string {
	if c.Mode == "serverless" && c.URL != "" {
		return c.URL
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%s/",
		c.User, c.Password, c.Host, c.Port)
}
