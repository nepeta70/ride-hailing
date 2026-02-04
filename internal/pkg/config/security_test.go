package config_test

import (
	"testing"

	. "github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestDefaultSecurityConfig(t *testing.T) {
	cfg := DefaultSecurityConfig()

	assert.Equal(t, 10.0, cfg.RateLimit)
	assert.Equal(t, 20, cfg.RateBurst)
	assert.Equal(t, int64(1), cfg.MaxBodyMB)
}

func TestMaxBodyBytes(t *testing.T) {
	cfg := SecurityConfig{MaxBodyMB: 2}
	cfg.Init()
	expectedBytes := int64(2 * 1024 * 1024)

	assert.Equal(t, expectedBytes, cfg.MaxBodyBytes)
}

func TestSecurityConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     SecurityConfig
		wantErr bool
	}{
		{
			name:    "invalid rate limit",
			cfg:     SecurityConfig{RateLimit: 0, MaxBodyMB: 1},
			wantErr: true,
		},
		{
			name:    "invalid max body",
			cfg:     SecurityConfig{RateLimit: 1, MaxBodyMB: 0},
			wantErr: true,
		},
		{
			name:    "valid config",
			cfg:     SecurityConfig{RateLimit: 1, MaxBodyMB: 1},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.IsType(t, &errors.ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
