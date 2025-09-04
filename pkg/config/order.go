package config

import (
	"fmt"
	"strings"
	"time"
)

// OrderConfig extends BaseConfig with order-specific configuration
type OrderConfig struct {
	*BaseConfig

	// Order processing configuration
	Order OrderProcessingConfig

	// Inventory service configuration
	Inventory InventoryServiceConfig

	// Order item configuration
	OrderItem OrderItemConfig

	// Database configuration
	Database DatabaseConfig

	// Graceful shutdown configuration
	GracefulShutdown GracefulShutdownConfig
}

// OrderProcessingConfig holds order processing configuration
type OrderProcessingConfig struct {
	DefaultCurrency string
	MaxOrderItems   int
	MaxOrderValue   int64
}

// InventoryServiceConfig holds inventory service configuration
type InventoryServiceConfig struct {
	URL     string
	Timeout time.Duration
}

// OrderItemConfig holds order item configuration
type OrderItemConfig struct {
	MaxQuantity         int32
	MinQuantity         int32
	MaxPrice            int64
	SupportedCurrencies []string
}

// GracefulShutdownConfig holds graceful shutdown configuration
type GracefulShutdownConfig struct {
	Timeout time.Duration
}

// LoadOrderConfig loads order service configuration from environment variables
func LoadOrderConfig() (*OrderConfig, error) {
	// Load base configuration
	baseCfg := LoadBaseConfig("order-service")

	// Override default ports for order service
	baseCfg.Port = getEnv("ORDER_SERVICE_PORT", "50052")
	baseCfg.MetricsPort = getEnv("ORDER_SERVICE_METRICS_PORT", "9095")

	cfg := &OrderConfig{
		BaseConfig: baseCfg,
	}

	// Load order processing configuration
	cfg.loadOrderProcessingConfig()

	// Load inventory service configuration
	cfg.loadInventoryServiceConfig()

	// Load order item configuration
	cfg.loadOrderItemConfig()

	// Load database configuration
	cfg.loadDatabaseConfig()

	// Load graceful shutdown configuration
	cfg.loadGracefulShutdownConfig()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("order configuration validation failed: %w", err)
	}

	return cfg, nil
}

// loadOrderProcessingConfig loads order processing configuration
func (cfg *OrderConfig) loadOrderProcessingConfig() {
	cfg.Order.DefaultCurrency = getEnv("ORDER_DEFAULT_CURRENCY", "USD")
	cfg.Order.MaxOrderItems = getEnvAsInt("ORDER_MAX_ITEMS", 100)
	cfg.Order.MaxOrderValue = int64(getEnvAsInt("ORDER_MAX_VALUE", 1000000)) // 1M in minor units
}

// loadInventoryServiceConfig loads inventory service configuration
func (cfg *OrderConfig) loadInventoryServiceConfig() {
	cfg.Inventory.URL = getEnv("INVENTORY_SERVICE_URL", "inventory-service:50053")
	cfg.Inventory.Timeout = getEnvAsDuration("INVENTORY_PROVIDER_TIMEOUT", 3*time.Second)
}

// loadOrderItemConfig loads order item configuration
func (cfg *OrderConfig) loadOrderItemConfig() {
	cfg.OrderItem.MaxQuantity = int32(getEnvAsInt("ORDER_ITEM_MAX_QUANTITY", 1000))
	cfg.OrderItem.MinQuantity = int32(getEnvAsInt("ORDER_ITEM_MIN_QUANTITY", 1))
	cfg.OrderItem.MaxPrice = int64(getEnvAsInt("ORDER_ITEM_MAX_PRICE", 100000)) // 1K in minor units

	supportedCurrencies := getEnv("ORDER_SUPPORTED_CURRENCIES", "USD,EUR,GBP,RUB")
	cfg.OrderItem.SupportedCurrencies = strings.Split(supportedCurrencies, ",")

	// Trim whitespace from currencies
	for i, currency := range cfg.OrderItem.SupportedCurrencies {
		cfg.OrderItem.SupportedCurrencies[i] = strings.TrimSpace(currency)
	}
}

// loadDatabaseConfig loads database configuration
func (cfg *OrderConfig) loadDatabaseConfig() {
	// Database config is loaded in BaseConfig, but we can override specific values here if needed
}

// loadGracefulShutdownConfig loads graceful shutdown configuration
func (cfg *OrderConfig) loadGracefulShutdownConfig() {
	cfg.GracefulShutdown.Timeout = getEnvAsDuration("ORDER_GRACEFUL_SHUTDOWN_TIMEOUT", 30*time.Second)
}

// Validate validates the order configuration
func (cfg *OrderConfig) Validate() error {
	// Validate base configuration
	if err := cfg.BaseConfig.Validate(); err != nil {
		return fmt.Errorf("base config validation failed: %w", err)
	}

	// Validate order processing configuration
	if cfg.Order.DefaultCurrency == "" {
		return fmt.Errorf("ORDER_DEFAULT_CURRENCY is required")
	}

	if cfg.Order.MaxOrderItems <= 0 {
		return fmt.Errorf("ORDER_MAX_ITEMS must be positive")
	}

	if cfg.Order.MaxOrderValue <= 0 {
		return fmt.Errorf("ORDER_MAX_VALUE must be positive")
	}

	// Validate inventory service configuration
	if cfg.Inventory.URL == "" {
		return fmt.Errorf("INVENTORY_SERVICE_URL is required")
	}

	if cfg.Inventory.Timeout <= 0 {
		return fmt.Errorf("INVENTORY_PROVIDER_TIMEOUT must be positive")
	}

	// Validate order item configuration
	if cfg.OrderItem.MaxQuantity <= 0 {
		return fmt.Errorf("ORDER_ITEM_MAX_QUANTITY must be positive")
	}

	if cfg.OrderItem.MinQuantity <= 0 {
		return fmt.Errorf("ORDER_ITEM_MIN_QUANTITY must be positive")
	}

	if cfg.OrderItem.MaxQuantity < cfg.OrderItem.MinQuantity {
		return fmt.Errorf("ORDER_ITEM_MAX_QUANTITY must be greater than or equal to ORDER_ITEM_MIN_QUANTITY")
	}

	if cfg.OrderItem.MaxPrice <= 0 {
		return fmt.Errorf("ORDER_ITEM_MAX_PRICE must be positive")
	}

	if len(cfg.OrderItem.SupportedCurrencies) == 0 {
		return fmt.Errorf("at least one supported currency is required")
	}

	return nil
}

// IsCurrencySupported checks if the given currency is supported
func (cfg *OrderConfig) IsCurrencySupported(currency string) bool {
	for _, supported := range cfg.OrderItem.SupportedCurrencies {
		if supported == currency {
			return true
		}
	}
	return false
}

// GetInventoryServiceURL returns the inventory service URL
func (cfg *OrderConfig) GetInventoryServiceURL() string {
	return cfg.Inventory.URL
}

// GetInventoryServiceTimeout returns the inventory service timeout
func (cfg *OrderConfig) GetInventoryServiceTimeout() time.Duration {
	return cfg.Inventory.Timeout
}
