package config

import (
	"log/slog"
	"os"
	"strings"
)

type LoggingConfig struct {
	Level string `json:"log_level" env:"LOG_LEVEL"`
}

// DefaultLoggingConfig returns the default LoggingConfig.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level: "INFO",
	}
}

func (c *LoggingConfig) GetConsoleLogger() slog.Handler {
	opts := &slog.HandlerOptions{
		Level: c.slogLevel(),
	}

	return slog.NewJSONHandler(os.Stdout, opts)
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
