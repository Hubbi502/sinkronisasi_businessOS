package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	ServerPort              string
	DatabaseURL             string
	RedisAddr               string
	GoogleSheetsCredentials string
	GoogleSheetsID          string
	MaintenanceSyncSecret   string
}

// Load reads .env (if present) and populates a Config struct from env vars.
func Load() *Config {
	// Ignore error — .env is optional (e.g. in Docker env vars are injected directly)
	_ = godotenv.Load()

	return &Config{
		ServerPort:              getEnv("SERVER_PORT", "8080"),
		DatabaseURL:             getEnv("DATABASE_URL", "host=localhost user=postgres password=postgres dbname=sinkronisasi_db port=5432 sslmode=disable"),
		RedisAddr:               getEnv("REDIS_ADDR", "localhost:6379"),
		GoogleSheetsCredentials: getEnv("GOOGLE_SHEETS_CREDENTIALS", ""),
		GoogleSheetsID:          getEnv("GOOGLE_SHEETS_ID", ""),
		MaintenanceSyncSecret:   getEnv("MAINTENANCE_SYNC_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
