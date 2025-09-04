package config

import (
	"fmt"
	"strings"
	"time"
)

// PaymentConfig extends BaseConfig with payment-specific configuration
type PaymentConfig struct {
	*BaseConfig

	// Redis configuration
	Redis RedisConfig

	// Payment processing configuration
	Payment PaymentProcessingConfig

	// Order totals configuration
	OrderTotals OrderTotalsConfig
}

// RedisConfig holds Redis-related configuration
type RedisConfig struct {
	Addr         string
	DB           int
	Password     string
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// PaymentProcessingConfig holds payment processing configuration
type PaymentProcessingConfig struct {
	ProcessTimeout   time.Duration
	PublishTimeout   time.Duration
	MaxRetries       int
	RetryDelay       time.Duration
	DefaultCurrency  string
	SupportedMethods []string
}

// OrderTotalsConfig holds order totals configuration
type OrderTotalsConfig struct {
	TTL time.Duration
}

// LoadPaymentConfig loads payment service configuration from environment variables
func LoadPaymentConfig() (*PaymentConfig, error) {
	// Load base configuration
	baseCfg := LoadBaseConfig("payment-service")

	// Override default ports for payment service
	baseCfg.Port = getEnv("PAYMENT_SERVICE_PORT", "50054")
	baseCfg.MetricsPort = getEnv("PAYMENT_SERVICE_METRICS_PORT", "9097")

	cfg := &PaymentConfig{
		BaseConfig: baseCfg,
	}

	// Load Redis configuration
	cfg.loadRedisConfig()

	// Load payment processing configuration
	cfg.loadPaymentProcessingConfig()

	// Load order totals configuration
	cfg.loadOrderTotalsConfig()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("payment configuration validation failed: %w", err)
	}

	return cfg, nil
}

// loadRedisConfig loads Redis-related configuration
func (cfg *PaymentConfig) loadRedisConfig() {
	// Support both REDIS_URL and REDIS_ADDR for backward compatibility
	redisURL := getEnv("REDIS_URL", "")
	redisAddr := getEnv("REDIS_ADDR", "")

	if redisURL != "" {
		// Parse REDIS_URL format (redis://host:port/db)
		if strings.HasPrefix(redisURL, "redis://") {
			cfg.Redis.Addr = strings.TrimPrefix(redisURL, "redis://")
		} else {
			cfg.Redis.Addr = redisURL
		}
	} else if redisAddr != "" {
		cfg.Redis.Addr = redisAddr
	} else {
		cfg.Redis.Addr = "redis:6379"
	}

	cfg.Redis.DB = getEnvAsInt("REDIS_DB", 0)
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.PoolSize = getEnvAsInt("REDIS_POOL_SIZE", 10)
	cfg.Redis.MinIdleConns = getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5)
	cfg.Redis.DialTimeout = getEnvAsDuration("REDIS_DIAL_TIMEOUT", 5*time.Second)
	cfg.Redis.ReadTimeout = getEnvAsDuration("REDIS_READ_TIMEOUT", 3*time.Second)
	cfg.Redis.WriteTimeout = getEnvAsDuration("REDIS_WRITE_TIMEOUT", 3*time.Second)
}

// loadPaymentProcessingConfig loads payment processing configuration
func (cfg *PaymentConfig) loadPaymentProcessingConfig() {
	cfg.Payment.ProcessTimeout = getEnvAsDuration("PAYMENT_PROCESS_TIMEOUT", 30*time.Second)
	cfg.Payment.PublishTimeout = getEnvAsDuration("KAFKA_PUBLISH_TIMEOUT", 10*time.Second)
	cfg.Payment.MaxRetries = getEnvAsInt("PAYMENT_MAX_RETRIES", 3)
	cfg.Payment.RetryDelay = getEnvAsDuration("PAYMENT_RETRY_DELAY", 1*time.Second)
	cfg.Payment.DefaultCurrency = getEnv("PAYMENT_DEFAULT_CURRENCY", "USD")

	supportedMethods := getEnv("PAYMENT_SUPPORTED_METHODS", "CREDIT_CARD,DEBIT_CARD,BANK_TRANSFER")
	cfg.Payment.SupportedMethods = strings.Split(supportedMethods, ",")

	// Trim whitespace from methods
	for i, method := range cfg.Payment.SupportedMethods {
		cfg.Payment.SupportedMethods[i] = strings.TrimSpace(method)
	}
}

// loadOrderTotalsConfig loads order totals configuration
func (cfg *PaymentConfig) loadOrderTotalsConfig() {
	cfg.OrderTotals.TTL = getEnvAsDuration("ORDER_TOTAL_TTL", 30*time.Minute)
}

// Validate validates the payment configuration
func (cfg *PaymentConfig) Validate() error {
	// Validate base configuration
	if err := cfg.BaseConfig.Validate(); err != nil {
		return fmt.Errorf("base config validation failed: %w", err)
	}

	// Validate Redis configuration
	if cfg.Redis.Addr == "" {
		return fmt.Errorf("Redis address is required")
	}

	// Validate payment configuration
	if cfg.Payment.ProcessTimeout <= 0 {
		return fmt.Errorf("PAYMENT_PROCESS_TIMEOUT must be positive")
	}

	if cfg.Payment.PublishTimeout <= 0 {
		return fmt.Errorf("KAFKA_PUBLISH_TIMEOUT must be positive")
	}

	if cfg.OrderTotals.TTL <= 0 {
		return fmt.Errorf("ORDER_TOTAL_TTL must be positive")
	}

	return nil
}

// GetRedisAddr returns the Redis address
func (cfg *PaymentConfig) GetRedisAddr() string {
	return cfg.Redis.Addr
}
