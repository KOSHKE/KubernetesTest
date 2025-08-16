package app

import "os"

type Config struct{ Port string }

func LoadConfigFromEnv() *Config { return &Config{Port: getEnv("PORT", "50053")} }

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
