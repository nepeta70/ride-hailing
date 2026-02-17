package telemetry

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type TelemetryConfig struct {
	PrometheusAddress    string        `json:"prometheus_address" env:"PROMETHEUS_ADDRESS"`
	OpentelemetryAddress string        `json:"opentelemetry_address" env:"OPENTELEMETRY_ADDRESS"`
	IntervalSeconds      int           `json:"telemetry_interval_seconds" env:"TELEMETRY_INTERVAL_SECONDS"`
	Interval             time.Duration `json:"-"`
}

func DefaultTelemetryConfig() TelemetryConfig {
	return TelemetryConfig{
		PrometheusAddress:    "http://localhost:9090",
		OpentelemetryAddress: "http://localhost:4317",
		IntervalSeconds:      15,
	}
}

func (tc *TelemetryConfig) Validate() error {
	if tc.PrometheusAddress == "" {
		return errors.NewValidationErrorf("Prometheus address is required")
	}
	if tc.OpentelemetryAddress == "" {
		return errors.NewValidationErrorf("OpenTelemetry address is required")
	}
	if tc.IntervalSeconds <= 0 {
		return errors.NewValidationErrorf("Telemetry interval must be greater than 0")
	}

	return nil
}

func (tc *TelemetryConfig) Init() {
	tc.Interval = time.Duration(tc.IntervalSeconds) * time.Second
}
