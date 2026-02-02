package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
)

type RideConfig struct {
	GoogleMapsAPIKey string `json:"google_maps_api_key" env:"RIDE_GOOGLE_MAPS_API_KEY"`
}

func DefaultRideConfig() RideConfig {
	return RideConfig{
		GoogleMapsAPIKey: "",
	}
}

type Config struct {
	config.BaseConfig
	RideConfig RideConfig `json:"ride"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		BaseConfig: config.DefaultBaseConfig(),
		RideConfig: DefaultRideConfig(),
	}

	cfg, err := config.LoadGeneric(path, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
