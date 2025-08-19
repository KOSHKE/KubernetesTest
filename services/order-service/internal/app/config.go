package app

import (
	"os"
	"time"
)

type Config struct {
	Port                     string
	DBHost                   string
	DBPort                   string
	DBName                   string
	DBUser                   string
	DBPass                   string
	InventoryServiceURL      string
	DefaultCurrency          string
	InventoryProviderTimeout time.Duration
	KafkaAutoOffsetReset     string
}

func LoadConfigFromEnv() *Config {
	timeout := 3 * time.Second
	if v := os.Getenv("INVENTORY_PROVIDER_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			timeout = d
		}
	}
	return &Config{
		Port:                     getEnv("PORT", "50052"),
		DBHost:                   getEnv("DB_HOST", "localhost"),
		DBPort:                   getEnv("DB_PORT", "5432"),
		DBName:                   getEnv("DB_NAME", "orderdb"),
		DBUser:                   getEnv("DB_USER", "admin"),
		DBPass:                   getEnv("DB_PASSWORD", "password"),
		InventoryServiceURL:      getEnv("INVENTORY_SERVICE_URL", "inventory-service:50053"),
		DefaultCurrency:          getEnv("DEFAULT_CURRENCY", "USD"),
		InventoryProviderTimeout: timeout,
		KafkaAutoOffsetReset:     getEnv("KAFKA_AUTO_OFFSET_RESET", "earliest"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
