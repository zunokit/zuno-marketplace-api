package config

import (
	"fmt"

	"github.com/zunokit/zuno-marketplace-api/shared/env"
)

// Config holds all configuration for the GraphQL gateway
type Config struct {
	Server        ServerConfig
	JWT           JWTConfig
	Services      ServicesConfig
	Features      FeatureConfig
	Redis         RedisConfig
	RabbitMQ      RabbitMQConfig
	Sentry        SentryConfig
	WebhookSecret string
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	HTTPAddr string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
}

// ServicesConfig holds URLs for dependent services
type ServicesConfig struct {
	AuthServiceURL       string
	UserServiceURL       string
	WalletServiceURL     string
	CollectionServiceURL string
	MediaServiceURL      string
}

// FeatureConfig holds feature flags
type FeatureConfig struct {
	PlaygroundEnabled bool
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
			HTTPAddr: env.GetString("GATEWAY_HTTP_ADDR", ":8081"),
		},
		JWT: JWTConfig{
			Secret: env.GetString("JWT_ACCESS_SECRET", ""),
		},
		Services: ServicesConfig{
			AuthServiceURL:       env.GetString("AUTH_SERVICE_URL", "localhost:50051"),
			UserServiceURL:       env.GetString("USER_SERVICE_URL", "localhost:50052"),
			WalletServiceURL:     env.GetString("WALLET_SERVICE_URL", "localhost:50053"),
			CollectionServiceURL: env.GetString("COLLECTION_SERVICE_URL", "localhost:50054"),
			MediaServiceURL:      env.GetString("MEDIA_SERVICE_URL", "localhost:50055"),
		},
		Features: FeatureConfig{
			PlaygroundEnabled: env.GetBool("GRAPHQL_PLAYGROUND", true),
		},
		Redis:    redisConfig,
		RabbitMQ: rabbitConfig,
		Sentry: SentryConfig{
			DSN:         env.GetString("SENTRY_DSN", ""),
			Environment: env.GetString("SENTRY_ENVIRONMENT", "development"),
		},
		WebhookSecret: env.GetString("WEBHOOK_SECRET", ""),
	}
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

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return ErrMissingJWTSecret
	}
	return nil
}

// Common configuration errors
var (
	ErrMissingJWTSecret = NewConfigError("JWT_ACCESS_SECRET environment variable is required")
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
