package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/redis"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
)

type KeysConfig struct {
	GoogleMapsAPIKey string `json:"google_maps_api_key" env:"GOOGLE_MAPS_API_KEY"`
}

func (rc *KeysConfig) Validate() error {

	return nil
}

func DefaultKeysConfig() KeysConfig {
	return KeysConfig{
		GoogleMapsAPIKey: "",
	}
}

type Config struct {
	config.BaseConfig
	KeysConfig KeysConfig        `json:"keys"`
	FareConfig FareConfig        `json:"fare"`
	Redis      redis.RedisConfig `json:"redis"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		BaseConfig: config.DefaultBaseConfig(),
		KeysConfig: DefaultKeysConfig(),
		FareConfig: DefaultFareConfig(),
		Redis:      redis.DefaultRedisConfig(),
	}

	cfg, err := config.LoadGeneric(path, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
