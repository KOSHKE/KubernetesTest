package app

import "os"

type Config struct {
	Port                string
	DBHost              string
	DBPort              string
	DBName              string
	DBUser              string
	DBPass              string
	InventoryServiceURL string
}

func LoadConfigFromEnv() *Config {
	return &Config{
		Port:                getEnv("PORT", "50052"),
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "5432"),
		DBName:              getEnv("DB_NAME", "orderdb"),
		DBUser:              getEnv("DB_USER", "admin"),
		DBPass:              getEnv("DB_PASSWORD", "password"),
		InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", "inventory-service:50053"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
