package config_test

import (
	"testing"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestDefaultLoggingConfig(t *testing.T) {
	cfg := config.DefaultLoggingConfig()

	assert.Equal(t, "INFO", cfg.Level)
	assert.Equal(t, "json", cfg.Format)
}

func TestConfigureLogger_DoesNotPanic(t *testing.T) {
	levels := []string{"DEBUG", "WARN", "ERROR", "INFO", "unknown"}
	formats := []string{"json", "text", "unknown"}

	for _, level := range levels {
		for _, format := range formats {
			level := level
			format := format

			t.Run(level+"/"+format, func(t *testing.T) {
				cfg := config.LoggingConfig{
					Level:  level,
					Format: format,
				}

				assert.NotPanics(t, func() {
					logger := cfg.ConfigureLogger()
					assert.NotNil(t, logger)
				}, "ConfigureLogger panicked for level=%q, format=%q", level, format)
			})
		}
	}
}
