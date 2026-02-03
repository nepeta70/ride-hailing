package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type SecurityConfig struct {
	RateLimit float64 `json:"rate_limit" env:"SECURITY_RATE_LIMIT"`
	RateBurst int     `json:"rate_burst" env:"SECURITY_RATE_BURST"`
	MaxBodyMB int64   `json:"max_body_mb" env:"SECURITY_MAX_BODY_MB"`
}

func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		RateLimit: 10,
		RateBurst: 20,
		MaxBodyMB: 1,
	}
}

func (s *SecurityConfig) MaxBodyBytes() int64 {
	return s.MaxBodyMB * 1024 * 1024
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

var _ Validator = (*SecurityConfig)(nil)
