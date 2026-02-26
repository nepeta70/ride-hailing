package config

type LoggingConfig struct {
	Level string `json:"log_level" env:"LOG_LEVEL"`
}

// DefaultLoggingConfig returns the default LoggingConfig.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level: "INFO",
	}
}
