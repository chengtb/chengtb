package config

import (
	"os"
	"strconv"
)

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

func Load() *Config {
	dispatchInterval, _ := strconv.Atoi(getEnv("DISPATCH_INTERVAL_SEC", "30"))
	dispatchWindow, _ := strconv.Atoi(getEnv("DISPATCH_WINDOW_MIN", "5"))

	return &Config{
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "3306"),
		DBUser:              getEnv("DB_USER", "root"),
		DBPassword:          getEnv("DB_PASSWORD", ""),
		DBName:              getEnv("DB_NAME", "restaurant"),
		ServerPort:          getEnv("SERVER_PORT", "8080"),
		DispatchIntervalSec: dispatchInterval,
		DispatchWindowMin:   dispatchWindow,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
