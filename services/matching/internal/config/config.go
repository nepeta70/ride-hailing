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
			APIKey:                 "locationsupersecretapikey",
			SenderID:               "b4c12247-9a8f-4926-8508-3619ff6e7af7",
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
