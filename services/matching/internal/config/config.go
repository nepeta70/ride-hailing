package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pubsub"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
)

type Config struct {
	config.BaseConfig
	Kafka           *pubsub.KafkaConfig    `json:"kafka"`
	LocationService *LocationServiceConfig `json:"location_service"`
}

func DefaultConfig() *Config {
	base := config.DefaultBaseConfig()
	base.ServiceName = "matching"
	return &Config{
		BaseConfig: base,
		Kafka:      pubsub.DefaultKafkaConfig(),
		LocationService: &LocationServiceConfig{
			LocationServiceAddress: "localhost:50051",
		},
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

type LocationServiceConfig struct {
	LocationServiceAddress string `json:"location_service_address" env:"LOCATION_SERVICE_ADDRESS"`
	APIKey                 string `json:"api_key" env:"LOCATION_SERVICE_API_KEY"`
}
