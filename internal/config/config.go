package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/mylxsw/n8n-parallels/internal/logger"
)

// Config represents the application configuration
type Config struct {
	Server ServerConfig       `json:"server"`
	Logger logger.Config      `json:"logger"`
}

// ServerConfig represents the HTTP server configuration
type ServerConfig struct {
	Port            int    `json:"port"`
	Host            string `json:"host"`
	ReadTimeout     int    `json:"read_timeout"`     // seconds
	WriteTimeout    int    `json:"write_timeout"`    // seconds
	ShutdownTimeout int    `json:"shutdown_timeout"` // seconds
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	config := &Config{
		Server: ServerConfig{
			Port:            getEnvAsInt("PORT", 8080),
			Host:            getEnv("HOST", "0.0.0.0"),
			ReadTimeout:     getEnvAsInt("READ_TIMEOUT", 30),
			WriteTimeout:    getEnvAsInt("WRITE_TIMEOUT", 30),
			ShutdownTimeout: getEnvAsInt("SHUTDOWN_TIMEOUT", 30),
		},
		Logger: logger.Config{
			Level:  logger.LogLevel(getEnv("LOG_LEVEL", "info")),
			Format: getEnv("LOG_FORMAT", "text"), // "text" or "json"
		},
	}

	return config
}

// LoadFromFile loads configuration from a JSON file
func LoadFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if they exist
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	if host := os.Getenv("HOST"); host != "" {
		config.Server.Host = host
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.Logger.Level = logger.LogLevel(logLevel)
	}

	if logFormat := os.Getenv("LOG_FORMAT"); logFormat != "" {
		config.Logger.Format = logFormat
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d, must be between 1 and 65535", c.Server.Port)
	}

	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout must be greater than 0")
	}

	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout must be greater than 0")
	}

	if c.Server.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown_timeout must be greater than 0")
	}

	validLevels := map[logger.LogLevel]bool{
		logger.LevelDebug: true,
		logger.LevelInfo:  true,
		logger.LevelWarn:  true,
		logger.LevelError: true,
	}

	if !validLevels[c.Logger.Level] {
		return fmt.Errorf("invalid log level: %s", c.Logger.Level)
	}

	if c.Logger.Format != "text" && c.Logger.Format != "json" {
		return fmt.Errorf("invalid log format: %s, must be 'text' or 'json'", c.Logger.Format)
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(name string, defaultValue int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}