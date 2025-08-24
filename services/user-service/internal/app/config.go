package app

import (
	"os"
	"time"
)

// Config holds service configuration values.
type Config struct {
	Port             string
	DBHost           string
	DBPort           string
	DBName           string
	DBUser           string
	DBPass           string
	RedisURL         string
	JWTSecret        string
	JWTRefreshSecret string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	MetricsPort      string
}

// LoadConfigFromEnv loads configuration from environment variables.
func LoadConfigFromEnv() *Config {
	return &Config{
		Port:             getEnv("HTTP_PORT", "50051"),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBName:           getEnv("DB_NAME", "userdb"),
		DBUser:           getEnv("DB_USER", "admin"),
		DBPass:           getEnv("DB_PASSWORD", "password"),
		RedisURL:         getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key"),
		AccessTokenTTL:   getEnvAsDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:  getEnvAsDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		MetricsPort:      getEnv("METRICS_PORT", "9090"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvAsDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if duration, err := time.ParseDuration(v); err == nil {
			return duration
		}
	}
	return def
}
