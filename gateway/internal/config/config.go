package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type BackendServiceConfig struct {
	Address string `json:"address" env:"ADDRESS"`
	APIKey  string `json:"api_key" env:"API_KEY"`
}

func (c *BackendServiceConfig) Validate() error {
	if c.Address == "" {
		return errors.NewValidationErrorf("backend service address is required")
	}
	return nil
}

type ServicesConfig struct {
	Ride     BackendServiceConfig `json:"ride" envPrefix:"RIDE_SERVICE_"`
	User     BackendServiceConfig `json:"user" envPrefix:"USER_SERVICE_"`
	Driver   BackendServiceConfig `json:"driver" envPrefix:"DRIVER_SERVICE_"`
	Location BackendServiceConfig `json:"location" envPrefix:"LOCATION_SERVICE_"`
}

type Config struct {
	config.BaseConfig
	Services ServicesConfig `json:"services"`
}

func DefaultConfig() *Config {
	base := config.DefaultBaseConfig()
	base.ServiceName = "gateway"
	base.Server.Port = 5000
	base.Server.Host = "0.0.0.0"

	return &Config{
		BaseConfig: base,
		Services: ServicesConfig{
			Ride: BackendServiceConfig{
				Address: "127.0.0.1:50052",
			},
			User: BackendServiceConfig{
				Address: "127.0.0.1:50055",
			},
			Driver: BackendServiceConfig{
				Address: "127.0.0.1:50054",
			},
			Location: BackendServiceConfig{
				Address: "127.0.0.1:50051",
			},
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()
	return config.LoadGeneric(path, cfg)
}
