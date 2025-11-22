package config

import "time"

// Config holds generic configuration for services
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
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// JWTConfig holds JWT token configuration
type JWTConfig struct {
	Secret            string
	RefreshSecret     string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
}

// ServicesConfig holds URLs for dependent services
type ServicesConfig struct {
	AuthServiceURL       string
	UserServiceURL       string
	WalletServiceURL     string
	CollectionServiceURL string
}
