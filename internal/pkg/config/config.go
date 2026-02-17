package config

import (
	"encoding/json"
	"os"
	"reflect"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"

	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/telemetry"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

// BaseConfig contains the stuff every single service needs
type BaseConfig struct {
	Server      ServerConfig              `json:"server"`
	Logging     LoggingConfig             `json:"logging"`
	Timeouts    TimeoutsConfig            `json:"timeouts"`
	Security    SecurityConfig            `json:"security"`
	Telemetry   telemetry.TelemetryConfig `json:"telemetry"`
	APIKey      string                    `json:"api_key" env:"API_KEY"`
	ServiceName string                    `json:"service_name" env:"SERVICE_NAME"`
}

// DefaultBaseConfig returns the boilerplate defaults
func DefaultBaseConfig() BaseConfig {
	return BaseConfig{
		Server:      DefaultServerConfig(),
		Timeouts:    DefaultTimeoutsConfig(),
		Logging:     DefaultLoggingConfig(),
		Security:    DefaultSecurityConfig(),
		Telemetry:   telemetry.DefaultTelemetryConfig(),
		APIKey:      "",
		ServiceName: "Service",
	}
}

func LoadGeneric[T any](path string, cfg *T) (*T, error) {
	// 1. Load Data
	_ = godotenv.Load()
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, errors.NewErrJSONUnmarshal(err)
		}
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	// 2. Run Lifecycle
	err = processConfig(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func processConfig(cfg any) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i).Addr().Interface()
		if validator, ok := field.(Validator); ok {
			if err := validator.Validate(); err != nil {
				return err
			}
		}
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i).Addr().Interface()
		if initializer, ok := field.(Initializer); ok {
			if err := initializer.Init(); err != nil {
				return err
			}
		}
	}

	return nil
}
