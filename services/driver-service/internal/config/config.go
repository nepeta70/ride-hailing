package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
)

type Config struct {
	config.BaseConfig
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		BaseConfig: config.DefaultBaseConfig(),
	}

	cfg, err := config.LoadGeneric(path, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
