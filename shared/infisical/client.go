/*
Package infisical provides integration with Infisical secrets management platform.
This package replaces local environment file usage with Infisical cloud-based secrets.

Required environment variables:
  - INFISICAL_PROJECT_ID: Infisical project ID
  - INFISICAL_ENVIRONMENT: Environment slug (dev, staging, prod)
  - INFISICAL_SECRET_PATH: Path to secrets (default: "/")
  - INFISICAL_UNIVERSAL_AUTH_CLIENT_ID: Machine identity client ID
  - INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET: Machine identity client secret

Optional:
  - INFISICAL_SITE_URL: Infisical instance URL (default: https://app.infisical.com)
  - INFISICAL_CACHE_TTL: Cache expiry in seconds (default: 300)
  - INFISICAL_ENABLED: Enable Infisical integration (default: false)
*/
package infisical

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	infisicalSdk "github.com/infisical/go-sdk"
)

var (
	globalClient *Client
	once         sync.Once
	initErr      error
)

// Client wraps the Infisical SDK client with caching and configuration
type Client struct {
	sdkClient   infisicalSdk.InfisicalClient
	projectID   string
	environment string
	secretPath  string
	cache       map[string]*cachedSecret
	cacheMu     sync.RWMutex
	cacheTTL    time.Duration
}

// cachedSecret holds a cached secret value with expiry
type cachedSecret struct {
	value     string
	expiresAt time.Time
}

// Config holds Infisical configuration
type Config struct {
	SiteURL         string
	ProjectID       string
	Environment     string
	SecretPath      string
	ClientID        string
	ClientSecret    string
	CacheTTLSeconds int
	Enabled         bool
}

// LoadConfig loads Infisical configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		SiteURL:         getEnv("INFISICAL_SITE_URL", "https://app.infisical.com"),
		ProjectID:       os.Getenv("INFISICAL_PROJECT_ID"),
		Environment:     os.Getenv("INFISICAL_ENVIRONMENT"),
		SecretPath:      getEnv("INFISICAL_SECRET_PATH", "/"),
		ClientID:        os.Getenv("INFISICAL_UNIVERSAL_AUTH_CLIENT_ID"),
		ClientSecret:    os.Getenv("INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET"),
		CacheTTLSeconds: getEnvInt("INFISICAL_CACHE_TTL", 300),
		Enabled:         getEnvBool("INFISICAL_ENABLED", false),
	}
}

// IsConfigured checks if Infisical has all required configuration
func (c *Config) IsConfigured() bool {
	return c.Enabled &&
		c.ProjectID != "" &&
		c.Environment != "" &&
		(c.ClientID != "" || os.Getenv("INFISICAL_UNIVERSAL_AUTH_CLIENT_ID") != "")
}

// NewClient creates a new Infisical client with the given configuration
func NewClient(ctx context.Context, config *Config) (*Client, error) {
	if !config.IsConfigured() {
		return nil, fmt.Errorf("infisical not properly configured: INFISICAL_ENABLED=true, PROJECT_ID, ENVIRONMENT, and auth credentials required")
	}

	sdkClient := infisicalSdk.NewInfisicalClient(ctx, infisicalSdk.Config{
		SiteUrl:              config.SiteURL,
		AutoTokenRefresh:     true,
		CacheExpiryInSeconds: config.CacheTTLSeconds,
		SilentMode:           true,
	})

	// Authenticate using Universal Auth
	_, err := sdkClient.Auth().UniversalAuthLogin(config.ClientID, config.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("infisical authentication failed: %w", err)
	}

	return &Client{
		sdkClient:   sdkClient,
		projectID:   config.ProjectID,
		environment: config.Environment,
		secretPath:  config.SecretPath,
		cache:       make(map[string]*cachedSecret),
		cacheTTL:    time.Duration(config.CacheTTLSeconds) * time.Second,
	}, nil
}

// GetSecret retrieves a secret from Infisical with caching
func (c *Client) GetSecret(key string) (string, error) {
	// Check cache first
	c.cacheMu.RLock()
	if cached, ok := c.cache[key]; ok && time.Now().Before(cached.expiresAt) {
		c.cacheMu.RUnlock()
		return cached.value, nil
	}
	c.cacheMu.RUnlock()

	// Fetch from Infisical
	secret, err := c.sdkClient.Secrets().Retrieve(infisicalSdk.RetrieveSecretOptions{
		SecretKey:   key,
		ProjectID:   c.projectID,
		Environment: c.environment,
		SecretPath:  c.secretPath,
	})

	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret %s: %w", key, err)
	}

	// Cache the result
	c.cacheMu.Lock()
	c.cache[key] = &cachedSecret{
		value:     secret.SecretValue,
		expiresAt: time.Now().Add(c.cacheTTL),
	}
	c.cacheMu.Unlock()

	return secret.SecretValue, nil
}

// GetSecretWithFallback retrieves a secret, falling back to environment variable if Infisical fails
func (c *Client) GetSecretWithFallback(key, fallback string) string {
	// First try Infisical
	if value, err := c.GetSecret(key); err == nil && value != "" {
		return value
	}

	// Fall back to environment variable
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

// ListSecrets retrieves all secrets from the configured path
func (c *Client) ListSecrets() (map[string]string, error) {
	secrets, err := c.sdkClient.Secrets().List(infisicalSdk.ListSecretsOptions{
		ProjectID:   c.projectID,
		Environment: c.environment,
		SecretPath:  c.secretPath,
		Recursive:   false,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	result := make(map[string]string)
	for _, secret := range secrets {
		result[secret.SecretKey] = secret.SecretValue
	}

	return result, nil
}

// InvalidateCache clears the secret cache
func (c *Client) InvalidateCache() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache = make(map[string]*cachedSecret)
}

// Init initializes the global Infisical client (thread-safe)
func Init(ctx context.Context) error {
	once.Do(func() {
		config := LoadConfig()
		if !config.Enabled {
			initErr = fmt.Errorf("infisical not enabled (set INFISICAL_ENABLED=true)")
			return
		}

		globalClient, initErr = NewClient(ctx, config)
	})

	return initErr
}

// GetClient returns the global Infisical client
func GetClient() *Client {
	return globalClient
}

// IsEnabled checks if Infisical is enabled and configured
func IsEnabled() bool {
	config := LoadConfig()
	return config.IsConfigured()
}

// GetSecret retrieves a secret from the global client
func GetSecret(key string) (string, error) {
	if globalClient == nil {
		return "", fmt.Errorf("infisical client not initialized")
	}
	return globalClient.GetSecret(key)
}

// GetSecretWithFallback retrieves a secret with environment fallback
func GetSecretWithFallback(key, fallback string) string {
	if globalClient == nil {
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
		return fallback
	}
	return globalClient.GetSecretWithFallback(key, fallback)
}

// Helper functions
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		return value == "true" || value == "1" || value == "yes"
	}
	return fallback
}
