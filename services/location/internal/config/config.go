package config

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pubsub"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
)

type Config struct {
	config.BaseConfig
	Kafka *pubsub.KafkaConfig `json:"kafka"`
	Redis rdstore.RedisConfig `json:"redis"`
	Logic *LogicConfig        `json:"logic"`
}

func DefaultConfig() *Config {
	base := config.DefaultBaseConfig()
	base.ServiceName = "location"
	return &Config{
		BaseConfig: base,
		Logic:      DefaultLogicConfig(),
		Redis:      rdstore.DefaultRedisConfig(),
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

func (c *Config) Init() error {
	c.Logic.LocationTTL = time.Duration(c.Logic.LocationTTLSeconds) * time.Second
	return nil
}
