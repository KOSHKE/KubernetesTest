package config

import (
	"fmt"
	"strings"
	"time"
)

// InventoryConfig extends BaseConfig with inventory-specific configuration
type InventoryConfig struct {
	*BaseConfig

	// Inventory processing configuration
	Inventory InventoryProcessingConfig

	// Database configuration
	Database DatabaseConfig

	// Graceful shutdown configuration
	GracefulShutdown GracefulShutdownConfig
}

// InventoryProcessingConfig holds inventory processing configuration
type InventoryProcessingConfig struct {
	DefaultCurrency       string
	MaxProductsPerPage    int
	ReservationTTLSeconds int
	SupportedCurrencies   []string
}

// LoadInventoryConfig loads inventory service configuration from environment variables
func LoadInventoryConfig() (*InventoryConfig, error) {
	// Load base configuration
	baseCfg := LoadBaseConfig("inventory-service")

	// Override default ports for inventory service
	baseCfg.Port = getEnv("INVENTORY_SERVICE_PORT", "50053")
	baseCfg.MetricsPort = getEnv("INVENTORY_SERVICE_METRICS_PORT", "9096")

	cfg := &InventoryConfig{
		BaseConfig: baseCfg,
	}

	// Load inventory processing configuration
	cfg.loadInventoryProcessingConfig()

	// Load database configuration
	cfg.loadDatabaseConfig()

	// Load graceful shutdown configuration
	cfg.loadGracefulShutdownConfig()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("inventory configuration validation failed: %w", err)
	}

	return cfg, nil
}

// loadInventoryProcessingConfig loads inventory processing configuration
func (cfg *InventoryConfig) loadInventoryProcessingConfig() {
	cfg.Inventory.DefaultCurrency = getEnv("INVENTORY_DEFAULT_CURRENCY", "USD")
	cfg.Inventory.MaxProductsPerPage = getEnvAsInt("INVENTORY_MAX_PRODUCTS_PER_PAGE", 100)
	cfg.Inventory.ReservationTTLSeconds = getEnvAsInt("INVENTORY_RESERVATION_TTL_SECONDS", 900)

	supportedCurrencies := getEnv("INVENTORY_SUPPORTED_CURRENCIES", "USD,EUR,GBP,RUB")
	cfg.Inventory.SupportedCurrencies = strings.Split(supportedCurrencies, ",")

	// Trim whitespace from currencies
	for i, currency := range cfg.Inventory.SupportedCurrencies {
		cfg.Inventory.SupportedCurrencies[i] = strings.TrimSpace(currency)
	}
}

// loadDatabaseConfig loads database configuration
func (cfg *InventoryConfig) loadDatabaseConfig() {
	// Database config is loaded in BaseConfig, but we can override specific values here if needed
}

// loadGracefulShutdownConfig loads graceful shutdown configuration
func (cfg *InventoryConfig) loadGracefulShutdownConfig() {
	cfg.GracefulShutdown.Timeout = getEnvAsDuration("INVENTORY_GRACEFUL_SHUTDOWN_TIMEOUT", 30*time.Second)
}

// Validate validates the inventory configuration
func (cfg *InventoryConfig) Validate() error {
	// Validate base configuration
	if err := cfg.BaseConfig.Validate(); err != nil {
		return fmt.Errorf("base config validation failed: %w", err)
	}

	// Validate inventory processing configuration
	if cfg.Inventory.DefaultCurrency == "" {
		return fmt.Errorf("INVENTORY_DEFAULT_CURRENCY is required")
	}

	if cfg.Inventory.MaxProductsPerPage <= 0 {
		return fmt.Errorf("INVENTORY_MAX_PRODUCTS_PER_PAGE must be positive")
	}

	if cfg.Inventory.ReservationTTLSeconds <= 0 {
		return fmt.Errorf("INVENTORY_RESERVATION_TTL_SECONDS must be positive")
	}

	if len(cfg.Inventory.SupportedCurrencies) == 0 {
		return fmt.Errorf("at least one supported currency is required")
	}

	return nil
}

// IsCurrencySupported checks if the given currency is supported
func (cfg *InventoryConfig) IsCurrencySupported(currency string) bool {
	for _, supported := range cfg.Inventory.SupportedCurrencies {
		if supported == currency {
			return true
		}
	}
	return false
}
