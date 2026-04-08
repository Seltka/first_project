package config

import (
	"os"
)

type Config struct {
	DBConn     string
	ServerPort string
}

func Load() *Config {
	return &Config{
		DBConn:     getEnv("DB_CONN", "postgres://postgres:password@localhost:5432/downloader?sslmode=disable"),
		ServerPort: getEnv("SERVER_PORT", ":8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
