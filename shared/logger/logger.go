/*
Package logger provides a structured logging utility for all services.
Uses zerolog for high-performance JSON logging.
*/
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger is a structured logger wrapper
type Logger struct {
	zlog zerolog.Logger
}

// Level represents log level
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelFatal Level = "fatal"
)

// Config holds logger configuration
type Config struct {
	Level       Level
	ServiceName string
	Pretty      bool // Enable pretty console output (for development)
}

// New creates a new logger instance
func New(cfg *Config) *Logger {
	var output io.Writer = os.Stdout

	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339

	// Pretty console output for development
	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Parse log level
	var level zerolog.Level
	switch cfg.Level {
	case LevelDebug:
		level = zerolog.DebugLevel
	case LevelWarn:
		level = zerolog.WarnLevel
	case LevelError:
		level = zerolog.ErrorLevel
	case LevelFatal:
		level = zerolog.FatalLevel
	default:
		level = zerolog.InfoLevel
	}

	// Create logger with service name context
	zlog := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Logger()

	return &Logger{zlog: zlog}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.zlog.Debug().Msg(msg)
}

// Debugf logs a debug message with formatting
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.zlog.Debug().Msgf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.zlog.Info().Msg(msg)
}

// Infof logs an info message with formatting
func (l *Logger) Infof(format string, args ...interface{}) {
	l.zlog.Info().Msgf(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.zlog.Warn().Msg(msg)
}

// Warnf logs a warning message with formatting
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.zlog.Warn().Msgf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.zlog.Error().Msg(msg)
}

// Errorf logs an error message with formatting
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.zlog.Error().Msgf(format, args...)
}

// ErrorWithErr logs an error with error object
func (l *Logger) ErrorWithErr(err error, msg string) {
	l.zlog.Error().Err(err).Msg(msg)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.zlog.Fatal().Msg(msg)
}

// Fatalf logs a fatal message with formatting and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.zlog.Fatal().Msgf(format, args...)
}

// FatalWithErr logs a fatal error with error object and exits
func (l *Logger) FatalWithErr(err error, msg string) {
	l.zlog.Fatal().Err(err).Msg(msg)
}

// WithField adds a field to the logger context
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		zlog: l.zlog.With().Interface(key, value).Logger(),
	}
}

// WithFields adds multiple fields to the logger context
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zlog := l.zlog.With()
	for k, v := range fields {
		zlog = zlog.Interface(k, v)
	}
	return &Logger{zlog: zlog.Logger()}
}

// WithError adds error field to the logger context
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		zlog: l.zlog.With().Err(err).Logger(),
	}
}

// Global logger instance
var global *Logger

// Init initializes the global logger
func Init(cfg *Config) {
	global = New(cfg)
}

// Get returns the global logger instance
func Get() *Logger {
	if global == nil {
		// Default logger if not initialized
		global = New(&Config{
			Level:       LevelInfo,
			ServiceName: "unknown",
			Pretty:      false,
		})
	}
	return global
}

// Global convenience functions

// Debug logs a debug message using global logger
func Debug(msg string) {
	Get().Debug(msg)
}

// Debugf logs a debug message with formatting using global logger
func Debugf(format string, args ...interface{}) {
	Get().Debugf(format, args...)
}

// Info logs an info message using global logger
func Info(msg string) {
	Get().Info(msg)
}

// Infof logs an info message with formatting using global logger
func Infof(format string, args ...interface{}) {
	Get().Infof(format, args...)
}

// Warn logs a warning message using global logger
func Warn(msg string) {
	Get().Warn(msg)
}

// Warnf logs a warning message with formatting using global logger
func Warnf(format string, args ...interface{}) {
	Get().Warnf(format, args...)
}

// Error logs an error message using global logger
func Error(msg string) {
	Get().Error(msg)
}

// Errorf logs an error message with formatting using global logger
func Errorf(format string, args ...interface{}) {
	Get().Errorf(format, args...)
}

// ErrorWithErr logs an error with error object using global logger
func ErrorWithErr(err error, msg string) {
	Get().ErrorWithErr(err, msg)
}

// Fatal logs a fatal message and exits using global logger
func Fatal(msg string) {
	Get().Fatal(msg)
}

// Fatalf logs a fatal message with formatting and exits using global logger
func Fatalf(format string, args ...interface{}) {
	Get().Fatalf(format, args...)
}

// FatalWithErr logs a fatal error with error object and exits using global logger
func FatalWithErr(err error, msg string) {
	Get().FatalWithErr(err, msg)
}

func init() {
	// Set global zerolog defaults
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
