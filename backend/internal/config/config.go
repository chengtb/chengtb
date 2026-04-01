package config

import (
	"log"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config is the top-level application configuration.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Business BusinessConfig `yaml:"business"`
}

// ServerConfig holds HTTP/WebSocket server settings.
type ServerConfig struct {
	Port string `yaml:"port"`
}

// DatabaseConfig holds database connection and pool settings.
type DatabaseConfig struct {
	Host               string `yaml:"host"`
	Port               string `yaml:"port"`
	User               string `yaml:"user"`
	Password           string `yaml:"password"`
	Name               string `yaml:"name"`
	MaxIdleConns       int    `yaml:"max_idle_conns"`
	MaxOpenConns       int    `yaml:"max_open_conns"`
	ConnMaxLifetimeMin int    `yaml:"conn_max_lifetime_minutes"`
}

// BusinessConfig holds business-logic parameters.
type BusinessConfig struct {
	MergeWindowMinutes int                    `yaml:"merge_window_minutes"`
	RefundPermissions  RefundPermissionConfig `yaml:"refund_permissions"`
}

// RefundPermissionConfig defines the minimum staff role required
// to approve each category of refund.
type RefundPermissionConfig struct {
	NotMade       string `yaml:"not_made"`        // e.g. "waiter"
	MadeNotServed string `yaml:"made_not_served"` // e.g. "leader"
	Served        string `yaml:"served"`          // e.g. "manager"
}

// DefaultConfig returns a Config populated with built-in defaults.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host:               "127.0.0.1",
			Port:               "3306",
			User:               "root",
			Password:           "",
			Name:               "restaurant_kds",
			MaxIdleConns:       10,
			MaxOpenConns:       100,
			ConnMaxLifetimeMin: 60,
		},
		Business: BusinessConfig{
			MergeWindowMinutes: 5,
			RefundPermissions: RefundPermissionConfig{
				NotMade:       "waiter",
				MadeNotServed: "leader",
				Served:        "manager",
			},
		},
	}
}

// Load reads configuration from config.yaml (if present), applies
// built-in defaults for any missing fields, and finally lets
// environment variables override individual values.
func Load() *Config {
	cfg := DefaultConfig()

	data, err := os.ReadFile("config.yaml")
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			log.Printf("Warning: failed to parse config.yaml: %v (using defaults)", err)
		}
	}

	applyEnvOverrides(cfg)
	return cfg
}

// LoadFromBytes creates a Config from raw YAML bytes, using built-in
// defaults for any values not present in the YAML.
func LoadFromBytes(data []byte) *Config {
	cfg := DefaultConfig()
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			log.Printf("Warning: failed to parse YAML config: %v (using defaults)", err)
		}
	}
	return cfg
}

// DSN builds a MySQL DSN from the database configuration.
func (c *Config) DSN() string {
	return c.Database.User + ":" + c.Database.Password +
		"@tcp(" + c.Database.Host + ":" + c.Database.Port + ")/" +
		c.Database.Name + "?charset=utf8mb4&parseTime=True&loc=Local"
}

// applyEnvOverrides lets environment variables take precedence over
// file / default values, preserving backward compatibility.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		cfg.Database.Port = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.Name = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		cfg.Server.Port = v
	}
	if v := os.Getenv("MERGE_WINDOW_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Business.MergeWindowMinutes = n
		}
	}
}
