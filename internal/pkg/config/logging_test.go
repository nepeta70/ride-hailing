package config_test

import (
	"testing"

	. "github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestDefaultLoggingConfig(t *testing.T) {
	cfg := DefaultLoggingConfig()

	assert.Equal(t, "INFO", cfg.Level)
}

func TestConfigureLogger_DoesNotPanic(t *testing.T) {
	levels := []string{"DEBUG", "WARN", "ERROR", "INFO", "unknown"}
	formats := []string{"json", "text", "unknown"}

	for _, level := range levels {
		for _, format := range formats {
			level := level
			format := format

			t.Run(level+"/"+format, func(t *testing.T) {
				cfg := LoggingConfig{
					Level: level,
				}

				assert.NotPanics(t, func() {
					logger := cfg.GetConsoleLogger()
					assert.NotNil(t, logger)
				}, "GetConsoleLogger panicked for level=%q", level)
			})
		}
	}
}
