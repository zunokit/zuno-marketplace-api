package config

import (
	"os"
	"testing"
)

func TestConfig_Load_DockerMode(t *testing.T) {
	os.Setenv("INFRA_MODE", "docker")
	os.Setenv("POSTGRES_HOST", "test-host")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	os.Setenv("POSTGRES_DATABASE", "testdb")
	os.Setenv("POSTGRES_SSL_MODE", "require")

	defer func() {
		os.Unsetenv("INFRA_MODE")
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_PORT")
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("POSTGRES_DATABASE")
		os.Unsetenv("POSTGRES_SSL_MODE")
	}()

	cfg := Load()

	if cfg.Database.Mode != "docker" {
		t.Errorf("Expected Mode=docker, got %s", cfg.Database.Mode)
	}
	if cfg.Database.Host != "test-host" {
		t.Errorf("Expected Host=test-host, got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != "5432" {
		t.Errorf("Expected Port=5432, got %s", cfg.Database.Port)
	}
}

func TestConfig_Load_ServerlessMode(t *testing.T) {
	os.Setenv("INFRA_MODE", "serverless")
	os.Setenv("DATABASE_URL", "postgresql://user:pass@host:5432/db")

	defer func() {
		os.Unsetenv("INFRA_MODE")
		os.Unsetenv("DATABASE_URL")
	}()

	cfg := Load()

	if cfg.Database.Mode != "serverless" {
		t.Errorf("Expected Mode=serverless, got %s", cfg.Database.Mode)
	}
	if cfg.Database.URL != "postgresql://user:pass@host:5432/db" {
		t.Errorf("Expected URL=postgresql://user:pass@host:5432/db, got %s", cfg.Database.URL)
	}
}

func TestConfig_Load_UnsetINFRA_MODE(t *testing.T) {
	os.Unsetenv("INFRA_MODE")

	cfg := Load()

	if cfg.Database.Mode != "docker" {
		t.Errorf("Expected default Mode=docker, got %s", cfg.Database.Mode)
	}
}

func TestDatabaseConfig_GetDSN_DockerMode(t *testing.T) {
	cfg := DatabaseConfig{
		Mode:     "docker",
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
	}

	dsn := cfg.GetDSN()
	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"

	if dsn != expected {
		t.Errorf("Expected DSN=%s, got %s", expected, dsn)
	}
}

func TestDatabaseConfig_GetDSN_ServerlessMode(t *testing.T) {
	cfg := DatabaseConfig{
		Mode: "serverless",
		URL:  "postgresql://user:pass@host:5432/db",
	}

	dsn := cfg.GetDSN()
	expected := "postgresql://user:pass@host:5432/db"

	if dsn != expected {
		t.Errorf("Expected DSN=%s, got %s", expected, dsn)
	}
}
