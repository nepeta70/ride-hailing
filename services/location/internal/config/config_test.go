package config_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/nepeta70/ride-hailing/services/location/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, "location", cfg.ServiceName)
	assert.Equal(t, 300, cfg.Logic.LocationTTLSeconds)
	assert.Equal(t, float32(0.5), cfg.Logic.MinRadiusSearchKm)

	// Ensure defaults pass validation
	assert.NoError(t, cfg.Logic.Validate())
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		logic   LogicConfig
		wantErr bool
	}{
		{
			name: "valid logic",
			logic: LogicConfig{
				LocationTTLSeconds: 60,
				TopKNearby:         10,
				MinRadiusSearchKm:  1.0,
				MaxRadiusSearchKm:  10.0,
			},
			wantErr: false,
		},
		{
			name: "invalid radius order",
			logic: LogicConfig{
				LocationTTLSeconds: 60,
				TopKNearby:         10,
				MinRadiusSearchKm:  10.0, // Min > Max
				MaxRadiusSearchKm:  5.0,
			},
			wantErr: true,
		},
		{
			name: "negative ttl",
			logic: LogicConfig{
				LocationTTLSeconds: -1,
				TopKNearby:         5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Logic = &tt.logic
			if tt.wantErr {
				assert.Error(t, cfg.Logic.Validate())
			} else {
				assert.NoError(t, cfg.Logic.Validate())
			}
		})
	}
}

func TestLoad_JSON(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]any
		expect func(*testing.T, *Config)
	}{
		{
			name: "valid json overrides logic and redis",
			input: map[string]any{
				"logic": map[string]any{
					"location_ttl_seconds": 120,
					"max_radius_search_km": 15.5,
				},
			},
			expect: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 120, cfg.Logic.LocationTTLSeconds)
				assert.Equal(t, float32(15.5), cfg.Logic.MaxRadiusSearchKm)
				// Ensure non-overridden defaults stay same
				assert.Equal(t, 10, cfg.Logic.TopKNearby)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "location-cfg-*.json")
			require.NoError(t, err)
			defer os.Remove(file.Name())

			require.NoError(t, json.NewEncoder(file).Encode(tt.input))
			require.NoError(t, file.Close())

			cfg, err := Load(file.Name())
			assert.NoError(t, err)
			tt.expect(t, cfg)
		})
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	// If file is missing, LoadGeneric should fall back to DefaultConfig()
	cfg, err := Load("/nonexistent/location-cfg.json")

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "location", cfg.ServiceName)
	assert.Equal(t, 300, cfg.Logic.LocationTTLSeconds)
}
