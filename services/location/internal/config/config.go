package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pubsub"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type Config struct {
	config.BaseConfig
	Kafka *pubsub.KafkaConfig `json:"kafka"`
	Redis rdstore.RedisConfig `json:"redis"`
	Logic LogicConfig         `json:"logic"`
}

type LogicConfig struct {
	LocationTTLSeconds int `json:"location_ttl_seconds" env:"LOCATION_TTL"`
	TopKNearby         int `json:"top_k_nearby" env:"TOP_K_NEARBY"`
	// Note: The min and max radius could be moved to the database and configured by region
	MinRadiusSearchKm float32 `json:"min_radius_search_km" env:"MIN_RADIUS_SEARCH_KM"`
	MaxRadiusSearchKm float32 `json:"max_radius_search_km" env:"MAX_RADIUS_SEARCH_KM"`
}

func (c *Config) Validate() error {
	if c.Logic.LocationTTLSeconds <= 0 {
		return errors.NewValidationErrorf("location_ttl_seconds must be greater than 0")
	}
	if c.Logic.TopKNearby <= 0 {
		return errors.NewValidationErrorf("top_k_nearby must be greater than 0")
	}
	if c.Logic.MinRadiusSearchKm <= 0 {
		return errors.NewValidationErrorf("min_radius_search_km must be greater than 0")
	}
	if c.Logic.MaxRadiusSearchKm <= 0 {
		return errors.NewValidationErrorf("max_radius_search_km must be greater than 0")
	}
	if c.Logic.MinRadiusSearchKm > c.Logic.MaxRadiusSearchKm {
		return errors.NewValidationErrorf("min_radius_search_km cannot be greater than max_radius_search_km")
	}
	return nil
}

func DefaultConfig() *Config {
	base := config.DefaultBaseConfig()
	base.ServiceName = "location"
	return &Config{
		BaseConfig: base,
		Logic: LogicConfig{
			LocationTTLSeconds: 300,
			TopKNearby:         10,
			MinRadiusSearchKm:  0.5,
			MaxRadiusSearchKm:  5.0},
		Redis: rdstore.DefaultRedisConfig(),
		Kafka: pubsub.DefaultKafkaConfig(),
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
