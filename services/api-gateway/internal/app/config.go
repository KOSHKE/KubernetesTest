package app

import cfgpkg "api-gateway/internal/config"

// Reuse existing config.Config to avoid duplicate types
type Config = cfgpkg.Config

func LoadConfigFromEnv() *Config { return cfgpkg.Load() }
