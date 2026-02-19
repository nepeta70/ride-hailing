package config_test

import (
	"testing"
	"time"

	. "github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestDefaultServerConfig(t *testing.T) {
	cfg := DefaultServerConfig()

	assert.Equal(t, 5001, cfg.Port)
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, 5, cfg.ReadTimeoutSeconds)
	assert.Equal(t, 5*time.Second, cfg.ReadTimeout)
}

func TestServerConfig_SetDurations(t *testing.T) {
	t.Run("converts seconds to time.Duration correctly", func(t *testing.T) {
		cfg := ServerConfig{
			ReadTimeoutSeconds:  10,
			WriteTimeoutSeconds: 20,
			IdleTimeoutSeconds:  120,
		}

		cfg.Init()

		assert.Equal(t, 10*time.Second, cfg.ReadTimeout)
		assert.Equal(t, 20*time.Second, cfg.WriteTimeout)
		assert.Equal(t, 120*time.Second, cfg.IdleTimeout)
	})
}

func TestServerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ServerConfig
		wantErr bool
	}{
		{
			name:    "port too low",
			cfg:     ServerConfig{Port: 0, Host: "127.0.0.1", ReadTimeoutSeconds: 5, WriteTimeoutSeconds: 5, IdleTimeoutSeconds: 5},
			wantErr: true,
		},
		{
			name:    "empty host",
			cfg:     ServerConfig{Port: 5001, Host: "", ReadTimeoutSeconds: 5, WriteTimeoutSeconds: 5, IdleTimeoutSeconds: 5},
			wantErr: true,
		},
		{
			name:    "zero read timeout",
			cfg:     ServerConfig{Port: 5001, Host: "127.0.0.1", ReadTimeoutSeconds: 0, WriteTimeoutSeconds: 5, IdleTimeoutSeconds: 5},
			wantErr: true,
		},
		{
			name:    "negative write timeout",
			cfg:     ServerConfig{Port: 5001, Host: "127.0.0.1", ReadTimeoutSeconds: 5, WriteTimeoutSeconds: -1, IdleTimeoutSeconds: 5},
			wantErr: true,
		},
		{
			name:    "negative idle timeout",
			cfg:     ServerConfig{Port: 5001, Host: "127.0.0.1", ReadTimeoutSeconds: 5, WriteTimeoutSeconds: 5, IdleTimeoutSeconds: -1},
			wantErr: true,
		},
		{
			name: "valid config with timeouts",
			cfg: ServerConfig{
				Port:                5001,
				Host:                "localhost",
				ReadTimeoutSeconds:  5,
				WriteTimeoutSeconds: 10,
				IdleTimeoutSeconds:  120,
			},
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
