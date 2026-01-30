package config

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type TimeoutsConfig struct {
	ShutdownDelayInSeconds int           `json:"shutdown_timeout_seconds" env:"SHUTDOWN_TIMEOUT_SECONDS"`
	RequestTimeoutSeconds  int           `json:"request_timeout_seconds" env:"REQUEST_TIMEOUT_SECONDS"`
	ShutdownTimeout        time.Duration `json:"-"`
	RequestTimeout         time.Duration `json:"-"`
}

// DefaultTimeoutsConfig returns the default TimeoutsConfig.
func DefaultTimeoutsConfig() TimeoutsConfig {
	return TimeoutsConfig{
		ShutdownDelayInSeconds: 10,
		RequestTimeoutSeconds:  5,
		ShutdownTimeout:        10 * time.Second,
		RequestTimeout:         5 * time.Second,
	}
}

func (c *TimeoutsConfig) Validate() error {
	if c.RequestTimeoutSeconds <= 0 {
		return errors.NewValidationErrorf("request timeout must be greater than zero")
	}
	if c.ShutdownDelayInSeconds < 0 {
		return errors.NewValidationErrorf("shutdown delay cannot be negative")
	}
	return nil
}

// Init converts integer seconds to time.Duration fields.
func (c *TimeoutsConfig) Init() error {
	c.ShutdownTimeout = time.Duration(c.ShutdownDelayInSeconds) * time.Second
	c.RequestTimeout = time.Duration(c.RequestTimeoutSeconds) * time.Second
	return nil
}

var _ Initializer = (*TimeoutsConfig)(nil)
var _ Validator = (*TimeoutsConfig)(nil)
