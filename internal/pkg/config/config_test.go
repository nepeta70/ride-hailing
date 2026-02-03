package config_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/nepeta70/ride-hailing/internal/pkg/config"
)

type ConfigMock struct {
	Valid bool
}

func (d *ConfigMock) Validate() error {
	if !d.Valid {
		return os.ErrInvalid
	}
	return nil
}

func (d *ConfigMock) Init() error {
	return nil
}

func TestDefaultBaseConfig(t *testing.T) {
	cfg := DefaultBaseConfig()
	assert.NotEqual(t, BaseConfig{}, cfg, "DefaultBaseConfig should not return zero value")
}

func TestLoadGeneric_FileNotFound(t *testing.T) {
	type TestConfig struct {
		A string `json:"a"`
	}

	var cfg TestConfig
	_, err := LoadGeneric("nonexistent.json", &cfg)

	assert.True(t,
		err == nil || os.IsNotExist(err),
		"expected file not found error or nil, got %v", err,
	)
}

func TestLoadGeneric_Validator(t *testing.T) {
	type TestConfig struct {
		Dummy ConfigMock
	}

	tests := []struct {
		name      string
		valid     bool
		expectErr bool
	}{
		{
			name:      "invalid config",
			valid:     false,
			expectErr: true,
		},
		{
			name:      "valid config",
			valid:     true,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp, err := os.CreateTemp("", "testcfg-*.json")
			assert.NoError(t, err)
			defer os.Remove(tmp.Name())

			err = json.NewEncoder(tmp).Encode(map[string]any{
				"Dummy": map[string]any{"Valid": tt.valid},
			})
			assert.NoError(t, err)
			assert.NoError(t, tmp.Close())

			var cfg TestConfig
			_, err = LoadGeneric(tmp.Name(), &cfg)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadGeneric_JSONUnmarshal(t *testing.T) {
	type TestConfig struct {
		A string `json:"a"`
	}

	tests := []struct {
		name     string
		input    map[string]string
		expected string
	}{
		{
			name:     "valid json",
			input:    map[string]string{"a": "hello"},
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "testcfg-*.json")
			assert.NoError(t, err)
			defer os.Remove(file.Name())

			err = json.NewEncoder(file).Encode(tt.input)
			assert.NoError(t, err)
			assert.NoError(t, file.Close())

			var cfg TestConfig
			_, err = LoadGeneric(file.Name(), &cfg)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.A)
		})
	}
}
