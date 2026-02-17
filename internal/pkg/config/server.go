package config

import (
	"strings"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type ServerConfig struct {
	Port int    `json:"port" env:"SERVER_PORT"`
	Host string `json:"host" env:"SERVER_HOST"`
	// Integer fields for JSON/Env mapping
	ReadTimeoutSeconds         int `json:"read_timeout_seconds" env:"SERVER_READ_TIMEOUT_SECONDS"`
	WriteTimeoutSeconds        int `json:"write_timeout_seconds" env:"SERVER_WRITE_TIMEOUT_SECONDS"`
	IdleTimeoutSeconds         int `json:"idle_timeout_seconds" env:"SERVER_IDLE_TIMEOUT_SECONDS"`
	HealthCheckIntervalSeconds int `json:"health_check_interval_seconds" env:"SERVER_HEALTH_CHECK_INTERVAL_SECONDS"`

	// Duration fields (ignored by JSON) used by the http.Server
	ReadTimeout         time.Duration `json:"-"`
	WriteTimeout        time.Duration `json:"-"`
	IdleTimeout         time.Duration `json:"-"`
	HealthCheckInterval time.Duration `json:"-"`
}

func DefaultServerConfig() ServerConfig {
	cfg := ServerConfig{
		Port:                       5001,
		Host:                       "127.0.0.1",
		ReadTimeoutSeconds:         5,
		WriteTimeoutSeconds:        10,
		IdleTimeoutSeconds:         120,
		HealthCheckIntervalSeconds: 30,
	}
	cfg.Init()
	return cfg
}

func (c *ServerConfig) Init() error {
	c.ReadTimeout = time.Duration(c.ReadTimeoutSeconds) * time.Second
	c.WriteTimeout = time.Duration(c.WriteTimeoutSeconds) * time.Second
	c.IdleTimeout = time.Duration(c.IdleTimeoutSeconds) * time.Second
	c.HealthCheckInterval = time.Duration(c.HealthCheckIntervalSeconds) * time.Second
	return nil
}

func (c *ServerConfig) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return errors.NewValidationErrorf("invalid server port: %d", c.Port)
	}
	if strings.TrimSpace(c.Host) == "" {
		return errors.NewValidationErrorf("server host cannot be empty")
	}
	if c.ReadTimeoutSeconds <= 0 || c.WriteTimeoutSeconds <= 0 || c.IdleTimeoutSeconds <= 0 {
		return errors.NewValidationErrorf("server timeouts must be positive durations")
	}
	if c.HealthCheckIntervalSeconds <= 0 {
		return errors.NewValidationErrorf("health check interval must be a positive duration")
	}
	return nil
}

var _ Initializer = (*ServerConfig)(nil)
var _ Validator = (*ServerConfig)(nil)
