/*
Package database provides shared database connection utilities for all services.
*/
package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	LogLevel logger.LogLevel
}

// GetDSN returns the PostgreSQL connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// Connect establishes a connection to PostgreSQL database using GORM
func Connect(cfg *Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	// Set default log level if not specified
	logLevel := cfg.LogLevel
	if logLevel == 0 {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connected successfully")
	return db, nil
}

// MustConnect establishes a connection to PostgreSQL database and panics on error
func MustConnect(cfg *Config) *gorm.DB {
	db, err := Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}
