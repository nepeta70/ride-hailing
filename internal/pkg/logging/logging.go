package logging

import (
	"log/slog"
	"os"
	"strings"
)

type LoggingConfig struct {
	Format string `json:"log_format"`
	Level  string `json:"log_level"`
}

// ConfigureLogger initializes the global slog instance based on config
func (c *LoggingConfig) ConfigureLogger() *slog.Logger {
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

// OverrideLogLevelFromEnv sets the LoggingConfig.Level from the LOG_LEVEL environment variable if present.
func (c *LoggingConfig) OverrideLogLevelFromEnv() {
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		c.Level = logLevel
	}
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

// DefaultLoggingConfig returns the default LoggingConfig.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:  "INFO",
		Format: "json",
	}
}
