package app

import (
	"ecommerce-platform/pkg/config"
)

// Config holds service configuration values.
type Config struct {
	*config.UserConfig
}

// LoadConfigFromEnv loads configuration from environment variables.
func LoadConfigFromEnv() (*Config, error) {
	userCfg, err := config.LoadUserConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		UserConfig: userCfg,
	}, nil
}
