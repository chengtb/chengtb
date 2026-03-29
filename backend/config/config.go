package config

import (
	"log"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// yamlConfig mirrors the structure of config.yaml.
type yamlConfig struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"database"`
	Dispatch struct {
		IntervalSec int `yaml:"interval_sec"`
		WindowMin   int `yaml:"window_min"`
	} `yaml:"dispatch"`
}

// Config holds the resolved runtime configuration.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string

	DispatchIntervalSec int
	DispatchWindowMin   int
}

// Load reads configuration from a YAML file.
// The file path defaults to "config.yaml" and can be overridden with the
// CONFIG_FILE environment variable. Individual fields can still be overridden
// by environment variables (same names as before), which take precedence over
// the YAML values.
func Load() *Config {
	filePath := getEnv("CONFIG_FILE", "config.yaml")

	var y yamlConfig
	// Set built-in defaults before parsing YAML.
	y.Server.Port = "8080"
	y.Database.Host = "localhost"
	y.Database.Port = "3306"
	y.Database.User = "root"
	y.Database.Name = "restaurant"
	y.Dispatch.IntervalSec = 30
	y.Dispatch.WindowMin = 5

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("config: cannot read %s (%v), using defaults/env vars", filePath, err)
	} else {
		if err := yaml.Unmarshal(data, &y); err != nil {
			log.Fatalf("config: failed to parse %s: %v", filePath, err)
		}
		log.Printf("config: loaded from %s", filePath)
	}

	// Environment variables override YAML values.
	serverPort := pickStr(getEnv("SERVER_PORT", ""), y.Server.Port)
	dbHost := pickStr(getEnv("DB_HOST", ""), y.Database.Host)
	dbPort := pickStr(getEnv("DB_PORT", ""), y.Database.Port)
	dbUser := pickStr(getEnv("DB_USER", ""), y.Database.User)
	dbPassword := pickStr(getEnv("DB_PASSWORD", ""), y.Database.Password)
	dbName := pickStr(getEnv("DB_NAME", ""), y.Database.Name)

	dispatchInterval := y.Dispatch.IntervalSec
	if v := getEnv("DISPATCH_INTERVAL_SEC", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			dispatchInterval = n
		}
	}
	dispatchWindow := y.Dispatch.WindowMin
	if v := getEnv("DISPATCH_WINDOW_MIN", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			dispatchWindow = n
		}
	}

	return &Config{
		DBHost:              dbHost,
		DBPort:              dbPort,
		DBUser:              dbUser,
		DBPassword:          dbPassword,
		DBName:              dbName,
		ServerPort:          serverPort,
		DispatchIntervalSec: dispatchInterval,
		DispatchWindowMin:   dispatchWindow,
	}
}

// pickStr returns override if non-empty, otherwise fallback.
func pickStr(override, fallback string) string {
	if override != "" {
		return override
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
