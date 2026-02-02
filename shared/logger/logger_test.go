package logger

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
	}{
		{
			name: "info level logger",
			cfg: &Config{
				Level:       LevelInfo,
				ServiceName: "test-service",
				Pretty:      false,
			},
		},
		{
			name: "debug level logger",
			cfg: &Config{
				Level:       LevelDebug,
				ServiceName: "debug-service",
				Pretty:      false,
			},
		},
		{
			name: "pretty console output",
			cfg: &Config{
				Level:       LevelInfo,
				ServiceName: "pretty-service",
				Pretty:      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.cfg)
			if logger == nil {
				t.Error("New() returned nil logger")
			}
		})
	}
}

func TestLogger_Info(t *testing.T) {
	// Create logger for testing
	cfg := &Config{
		Level:       LevelInfo,
		ServiceName: "test-service",
		Pretty:      false,
	}
	logger := New(cfg)

	// Verify the logger methods don't panic
	logger.Info("test message")
	logger.Infof("test %s", "formatted")
}

func TestLogger_Debug(t *testing.T) {
	cfg := &Config{
		Level:       LevelDebug,
		ServiceName: "test-service",
		Pretty:      false,
	}
	logger := New(cfg)

	logger.Debug("debug message")
	logger.Debugf("debug %s", "formatted")
}

func TestLogger_Warn(t *testing.T) {
	cfg := &Config{
		Level:       LevelWarn,
		ServiceName: "test-service",
		Pretty:      false,
	}
	logger := New(cfg)

	logger.Warn("warning message")
	logger.Warnf("warning %s", "formatted")
}

func TestLogger_Error(t *testing.T) {
	cfg := &Config{
		Level:       LevelError,
		ServiceName: "test-service",
		Pretty:      false,
	}
	logger := New(cfg)

	logger.Error("error message")
	logger.Errorf("error %s", "formatted")
}

func TestLogger_WithField(t *testing.T) {
	cfg := &Config{
		Level:       LevelInfo,
		ServiceName: "test-service",
		Pretty:      false,
	}
	logger := New(cfg)

	contextLogger := logger.WithField("user_id", "123")
	if contextLogger == nil {
		t.Error("WithField() returned nil")
	}

	contextLogger.Info("message with context")
}

func TestLogger_WithFields(t *testing.T) {
	cfg := &Config{
		Level:       LevelInfo,
		ServiceName: "test-service",
		Pretty:      false,
	}
	logger := New(cfg)

	fields := map[string]interface{}{
		"user_id":    "123",
		"request_id": "abc",
		"ip":         "127.0.0.1",
	}

	contextLogger := logger.WithFields(fields)
	if contextLogger == nil {
		t.Error("WithFields() returned nil")
	}

	contextLogger.Info("message with multiple fields")
}

func TestLogger_WithError(t *testing.T) {
	cfg := &Config{
		Level:       LevelError,
		ServiceName: "test-service",
		Pretty:      false,
	}
	logger := New(cfg)

	err := &testError{msg: "test error"}
	contextLogger := logger.WithError(err)
	if contextLogger == nil {
		t.Error("WithError() returned nil")
	}

	contextLogger.Error("error occurred")
}

func TestGlobalLogger(t *testing.T) {
	// Initialize global logger
	Init(&Config{
		Level:       LevelInfo,
		ServiceName: "global-test",
		Pretty:      false,
	})

	// Test global functions
	Info("global info")
	Infof("global info %s", "formatted")
	Debug("global debug")
	Debugf("global debug %s", "formatted")
	Warn("global warn")
	Warnf("global warn %s", "formatted")
	Error("global error")
	Errorf("global error %s", "formatted")

	// Get global logger
	logger := Get()
	if logger == nil {
		t.Error("Get() returned nil")
	}
}

func TestGet_WithoutInit(t *testing.T) {
	// Reset global logger
	global = nil

	// Should return default logger
	logger := Get()
	if logger == nil {
		t.Error("Get() should return default logger when not initialized")
	}

	// Should be able to log
	logger.Info("test message")
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{"debug level", LevelDebug},
		{"info level", LevelInfo},
		{"warn level", LevelWarn},
		{"error level", LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Level:       tt.level,
				ServiceName: "level-test",
				Pretty:      false,
			}
			logger := New(cfg)
			if logger == nil {
				t.Errorf("Failed to create logger with level %s", tt.level)
			}
		})
	}
}

func TestConfig_DefaultLevel(t *testing.T) {
	// Test with invalid level defaults to Info
	cfg := &Config{
		Level:       Level("invalid"),
		ServiceName: "default-test",
		Pretty:      false,
	}
	logger := New(cfg)
	if logger == nil {
		t.Error("Failed to create logger with default level")
	}
}

// Helper types for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestJSONOutput(t *testing.T) {
	// Note: This is a simplified test
	// In real scenarios, you'd redirect zerolog output to a buffer
	cfg := &Config{
		Level:       LevelInfo,
		ServiceName: "json-test",
		Pretty:      false,
	}

	logger := New(cfg)
	logger.Info("test json message")

	// Verify logger was created
	if logger == nil {
		t.Error("Failed to create JSON logger")
	}
}

func TestPrettyOutput(t *testing.T) {
	cfg := &Config{
		Level:       LevelInfo,
		ServiceName: "pretty-test",
		Pretty:      true,
	}

	logger := New(cfg)
	if logger == nil {
		t.Error("Failed to create pretty logger")
	}

	logger.Info("pretty formatted message")
	logger.WithField("key", "value").Info("pretty with context")
}

func TestContextualLogging(t *testing.T) {
	cfg := &Config{
		Level:       LevelInfo,
		ServiceName: "context-test",
		Pretty:      false,
	}

	logger := New(cfg)

	// Chain context
	contextLogger := logger.
		WithField("request_id", "req-123").
		WithField("user_id", "user-456")

	if contextLogger == nil {
		t.Error("Failed to create contextual logger")
	}

	contextLogger.Info("contextual log message")
}
