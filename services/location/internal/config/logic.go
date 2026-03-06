package config

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type LogicConfig struct {
	LocationTTLSeconds int `json:"location_ttl_seconds" env:"LOCATION_TTL"`
	TopKNearby         int `json:"top_k_nearby" env:"TOP_K_NEARBY"`
	// Note: The min and max radius could be moved to the database and configured by region
	MinRadiusSearchKm float32 `json:"min_radius_search_km" env:"MIN_RADIUS_SEARCH_KM"`
	MaxRadiusSearchKm float32 `json:"max_radius_search_km" env:"MAX_RADIUS_SEARCH_KM"`

	LocationTTL time.Duration `json:"-"` // Calculated field based on LocationTTLSeconds
}

func (cfg *LogicConfig) Init() error {
	cfg.LocationTTL = time.Duration(cfg.LocationTTLSeconds) * time.Second
	return nil
}

func (cfg *LogicConfig) Validate() error {
	if cfg.LocationTTLSeconds <= 0 {
		return errors.NewValidationErrorf("location TTL must be greater than zero")
	}
	if cfg.TopKNearby <= 0 {
		return errors.NewValidationErrorf("top K nearby must be greater than zero")
	}
	if cfg.MinRadiusSearchKm <= 0 {
		return errors.NewValidationErrorf("min radius search must be greater than zero")
	}
	if cfg.MaxRadiusSearchKm <= 0 {
		return errors.NewValidationErrorf("max radius search must be greater than zero")
	}
	if cfg.MinRadiusSearchKm > cfg.MaxRadiusSearchKm {
		return errors.NewValidationErrorf("min radius search cannot be greater than max radius search")
	}
	return nil
}

func DefaultLogicConfig() *LogicConfig {
	cfg := &LogicConfig{
		LocationTTLSeconds: 300,
		TopKNearby:         10,
		MinRadiusSearchKm:  0.5,
		MaxRadiusSearchKm:  5.0}
	cfg.Init()
	return cfg
}
