package app

import "os"

type Config struct {
	Port   string
	DBHost string
	DBPort string
	DBName string
	DBUser string
	DBPass string
}

func LoadConfigFromEnv() *Config {
	return &Config{
		Port:   getEnv("PORT", "50053"),
		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "5432"),
		DBName: getEnv("DB_NAME", "inventorydb"),
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
