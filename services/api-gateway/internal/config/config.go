package config

import (
	"ecommerce-platform/pkg/config"
)

// Config wraps the common API Gateway configuration
type Config struct {
	*config.APIGatewayConfig
}

// Load loads API Gateway configuration
func Load() (*Config, error) {
	cfg, err := config.LoadAPIGatewayConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		APIGatewayConfig: cfg,
	}, nil
}
