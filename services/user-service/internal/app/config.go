package app

import "os"

// Config holds service configuration values.
type Config struct {
	Port   string
	DBHost string
	DBPort string
	DBName string
	DBUser string
	DBPass string
}

// LoadConfigFromEnv loads configuration from environment variables.
func LoadConfigFromEnv() *Config {
	return &Config{
		Port:   getEnv("PORT", "50051"),
		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "5432"),
		DBName: getEnv("DB_NAME", "orderdb"),
		DBUser: getEnv("DB_USER", "admin"),
		DBPass: getEnv("DB_PASSWORD", "password"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
