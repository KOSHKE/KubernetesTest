package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"ecommerce-platform/pkg/config"
)

// InventoryConfig represents inventory-specific configuration
type InventoryConfig struct {
	Port                  string
	MetricsPort           string
	DefaultCurrency       string
	MaxProductsPerPage    int
	ReservationTTLSeconds int
	SupportedCurrencies   []string
}

// Config represents the inventory service configuration
type Config struct {
	*config.BaseConfig
	Inventory InventoryConfig
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() *Config {
	baseConfig := &config.BaseConfig{
		Database: config.DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "inventory"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Kafka: config.KafkaConfig{
			Brokers: getEnvStringSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		},
		GRPC: config.GRPCConfig{
			Port: getEnvInt("GRPC_PORT", 50052),
		},
		HTTP: config.HTTPConfig{
			Port: getEnvInt("HTTP_PORT", 8082),
		},
	}

	return &Config{
		BaseConfig: baseConfig,
		Inventory: InventoryConfig{
			Port:                  getEnv("INVENTORY_SERVICE_PORT", "50053"),
			MetricsPort:           getEnv("INVENTORY_SERVICE_METRICS_PORT", "9096"),
			DefaultCurrency:       getEnv("INVENTORY_DEFAULT_CURRENCY", "USD"),
			MaxProductsPerPage:    getEnvInt("INVENTORY_MAX_PRODUCTS_PER_PAGE", 100),
			ReservationTTLSeconds: getEnvInt("INVENTORY_RESERVATION_TTL_SECONDS", 900),
			SupportedCurrencies:   getEnvStringSlice("INVENTORY_SUPPORTED_CURRENCIES", []string{"USD", "EUR", "GBP", "RUB"}),
		},
	}
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.Name, c.Database.SSLMode)
}

// GetKafkaBrokers returns the Kafka broker addresses
func (c *Config) GetKafkaBrokers() []string {
	return c.Kafka.Brokers
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
