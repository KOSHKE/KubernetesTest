package app

import (
	"os"
	"strconv"
)

type Config struct {
	Port                  string
	DBHost                string
	DBPort                string
	DBName                string
	DBUser                string
	DBPass                string
	ReservationTTLSeconds int
}

func LoadConfigFromEnv() *Config {
	return &Config{
		Port:                  getEnv("PORT", "50053"),
		DBHost:                getEnv("DB_HOST", "localhost"),
		DBPort:                getEnv("DB_PORT", "5432"),
		DBName:                getEnv("DB_NAME", "inventorydb"),
		DBUser:                getEnv("DB_USER", "admin"),
		DBPass:                getEnv("DB_PASSWORD", "password"),
		ReservationTTLSeconds: getEnvInt("RESERVATION_TTL_SECONDS", 900),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
