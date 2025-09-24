package config

import (
	"fmt"
	"strings"
	"time"
)

// APIGatewayConfig extends BaseConfig with API Gateway specific configuration
type APIGatewayConfig struct {
	*BaseConfig

	// JWT configuration
	JWT JWTConfig

	// Service URLs configuration
	Services ServiceURLsConfig

	// CORS configuration
	CORS CORSConfig

	// Rate limiting configuration
	RateLimit RateLimitConfig

	// Graceful shutdown configuration
	GracefulShutdown GracefulShutdownConfig
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	Issuer        string
	Audience      string
	Algorithm     string
}

// ServiceURLsConfig holds service URLs configuration
type ServiceURLsConfig struct {
	UserServiceURL      string
	OrderServiceURL     string
	InventoryServiceURL string
	PaymentServiceURL   string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool
	RequestsPerMinute int
	BurstSize         int
}

// LoadAPIGatewayConfig loads API Gateway configuration from environment variables
func LoadAPIGatewayConfig() (*APIGatewayConfig, error) {
	// Load base configuration
	baseCfg := LoadBaseConfig("api-gateway")

	// Override default ports for API Gateway
	baseCfg.Port = getEnv("API_GATEWAY_PORT", "8080")
	baseCfg.MetricsPort = getEnv("API_GATEWAY_METRICS_PORT", "8081")

	cfg := &APIGatewayConfig{
		BaseConfig: baseCfg,
	}

	// Load JWT configuration
	cfg.loadJWTConfig()

	// Load service URLs configuration
	cfg.loadServiceURLsConfig()

	// Load CORS configuration
	cfg.loadCORSConfig()

	// Load rate limiting configuration
	cfg.loadRateLimitConfig()

	// Load graceful shutdown configuration
	cfg.loadGracefulShutdownConfig()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("API Gateway configuration validation failed: %w", err)
	}

	return cfg, nil
}

// loadJWTConfig loads JWT configuration
func (cfg *APIGatewayConfig) loadJWTConfig() {
	cfg.JWT.AccessSecret = getEnv("JWT_ACCESS_SECRET", "your-access-secret-key")
	cfg.JWT.RefreshSecret = getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key")
	cfg.JWT.AccessTTL = getEnvAsDuration("JWT_ACCESS_TTL", 15*time.Minute)
	cfg.JWT.RefreshTTL = getEnvAsDuration("JWT_REFRESH_TTL", 7*24*time.Hour)
	cfg.JWT.Issuer = getEnv("JWT_ISSUER", "api-gateway")
	cfg.JWT.Audience = getEnv("JWT_AUDIENCE", "ecommerce-platform")
	cfg.JWT.Algorithm = getEnv("JWT_ALGORITHM", "HS256")
}

// loadServiceURLsConfig loads service URLs configuration
func (cfg *APIGatewayConfig) loadServiceURLsConfig() {
	cfg.Services.UserServiceURL = getEnv("USER_SERVICE_URL", "user-service:50051")
	cfg.Services.OrderServiceURL = getEnv("ORDER_SERVICE_URL", "order-service:50052")
	cfg.Services.InventoryServiceURL = getEnv("INVENTORY_SERVICE_URL", "inventory-service:50053")
	cfg.Services.PaymentServiceURL = getEnv("PAYMENT_SERVICE_URL", "payment-service:50054")
}

// loadCORSConfig loads CORS configuration
func (cfg *APIGatewayConfig) loadCORSConfig() {
	// Parse allowed origins
	origins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001")
	cfg.CORS.AllowedOrigins = splitAndTrim(origins, ",")

	// Parse allowed methods
	methods := getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS")
	cfg.CORS.AllowedMethods = splitAndTrim(methods, ",")

	// Parse allowed headers
	headers := getEnv("CORS_ALLOWED_HEADERS", "Authorization,Content-Type,Accept,Origin,X-Requested-With")
	cfg.CORS.AllowedHeaders = splitAndTrim(headers, ",")

	cfg.CORS.AllowCredentials = getEnvAsBool("CORS_ALLOW_CREDENTIALS", true)
	cfg.CORS.MaxAge = getEnvAsDuration("CORS_MAX_AGE", 12*time.Hour)
}

// loadRateLimitConfig loads rate limiting configuration
func (cfg *APIGatewayConfig) loadRateLimitConfig() {
	cfg.RateLimit.Enabled = getEnvAsBool("RATE_LIMIT_ENABLED", false)
	cfg.RateLimit.RequestsPerMinute = getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100)
	cfg.RateLimit.BurstSize = getEnvAsInt("RATE_LIMIT_BURST_SIZE", 10)
}

// loadGracefulShutdownConfig loads graceful shutdown configuration
func (cfg *APIGatewayConfig) loadGracefulShutdownConfig() {
	cfg.GracefulShutdown.Enabled = getEnvAsBool("GRACEFUL_SHUTDOWN_ENABLED", true)
	cfg.GracefulShutdown.Timeout = getEnvAsDuration("GRACEFUL_SHUTDOWN_TIMEOUT", 30*time.Second)
	cfg.GracefulShutdown.WaitTimeout = getEnvAsDuration("GRACEFUL_SHUTDOWN_WAIT_TIMEOUT", 10*time.Second)
}

// Validate validates the API Gateway configuration
func (cfg *APIGatewayConfig) Validate() error {
	// Validate base configuration
	if err := cfg.BaseConfig.Validate(); err != nil {
		return fmt.Errorf("base config validation failed: %w", err)
	}

	// Validate JWT configuration
	if cfg.JWT.AccessSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET is required")
	}

	if cfg.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT_REFRESH_SECRET is required")
	}

	if cfg.JWT.AccessTTL <= 0 {
		return fmt.Errorf("JWT_ACCESS_TTL must be positive")
	}

	if cfg.JWT.RefreshTTL <= 0 {
		return fmt.Errorf("JWT_REFRESH_TTL must be positive")
	}

	if cfg.JWT.Issuer == "" {
		return fmt.Errorf("JWT_ISSUER is required")
	}

	if cfg.JWT.Audience == "" {
		return fmt.Errorf("JWT_AUDIENCE is required")
	}

	// Validate service URLs
	if cfg.Services.UserServiceURL == "" {
		return fmt.Errorf("USER_SERVICE_URL is required")
	}

	if cfg.Services.OrderServiceURL == "" {
		return fmt.Errorf("ORDER_SERVICE_URL is required")
	}

	if cfg.Services.InventoryServiceURL == "" {
		return fmt.Errorf("INVENTORY_SERVICE_URL is required")
	}

	if cfg.Services.PaymentServiceURL == "" {
		return fmt.Errorf("PAYMENT_SERVICE_URL is required")
	}

	// Validate CORS configuration
	if len(cfg.CORS.AllowedOrigins) == 0 {
		return fmt.Errorf("at least one CORS allowed origin is required")
	}

	if len(cfg.CORS.AllowedMethods) == 0 {
		return fmt.Errorf("at least one CORS allowed method is required")
	}

	if len(cfg.CORS.AllowedHeaders) == 0 {
		return fmt.Errorf("at least one CORS allowed header is required")
	}

	// Validate rate limiting configuration
	if cfg.RateLimit.Enabled {
		if cfg.RateLimit.RequestsPerMinute <= 0 {
			return fmt.Errorf("RATE_LIMIT_REQUESTS_PER_MINUTE must be positive when rate limiting is enabled")
		}

		if cfg.RateLimit.BurstSize <= 0 {
			return fmt.Errorf("RATE_LIMIT_BURST_SIZE must be positive when rate limiting is enabled")
		}
	}

	return nil
}

// GetJWTConfig returns JWT configuration for JWT manager
func (cfg *APIGatewayConfig) GetJWTConfig() JWTConfig {
	return cfg.JWT
}

// GetServiceURLs returns service URLs configuration
func (cfg *APIGatewayConfig) GetServiceURLs() ServiceURLsConfig {
	return cfg.Services
}

// GetCORSConfig returns CORS configuration
func (cfg *APIGatewayConfig) GetCORSConfig() CORSConfig {
	return cfg.CORS
}

// GetRateLimitConfig returns rate limiting configuration
func (cfg *APIGatewayConfig) GetRateLimitConfig() RateLimitConfig {
	return cfg.RateLimit
}

// IsCORSOriginAllowed checks if the given origin is allowed
func (cfg *APIGatewayConfig) IsCORSOriginAllowed(origin string) bool {
	for _, allowed := range cfg.CORS.AllowedOrigins {
		if allowed == origin || allowed == "*" {
			return true
		}
	}
	return false
}

// Helper function to split and trim strings
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
