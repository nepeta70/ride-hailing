package config

import "github.com/nepeta70/ride-hailing/internal/pkg/errors"

type LogicConfig struct {
	DistanceWeight     float32 `json:"distance_weight" env:"DISTANCE_WEIGHT"`
	AvailabilityWeight float32 `json:"availability_weight" env:"AVAILABILITY_WEIGHT"`
}

func (c *LogicConfig) Validate() error {
	if c.DistanceWeight <= 0 {
		return errors.NewValidationErrorf("distance_weight must be positive")
	}
	if c.AvailabilityWeight <= 0 {
		return errors.NewValidationErrorf("availability_weight must be positive")
	}
	return nil
}

func DefaultLogicConfig() *LogicConfig {
	return &LogicConfig{
		DistanceWeight:     4.0,
		AvailabilityWeight: 6.0,
	}
}
