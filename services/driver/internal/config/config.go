package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/mongodb"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
)

type Config struct {
	config.BaseConfig
	Mongo mongodb.MongoConfig `json:"mongo"`
}

func DefaultConfig() *Config {
	base := config.DefaultBaseConfig()
	base.ServiceName = "Driver Service"
	return &Config{
		BaseConfig: base,
		Mongo:      mongodb.DefaultMongoConfig(),
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
