package config

import (
	"os"
)

type Config struct {
	Port                string
	RedisURL            string
	FrontendOrigins     string
	UserServiceURL      string
	OrderServiceURL     string
	InventoryServiceURL string
	PaymentServiceURL   string
	JWTSecret           string
	JWTRefreshSecret    string
}

func Load() *Config {
	return &Config{
		Port:                getEnv("PORT", "8080"),
		RedisURL:            getEnv("REDIS_URL", "redis:6379"),
		FrontendOrigins:     getEnv("FRONTEND_ORIGINS", "http://localhost:3001"),
		UserServiceURL:      getEnv("USER_SERVICE_URL", "localhost:50051"),
		OrderServiceURL:     getEnv("ORDER_SERVICE_URL", "localhost:50052"),
		InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", "localhost:50053"),
		PaymentServiceURL:   getEnv("PAYMENT_SERVICE_URL", "localhost:50054"),
		JWTSecret:           getEnv("JWT_SECRET", "your-secret-key"),
		JWTRefreshSecret:    getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
