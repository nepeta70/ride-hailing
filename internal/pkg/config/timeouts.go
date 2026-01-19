package config

import "time"

type TimeoutsConfig struct {
	ResetSystemInSeconds int           `json:"reset_system_in_seconds" env:"RESET_SYSTEM_IN_SECONDS" envDefault:"10"`
	StandardInSeconds    int           `json:"standard_in_seconds" env:"STANDARD_IN_SECONDS" envDefault:"5"`
	SetFleetTimeout      time.Duration `json:"-"`
	StandardTimeout      time.Duration `json:"-"`
}

// SetDurations converts integer seconds to time.Duration fields.
func (c *TimeoutsConfig) SetDurations() {
	c.SetFleetTimeout = time.Duration(c.ResetSystemInSeconds) * time.Second
	c.StandardTimeout = time.Duration(c.StandardInSeconds) * time.Second
}

// DefaultTimeoutsConfig returns the default TimeoutsConfig.
func DefaultTimeoutsConfig() TimeoutsConfig {
	return TimeoutsConfig{
		ResetSystemInSeconds: 10,
		StandardInSeconds:    5,
		SetFleetTimeout:      10 * time.Second,
		StandardTimeout:      5 * time.Second,
	}
}
