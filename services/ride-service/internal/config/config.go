package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
)

type RideConfig struct {
	GoogleMapsAPIKey string `json:"google_maps_api_key"`
}

type Config struct {
	config.BaseConfig
	RideConfig RideConfig `json:"ride"`
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
