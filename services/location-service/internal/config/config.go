package config

import (
	"encoding/json"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/nepeta70/ride-hailing/internal/pkg/logging"
	"github.com/nepeta70/ride-hailing/internal/pkg/redis"
)

type Config struct {
	Server  ServerConfig          `json:"server"`
	Logging logging.LoggingConfig `json:"logging"`
	Redis   redis.RedisConfig     `json:"redis"`
	Logic   LogicConfig           `json:"logic"`
}

type ServerConfig struct {
	Port int    `json:"port" env:"SERVER_PORT" envDefault:"50051"`
	Host string `json:"host" env:"SERVER_HOST" envDefault:"127.0.0.1"`
}

type LogicConfig struct {
	GeohashPrecision   int `json:"geohash_precision" env:"GEOHASH_PRECISION" envDefault:"7"`
	LocationTTLSeconds int `json:"location_ttl_seconds" env:"LOCATION_TTL" envDefault:"60"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 50051,
			Host: "127.0.0.1",
		},
		Logic: LogicConfig{
			GeohashPrecision:   7,
			LocationTTLSeconds: 60,
		},
		Logging: logging.DefaultLoggingConfig(),
		Redis:   redis.DefaultRedisConfig(),
	}
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	// 1. Read JSON file first
	// Use os.ReadFile to get the bytes
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	// 2. Parse ENVs second (The Overrider)
	// This looks for the tags like env:"SERVER_PORT"
	// If found, it replaces whatever came from the JSON
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
