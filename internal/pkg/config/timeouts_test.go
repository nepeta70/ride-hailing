package config_test

import (
	"testing"
	"time"

	. "github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestDefaultTimeoutsConfig(t *testing.T) {
	cfg := DefaultTimeoutsConfig()

	assert.Equal(t, 10, cfg.ShutdownDelayInSeconds)
	assert.Equal(t, 5, cfg.RequestTimeoutSeconds)
	// Also verify the duration fields match the defaults
	assert.Equal(t, 10*time.Second, cfg.ShutdownTimeout)
	assert.Equal(t, 5*time.Second, cfg.RequestTimeout)
}

func TestSetDurations(t *testing.T) {
	cfg := TimeoutsConfig{
		ShutdownDelayInSeconds: 2,
		RequestTimeoutSeconds:  3,
	}

	cfg.Init()

	assert.Equal(t, 2*time.Second, cfg.ShutdownTimeout)
	assert.Equal(t, 3*time.Second, cfg.RequestTimeout)
}

func TestTimeoutsConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     TimeoutsConfig
		wantErr bool
	}{
		{
			name: "invalid request timeout",
			cfg: TimeoutsConfig{
				RequestTimeoutSeconds:      0,
				ShutdownDelayInSeconds:     1,
				HealthCheckIntervalSeconds: 5,
			},
			wantErr: true,
		},
		{
			name: "negative shutdown delay",
			cfg: TimeoutsConfig{
				RequestTimeoutSeconds:      1,
				ShutdownDelayInSeconds:     -1,
				HealthCheckIntervalSeconds: 5,
			},
			wantErr: true,
		},
		{
			name: "zero health check interval",
			cfg: TimeoutsConfig{
				RequestTimeoutSeconds:      1,
				ShutdownDelayInSeconds:     1,
				HealthCheckIntervalSeconds: 0,
			},
			wantErr: true,
		},
		{
			name: "valid timeouts",
			cfg: TimeoutsConfig{
				RequestTimeoutSeconds:      1,
				ShutdownDelayInSeconds:     1,
				HealthCheckIntervalSeconds: 5,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cfg.Init()
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
