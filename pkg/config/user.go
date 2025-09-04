package config

import (
	"fmt"
	"time"
)

// UserConfig extends BaseConfig with user-specific configuration
type UserConfig struct {
	*BaseConfig

	// Authentication configuration
	Auth AuthConfig

	// User service specific configuration
	User UserServiceConfig
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	StorageURL         string
	StorageTimeout     time.Duration
}

// UserServiceConfig holds user service specific configuration
type UserServiceConfig struct {
	MaxPasswordLength int
	MinPasswordLength int
	MaxNameLength     int
	MaxPhoneLength    int
	MaxEmailLength    int
}

// LoadUserConfig loads user service configuration from environment variables
func LoadUserConfig() (*UserConfig, error) {
	// Load base configuration
	baseCfg := LoadBaseConfig("user-service")

	// Override default ports for user service
	baseCfg.Port = getEnv("USER_SERVICE_PORT", "50051")
	baseCfg.MetricsPort = getEnv("USER_SERVICE_METRICS_PORT", "9090")

	cfg := &UserConfig{
		BaseConfig: baseCfg,
	}

	// Load authentication configuration
	cfg.loadAuthConfig()

	// Load user service configuration
	cfg.loadUserServiceConfig()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("user configuration validation failed: %w", err)
	}

	return cfg, nil
}

// loadAuthConfig loads authentication configuration
func (cfg *UserConfig) loadAuthConfig() {
	cfg.Auth.AccessTokenSecret = getEnv("JWT_ACCESS_SECRET", "")
	cfg.Auth.RefreshTokenSecret = getEnv("JWT_REFRESH_SECRET", "")
	cfg.Auth.AccessTokenTTL = getEnvAsDuration("ACCESS_TOKEN_TTL", 15*time.Minute)
	cfg.Auth.RefreshTokenTTL = getEnvAsDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour)
	cfg.Auth.StorageURL = getEnv("STORAGE_URL", "redis://localhost:6379")
	cfg.Auth.StorageTimeout = getEnvAsDuration("STORAGE_TIMEOUT", 5*time.Second)
}

// loadUserServiceConfig loads user service specific configuration
func (cfg *UserConfig) loadUserServiceConfig() {
	cfg.User.MaxPasswordLength = getEnvAsInt("USER_MAX_PASSWORD_LENGTH", 128)
	cfg.User.MinPasswordLength = getEnvAsInt("USER_MIN_PASSWORD_LENGTH", 8)
	cfg.User.MaxNameLength = getEnvAsInt("USER_MAX_NAME_LENGTH", 50)
	cfg.User.MaxPhoneLength = getEnvAsInt("USER_MAX_PHONE_LENGTH", 50)
	cfg.User.MaxEmailLength = getEnvAsInt("USER_MAX_EMAIL_LENGTH", 255)
}

// Validate validates the user configuration
func (cfg *UserConfig) Validate() error {
	// Validate base configuration
	if err := cfg.BaseConfig.Validate(); err != nil {
		return fmt.Errorf("base config validation failed: %w", err)
	}

	// Validate authentication configuration
	if cfg.Auth.AccessTokenSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET is required")
	}

	if cfg.Auth.RefreshTokenSecret == "" {
		return fmt.Errorf("JWT_REFRESH_SECRET is required")
	}

	if cfg.Auth.AccessTokenTTL <= 0 {
		return fmt.Errorf("ACCESS_TOKEN_TTL must be positive")
	}

	if cfg.Auth.RefreshTokenTTL <= 0 {
		return fmt.Errorf("REFRESH_TOKEN_TTL must be positive")
	}

	if cfg.Auth.StorageURL == "" {
		return fmt.Errorf("STORAGE_URL is required")
	}

	if cfg.Auth.StorageTimeout <= 0 {
		return fmt.Errorf("STORAGE_TIMEOUT must be positive")
	}

	// Validate user service configuration
	if cfg.User.MaxPasswordLength <= 0 {
		return fmt.Errorf("USER_MAX_PASSWORD_LENGTH must be positive")
	}

	if cfg.User.MinPasswordLength <= 0 {
		return fmt.Errorf("USER_MIN_PASSWORD_LENGTH must be positive")
	}

	if cfg.User.MaxPasswordLength < cfg.User.MinPasswordLength {
		return fmt.Errorf("USER_MAX_PASSWORD_LENGTH must be greater than or equal to USER_MIN_PASSWORD_LENGTH")
	}

	if cfg.User.MaxNameLength <= 0 {
		return fmt.Errorf("USER_MAX_NAME_LENGTH must be positive")
	}

	if cfg.User.MaxPhoneLength <= 0 {
		return fmt.Errorf("USER_MAX_PHONE_LENGTH must be positive")
	}

	if cfg.User.MaxEmailLength <= 0 {
		return fmt.Errorf("USER_MAX_EMAIL_LENGTH must be positive")
	}

	return nil
}

// GetAuthConfig returns authentication configuration
func (cfg *UserConfig) GetAuthConfig() AuthConfig {
	return cfg.Auth
}

// GetUserServiceConfig returns user service configuration
func (cfg *UserConfig) GetUserServiceConfig() UserServiceConfig {
	return cfg.User
}

// IsDevelopment returns true if running in development environment
func (cfg *UserConfig) IsDevelopment() bool {
	return cfg.BaseConfig.IsDevelopment()
}

// IsProduction returns true if running in production environment
func (cfg *UserConfig) IsProduction() bool {
	return cfg.BaseConfig.IsProduction()
}
