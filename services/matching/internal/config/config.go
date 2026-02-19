package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pubsub"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
)

type Config struct {
	config.BaseConfig
	Kafka *pubsub.KafkaConfig `json:"kafka"`
}

func DefaultConfig() *Config {
	base := config.DefaultBaseConfig()
	base.ServiceName = "matching"
	return &Config{
		BaseConfig: base,
		Kafka:      pubsub.DefaultKafkaConfig(),
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
