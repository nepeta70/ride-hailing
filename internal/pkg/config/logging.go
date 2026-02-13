package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type LoggingConfig struct {
	Format string `json:"log_format" env:"LOG_FORMAT"`
	Level  string `json:"log_level" env:"LOG_LEVEL"`
}

// DefaultLoggingConfig returns the default LoggingConfig.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:  "INFO",
		Format: "json",
	}
}

// ConfigureLogger initializes the global slog instance based on config
func (c *LoggingConfig) ConfigureLogger() ports.Logger {
	opts := &slog.HandlerOptions{
		Level: c.slogLevel(),
	}

	var handler slog.Handler
	if c.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

func (c *LoggingConfig) slogLevel() slog.Level {
	switch strings.ToUpper(c.Level) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
