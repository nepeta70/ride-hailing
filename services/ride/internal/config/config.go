package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
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
	KeysConfig KeysConfig             `json:"keys"`
	FareConfig FareConfig             `json:"fare"`
	Redis      rdstore.RedisConfig    `json:"redis"`
	Postgres   pgstore.PostgresConfig `json:"postgres"`
}

func DefaultConfig() *Config {
	base := config.DefaultBaseConfig()
	base.ServiceName = "ride"
	return &Config{
		BaseConfig: base,
		KeysConfig: DefaultKeysConfig(),
		FareConfig: DefaultFareConfig(),
		Redis:      rdstore.DefaultRedisConfig(),
		Postgres:   pgstore.DefaultPostgresConfig(),
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
