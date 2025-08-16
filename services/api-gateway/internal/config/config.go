package config

import (
	"os"
)

type Config struct {
	Port                string
	JWKSURL             string
	JWTIssuer           string
	JWTAudience         string
	RedisURL            string
	UserServiceURL      string
	OrderServiceURL     string
	InventoryServiceURL string
	PaymentServiceURL   string
}

func Load() *Config {
	return &Config{
		Port:                getEnv("PORT", "8080"),
		JWKSURL:             getEnv("JWKS_URL", "http://user-service:8081/.well-known/jwks.json"),
		JWTIssuer:           getEnv("JWT_ISSUER", "user-service"),
		JWTAudience:         getEnv("JWT_AUDIENCE", "api-gateway"),
		RedisURL:            getEnv("REDIS_URL", "redis:6379"),
		UserServiceURL:      getEnv("USER_SERVICE_URL", "localhost:50051"),
		OrderServiceURL:     getEnv("ORDER_SERVICE_URL", "localhost:50052"),
		InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", "localhost:50053"),
		PaymentServiceURL:   getEnv("PAYMENT_SERVICE_URL", "localhost:50054"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
