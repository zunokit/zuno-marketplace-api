/*
Package env provides environment variable access with Infisical integration.

This package extends the base environment variable functionality by adding
Infisical secrets management as a fallback or primary source.

Usage:

	// Standard environment variable access
	value := env.GetString("DATABASE_URL", "default")

	// With Infisical fallback (requires initialization)
	if err := env.InitInfisical(ctx); err != nil {
	    log.Printf("Infisical not available: %v", err)
	}

	// Now GetString will check Infisical if env var is not set
	value := env.GetString("DATABASE_URL", "default")

Configuration:

	Set INFISICAL_ENABLED=true and other INFISICAL_* variables to enable.
	See shared/infisical package for full configuration options.
*/
package env

import (
	"context"
	"os"
	"strconv"
	"sync"

	"github.com/zunokit/zuno-marketplace-api/shared/infisical"
)

var (
	infisicalEnabled bool
	initOnce         sync.Once
	initErr          error
)

// InitInfisical initializes the Infisical client for secret fallback.
// Call this early in your application startup.
// If Infisical is not configured, this returns an error but doesn't panic.
func InitInfisical(ctx context.Context) error {
	initOnce.Do(func() {
		if !infisical.IsEnabled() {
			initErr = nil // Not an error, just not enabled
			return
		}

		initErr = infisical.Init(ctx)
		if initErr == nil {
			infisicalEnabled = true
		}
	})
	return initErr
}

// IsInfisicalEnabled returns true if Infisical is enabled and initialized
func IsInfisicalEnabled() bool {
	return infisicalEnabled
}

// GetString retrieves a string value with Infisical fallback.
// Priority: 1) Environment variable, 2) Infisical secret, 3) Fallback value
func GetString(key, fallback string) string {
	// 1. Check environment variable first
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	// 2. Try Infisical if enabled
	if infisicalEnabled {
		if client := infisical.GetClient(); client != nil {
			if value, err := client.GetSecret(key); err == nil {
				return value
			}
		}
	}

	// 3. Return fallback
	return fallback
}

// GetStringFromInfisical retrieves a string value from Infisical first, then env var.
// Priority: 1) Infisical secret, 2) Environment variable, 3) Fallback value
// Use this when you want Infisical to override local environment variables.
func GetStringFromInfisical(key, fallback string) string {
	// 1. Try Infisical first if enabled
	if infisicalEnabled {
		if client := infisical.GetClient(); client != nil {
			if value, err := client.GetSecret(key); err == nil {
				return value
			}
		}
	}

	// 2. Check environment variable
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	// 3. Return fallback
	return fallback
}

// GetInt retrieves an integer value with Infisical fallback
func GetInt(key string, fallback int) int {
	value := GetString(key, "")
	if value == "" {
		return fallback
	}

	valAsInt, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return valAsInt
}

// GetIntFromInfisical retrieves an integer from Infisical first
func GetIntFromInfisical(key string, fallback int) int {
	value := GetStringFromInfisical(key, "")
	if value == "" {
		return fallback
	}

	valAsInt, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return valAsInt
}

// GetBool retrieves a boolean value with Infisical fallback
func GetBool(key string, fallback bool) bool {
	value := GetString(key, "")
	if value == "" {
		return fallback
	}

	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return boolVal
}

// GetBoolFromInfisical retrieves a boolean from Infisical first
func GetBoolFromInfisical(key string, fallback bool) bool {
	value := GetStringFromInfisical(key, "")
	if value == "" {
		return fallback
	}

	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return boolVal
}

// MustGetString retrieves a string that must exist (panics if not found)
func MustGetString(key string) string {
	value := GetString(key, "")
	if value == "" {
		panic("required environment variable not set: " + key)
	}
	return value
}

// MustGetStringFromInfisical retrieves a string from Infisical that must exist
func MustGetStringFromInfisical(key string) string {
	value := GetStringFromInfisical(key, "")
	if value == "" {
		panic("required secret not found in Infisical or environment: " + key)
	}
	return value
}

// InvalidateInfisicalCache clears the Infisical secret cache
func InvalidateInfisicalCache() {
	if client := infisical.GetClient(); client != nil {
		client.InvalidateCache()
	}
}

// GetInfisicalClient returns the underlying Infisical client for advanced usage
func GetInfisicalClient() *infisical.Client {
	return infisical.GetClient()
}
