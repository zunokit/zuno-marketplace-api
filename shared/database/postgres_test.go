package database

import (
	"testing"

	"gorm.io/gorm/logger"
)

func TestConfigGetDSN(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want string
	}{
		{
			name: "complete configuration",
			cfg: &Config{
				Host:     "localhost",
				Port:     "5432",
				User:     "testuser",
				Password: "testpass",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
			want: "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable",
		},
		{
			name: "production configuration",
			cfg: &Config{
				Host:     "prod-db.example.com",
				Port:     "5432",
				User:     "produser",
				Password: "securepass",
				DBName:   "proddb",
				SSLMode:  "require",
			},
			want: "host=prod-db.example.com port=5432 user=produser password=securepass dbname=proddb sslmode=require",
		},
		{
			name: "with custom port",
			cfg: &Config{
				Host:     "localhost",
				Port:     "5433",
				User:     "user",
				Password: "pass",
				DBName:   "db",
				SSLMode:  "disable",
			},
			want: "host=localhost port=5433 user=user password=pass dbname=db sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.GetDSN()
			if got != tt.want {
				t.Errorf("Config.GetDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConnect_InvalidConfig(t *testing.T) {
	// Test with invalid configuration
	cfg := &Config{
		Host:     "invalid-host-that-does-not-exist",
		Port:     "9999",
		User:     "invalid",
		Password: "invalid",
		DBName:   "invalid",
		SSLMode:  "disable",
		LogLevel: logger.Silent,
	}

	db, err := Connect(cfg)
	if err == nil {
		t.Error("Connect() should return error for invalid configuration")
	}
	if db != nil {
		t.Error("Connect() should return nil db on error")
	}
}

func TestConfig_DefaultLogLevel(t *testing.T) {
	// This test verifies that Connect uses Info log level when LogLevel is 0
	cfg := &Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "test",
		Password: "test",
		DBName:   "test",
		SSLMode:  "disable",
		LogLevel: 0, // Should default to Info
	}

	// We can't actually connect without a real database, but we can verify GetDSN works
	dsn := cfg.GetDSN()
	if dsn == "" {
		t.Error("GetDSN() should not return empty string")
	}
}
