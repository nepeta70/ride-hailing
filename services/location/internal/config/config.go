package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
)

type Config struct {
	config.BaseConfig
	Redis rdstore.RedisConfig `json:"redis"`
	Logic LogicConfig         `json:"logic"`
}

type LogicConfig struct {
	GeohashPrecision   int `json:"geohash_precision" env:"GEOHASH_PRECISION"`
	LocationTTLSeconds int `json:"location_ttl_seconds" env:"LOCATION_TTL"`
	TopKNearby         int `json:"top_k_nearby" env:"TOP_K_NEARBY"`
}

func DefaultConfig() *Config {
	base := config.DefaultBaseConfig()
	base.ServiceName = "Location Service"
	return &Config{
		BaseConfig: base,
		Logic: LogicConfig{
			GeohashPrecision:   7,
			LocationTTLSeconds: 300,
			TopKNearby:         5,
		},
		Redis: rdstore.DefaultRedisConfig(),
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
