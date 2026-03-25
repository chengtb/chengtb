package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration values.
type Config struct {
	// HTTP API server
	HTTPAddr string

	// NATS
	NATSAddr string

	// Log level: debug | info | warn | error
	LogLevel string

	// Service name used as NATS subject prefix
	ServiceName string
}

// Load reads configuration from environment variables, falling back to
// sensible defaults so the application works out of the box.
func Load() *Config {
	return &Config{
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		NATSAddr:    getEnv("NATS_ADDR", "nats://localhost:4222"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		ServiceName: getEnv("SERVICE_NAME", "user-service"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// GetIntEnv reads an integer from an environment variable.
func GetIntEnv(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}
