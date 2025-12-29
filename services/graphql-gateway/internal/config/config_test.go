package config

import (
	"os"
	"testing"
)

func TestConfig_Load_DockerMode(t *testing.T) {
	os.Setenv("INFRA_MODE", "docker")
	os.Setenv("REDIS_HOST", "redis-host")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("RABBITMQ_HOST", "rabbit-host")
	os.Setenv("RABBITMQ_PORT", "5673")
	os.Setenv("RABBITMQ_USER", "rabbit-user")
	os.Setenv("RABBITMQ_PASSWORD", "rabbit-pass")
	os.Setenv("RABBITMQ_EXCHANGE", "test-exchange")

	defer func() {
		os.Unsetenv("INFRA_MODE")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("RABBITMQ_HOST")
		os.Unsetenv("RABBITMQ_PORT")
		os.Unsetenv("RABBITMQ_USER")
		os.Unsetenv("RABBITMQ_PASSWORD")
		os.Unsetenv("RABBITMQ_EXCHANGE")
	}()

	cfg := Load()

	if cfg.Redis.Mode != "docker" {
		t.Errorf("Expected Mode=docker, got %s", cfg.Redis.Mode)
	}
	if cfg.Redis.Host != "redis-host" {
		t.Errorf("Expected Host=redis-host, got %s", cfg.Redis.Host)
	}
	if cfg.RabbitMQ.Host != "rabbit-host" {
		t.Errorf("Expected RabbitMQ.Host=rabbit-host, got %s", cfg.RabbitMQ.Host)
	}
}

func TestConfig_Load_ServerlessMode(t *testing.T) {
	os.Setenv("INFRA_MODE", "serverless")
	os.Setenv("REDIS_URL", "redis://localhost:6379")
	os.Setenv("CLOUDAMQP_URL", "amqp://user:pass@host:5672/")

	defer func() {
		os.Unsetenv("INFRA_MODE")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("CLOUDAMQP_URL")
	}()

	cfg := Load()

	if cfg.Redis.Mode != "serverless" {
		t.Errorf("Expected Mode=serverless, got %s", cfg.Redis.Mode)
	}
	if cfg.Redis.URL != "redis://localhost:6379" {
		t.Errorf("Expected URL=redis://localhost:6379, got %s", cfg.Redis.URL)
	}
	if cfg.RabbitMQ.URL != "amqp://user:pass@host:5672/" {
		t.Errorf("Expected RabbitMQ.URL=amqp://user:pass@host:5672/, got %s", cfg.RabbitMQ.URL)
	}
}

func TestConfig_Load_UnsetINFRA_MODE(t *testing.T) {
	os.Unsetenv("INFRA_MODE")

	cfg := Load()

	if cfg.Redis.Mode != "docker" {
		t.Errorf("Expected default Mode=docker, got %s", cfg.Redis.Mode)
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
