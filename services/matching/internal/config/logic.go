package config

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type LogicConfig struct {
	DistanceWeight               float32       `json:"distance_weight" env:"DISTANCE_WEIGHT"`
	AvailabilityWeight           float32       `json:"availability_weight" env:"AVAILABILITY_WEIGHT"`
	MatchingTimeoutMinutes       int           `json:"max_matching_minutes" env:"MAX_MATCHING_MINUTES"`
	MatchingRetryIntervalSeconds int           `json:"matching_retry_interval_seconds" env:"MATCHING_RETRY_INTERVAL_SECONDS"`
	MatchingTimeout              time.Duration `json:"-"`
	MatchingRetryInterval        time.Duration `json:"-"`
}

func (c *LogicConfig) Validate() error {
	if c.DistanceWeight <= 0 {
		return errors.NewValidationErrorf("distance_weight must be positive")
	}
	if c.AvailabilityWeight <= 0 {
		return errors.NewValidationErrorf("availability_weight must be positive")
	}
	if c.MatchingTimeoutMinutes <= 0 {
		return errors.NewValidationErrorf("max_matching_minutes must be positive")
	}
	if c.MatchingRetryIntervalSeconds <= 0 {
		return errors.NewValidationErrorf("matching_retry_interval_seconds must be positive")
	}
	return nil
}

func (c *LogicConfig) Init() error {
	c.MatchingTimeout = time.Duration(c.MatchingTimeoutMinutes) * time.Minute
	c.MatchingRetryInterval = time.Duration(c.MatchingRetryIntervalSeconds) * time.Second
	return nil
}

func DefaultLogicConfig() *LogicConfig {
	cfg := &LogicConfig{
		DistanceWeight:               4.0,
		AvailabilityWeight:           6.0,
		MatchingTimeoutMinutes:       3,
		MatchingRetryIntervalSeconds: 30,
	}
	cfg.Init()
	return cfg
}
