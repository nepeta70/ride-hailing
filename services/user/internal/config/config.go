package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pubsub"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
)

type Config struct {
	config.BaseConfig
	Redis    rdstore.RedisConfig    `json:"redis"`
	Postgres pgstore.PostgresConfig `json:"postgres"`
	Kafka    *pubsub.KafkaConfig    `json:"kafka"`
}

func DefaultConfig() *Config {
	base := config.DefaultBaseConfig()
	base.ServiceName = "user"
	return &Config{
		BaseConfig: base,
		Redis:      rdstore.DefaultRedisConfig(),
		Postgres:   pgstore.DefaultPostgresConfig(),
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
