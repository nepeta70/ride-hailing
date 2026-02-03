package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type SecurityConfig struct {
	RateLimit    float64 `json:"rate_limit" env:"SECURITY_RATE_LIMIT"`
	RateBurst    int     `json:"rate_burst" env:"SECURITY_RATE_BURST"`
	MaxBodyMB    int64   `json:"max_body_mb" env:"SECURITY_MAX_BODY_MB"`
	MaxBodyBytes int64   `json:"-"` // calculated field
}

func DefaultSecurityConfig() SecurityConfig {
	cfg := SecurityConfig{
		RateLimit: 10,
		RateBurst: 20,
		MaxBodyMB: 1,
	}
	cfg.Init()
	return cfg
}

func (s *SecurityConfig) Validate() error {
	if s.RateLimit <= 0 {
		return errors.NewValidationErrorf("security rate limit must be a positive value")
	}
	if s.MaxBodyMB <= 0 {
		return errors.NewValidationErrorf("max body size must be at least 1MB")
	}
	return nil
}

func (s *SecurityConfig) Init() error {
	s.MaxBodyBytes = s.MaxBodyMB * 1024 * 1024
	return nil
}

var _ Initializer = (*SecurityConfig)(nil)

var _ Validator = (*SecurityConfig)(nil)
