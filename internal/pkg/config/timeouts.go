package config

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type TimeoutsConfig struct {
	ShutdownDelayInSeconds     int `json:"shutdown_timeout_seconds" env:"SHUTDOWN_TIMEOUT_SECONDS"`
	RequestTimeoutSeconds      int `json:"request_timeout_seconds" env:"REQUEST_TIMEOUT_SECONDS"`
	HealthCheckIntervalSeconds int `json:"health_check_interval_seconds" env:"SERVER_HEALTH_CHECK_INTERVAL_SECONDS"`
	// Max allowed clock drift/network delay
	MaxClockDriftSeconds int `json:"max_clock_drift_seconds" env:"MAX_CLOCK_DRIFT_SECONDS"`

	ShutdownTimeout     time.Duration `json:"-"`
	RequestTimeout      time.Duration `json:"-"`
	HealthCheckInterval time.Duration `json:"-"`
}

func DefaultTimeoutsConfig() TimeoutsConfig {
	cfg := TimeoutsConfig{
		ShutdownDelayInSeconds:     10,
		RequestTimeoutSeconds:      5,
		HealthCheckIntervalSeconds: 10,
		MaxClockDriftSeconds:       2,
	}

	cfg.Init()
	return cfg
}

func (c *TimeoutsConfig) Validate() error {
	if c.RequestTimeoutSeconds <= 0 {
		return errors.NewValidationErrorf("request timeout must be greater than zero")
	}
	if c.ShutdownDelayInSeconds < 0 {
		return errors.NewValidationErrorf("shutdown delay cannot be negative")
	}
	if c.HealthCheckIntervalSeconds <= 0 {
		return errors.NewValidationErrorf("health check interval must be a positive duration")
	}
	return nil
}

// Init converts integer seconds to time.Duration fields.
func (c *TimeoutsConfig) Init() error {
	c.ShutdownTimeout = time.Duration(c.ShutdownDelayInSeconds) * time.Second
	c.RequestTimeout = time.Duration(c.RequestTimeoutSeconds) * time.Second
	c.HealthCheckInterval = time.Duration(c.HealthCheckIntervalSeconds) * time.Second
	return nil
}

var _ Initializer = (*TimeoutsConfig)(nil)
var _ Validator = (*TimeoutsConfig)(nil)
