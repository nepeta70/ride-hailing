package config

import (
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
	Port int    `json:"port"`
	Host string `json:"host"`
}

type LogicConfig struct {
	GeohashPrecision   int `json:"geohash_precision"`
	LocationTTLSeconds int `json:"location_ttl_seconds"`
}
