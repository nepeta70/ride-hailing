package config_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/nepeta70/ride-hailing/services/location-service/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, 7, cfg.Logic.GeohashPrecision)
	assert.Equal(t, 300, cfg.Logic.LocationTTLSeconds)
	assert.Equal(t, 5, cfg.Logic.TopKNearby)
}

func TestLoad_JSON(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]any
		expect LogicConfig
	}{
		{
			name: "valid json overrides defaults",
			input: map[string]any{
				"logic": map[string]any{
					"geohash_precision":    9,
					"location_ttl_seconds": 100,
					"top_k_nearby":         10,
				},
			},
			expect: LogicConfig{
				GeohashPrecision:   9,
				LocationTTLSeconds: 100,
				TopKNearby:         10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "location-cfg-*.json")
			assert.NoError(t, err)
			defer os.Remove(file.Name())

			assert.NoError(t, json.NewEncoder(file).Encode(tt.input))
			assert.NoError(t, file.Close())

			cfg, err := Load(file.Name())
			assert.NoError(t, err)

			assert.Equal(t, tt.expect.GeohashPrecision, cfg.Logic.GeohashPrecision)
			assert.Equal(t, tt.expect.LocationTTLSeconds, cfg.Logic.LocationTTLSeconds)
			assert.Equal(t, tt.expect.TopKNearby, cfg.Logic.TopKNearby)
		})
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	cfg, err := Load("/nonexistent/location-cfg.json")

	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	def := DefaultConfig()
	assert.Equal(t, *def, *cfg)
}
