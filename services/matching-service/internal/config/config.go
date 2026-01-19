package config

import (
	"encoding/json"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/logging"
)

type Config struct {
	Server  config.ServerConfig   `json:"server"`
	Logging logging.LoggingConfig `json:"logging"`
}

func DefaultConfig() *Config {
	return &Config{
		Server:  config.DefaultServerConfig(),
		Logging: logging.DefaultLoggingConfig(),
	}
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	// 1. Read JSON file first
	// Use os.ReadFile to get the bytes
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	// 2. Parse ENVs second (The Overrider)
	// This looks for the tags like env:"SERVER_PORT"
	// If found, it replaces whatever came from the JSON
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
