package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
)

type Config struct {
	config.BaseConfig
}

func DefaultConfig() *Config {
	base := config.DefaultBaseConfig()
	base.ServiceName = "User Service"
	return &Config{
		BaseConfig: base,
	}
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	cfg, err := config.LoadGeneric(path, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
