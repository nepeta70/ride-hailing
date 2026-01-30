package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/redis"
)

type Config struct {
	config.BaseConfig
	Redis redis.RedisConfig `json:"redis"`
	Logic LogicConfig       `json:"logic"`
}

type LogicConfig struct {
	GeohashPrecision   int `json:"geohash_precision" env:"GEOHASH_PRECISION"`
	LocationTTLSeconds int `json:"location_ttl_seconds" env:"LOCATION_TTL"`
	TopKNearby         int `json:"top_k_nearby" env:"TOP_K_NEARBY"`
}

func DefaultConfig() *Config {
	return &Config{
		BaseConfig: config.DefaultBaseConfig(),
		Logic: LogicConfig{
			GeohashPrecision:   7,
			LocationTTLSeconds: 300,
			TopKNearby:         5,
		},
		Redis: redis.DefaultRedisConfig(),
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
