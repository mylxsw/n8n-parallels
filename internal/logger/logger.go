package logger

import (
	"log/slog"
	"os"
	"strings"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Config represents logger configuration
type Config struct {
	Level  LogLevel `json:"level"`
	Format string   `json:"format"` // "text" or "json"
}

// New creates a new structured logger
func New(config Config) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(string(config.Level)) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// NewDefault creates a default logger with reasonable defaults
func NewDefault() *slog.Logger {
	return New(Config{
		Level:  LevelInfo,
		Format: "text",
	})
}