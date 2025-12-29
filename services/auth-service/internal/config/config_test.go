package config

import (
	"os"
	"testing"
)

func TestConfig_Load_DockerMode(t *testing.T) {
	// Set env vars for docker mode
	os.Setenv("INFRA_MODE", "docker")
	os.Setenv("POSTGRES_HOST", "test-host")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	os.Setenv("POSTGRES_DATABASE", "testdb")
	os.Setenv("POSTGRES_SSL_MODE", "require")
	os.Setenv("REDIS_HOST", "redis-host")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("RABBITMQ_HOST", "rabbit-host")
	os.Setenv("RABBITMQ_PORT", "5673")
	os.Setenv("RABBITMQ_USER", "rabbit-user")
	os.Setenv("RABBITMQ_PASSWORD", "rabbit-pass")
	os.Setenv("RABBITMQ_EXCHANGE", "test-exchange")

	defer func() {
		os.Unsetenv("INFRA_MODE")
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_PORT")
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("POSTGRES_DATABASE")
		os.Unsetenv("POSTGRES_SSL_MODE")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("RABBITMQ_HOST")
		os.Unsetenv("RABBITMQ_PORT")
		os.Unsetenv("RABBITMQ_USER")
		os.Unsetenv("RABBITMQ_PASSWORD")
		os.Unsetenv("RABBITMQ_EXCHANGE")
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
	if cfg.Redis.Host != "redis-host" {
		t.Errorf("Expected Redis.Host=redis-host, got %s", cfg.Redis.Host)
	}
	if cfg.RabbitMQ.Host != "rabbit-host" {
		t.Errorf("Expected RabbitMQ.Host=rabbit-host, got %s", cfg.RabbitMQ.Host)
	}
}

func TestConfig_Load_ServerlessMode(t *testing.T) {
	os.Setenv("INFRA_MODE", "serverless")
	os.Setenv("DATABASE_URL", "postgresql://user:pass@host:5432/db")
	os.Setenv("REDIS_URL", "redis://localhost:6379")
	os.Setenv("CLOUDAMQP_URL", "amqp://user:pass@host:5672/")

	defer func() {
		os.Unsetenv("INFRA_MODE")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("CLOUDAMQP_URL")
	}()

	cfg := Load()

	if cfg.Database.Mode != "serverless" {
		t.Errorf("Expected Mode=serverless, got %s", cfg.Database.Mode)
	}
	if cfg.Database.URL != "postgresql://user:pass@host:5432/db" {
		t.Errorf("Expected URL=postgresql://user:pass@host:5432/db, got %s", cfg.Database.URL)
	}
	if cfg.Redis.URL != "redis://localhost:6379" {
		t.Errorf("Expected Redis.URL=redis://localhost:6379, got %s", cfg.Redis.URL)
	}
	if cfg.RabbitMQ.URL != "amqp://user:pass@host:5672/" {
		t.Errorf("Expected RabbitMQ.URL=amqp://user:pass@host:5672/, got %s", cfg.RabbitMQ.URL)
	}
}

func TestConfig_Load_UnsetINFRA_MODE(t *testing.T) {
	// Ensure INFRA_MODE is not set
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

func TestRedisConfig_GetAddr_DockerMode(t *testing.T) {
	cfg := RedisConfig{
		Mode: "docker",
		Host: "localhost",
		Port: "6379",
	}

	addr := cfg.GetAddr()
	expected := "localhost:6379"

	if addr != expected {
		t.Errorf("Expected addr=%s, got %s", expected, addr)
	}
}

func TestRedisConfig_GetAddr_ServerlessMode(t *testing.T) {
	cfg := RedisConfig{
		Mode: "serverless",
		URL:  "redis://default:pass@host:6379",
	}

	addr := cfg.GetAddr()
	expected := "redis://default:pass@host:6379"

	if addr != expected {
		t.Errorf("Expected addr=%s, got %s", expected, addr)
	}
}

func TestRabbitMQConfig_GetURL_DockerMode(t *testing.T) {
	cfg := RabbitMQConfig{
		Mode:     "docker",
		Host:     "localhost",
		Port:     "5672",
		User:     "guest",
		Password: "guest",
	}

	url := cfg.GetURL()
	expected := "amqp://guest:guest@localhost:5672/"

	if url != expected {
		t.Errorf("Expected URL=%s, got %s", expected, url)
	}
}

func TestRabbitMQConfig_GetURL_ServerlessMode(t *testing.T) {
	cfg := RabbitMQConfig{
		Mode: "serverless",
		URL:  "amqp://user:pass@host:5672/vhost",
	}

	url := cfg.GetURL()
	expected := "amqp://user:pass@host:5672/vhost"

	if url != expected {
		t.Errorf("Expected URL=%s, got %s", expected, url)
	}
}
