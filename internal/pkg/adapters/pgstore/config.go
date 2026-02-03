package pgstore

import (
	"fmt"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type PostgresConfig struct {
	User        string        `json:"user" env:"POSTGRES_USER"`
	Password    string        `json:"password" env:"POSTGRES_PASSWORD"`
	DBName      string        `json:"dbname" env:"POSTGRES_DB"`
	Host        string        `json:"host" env:"POSTGRES_HOST"`
	Port        int           `json:"port" env:"POSTGRES_PORT"`
	SSLMode     string        `json:"ssl_mode" env:"POSTGRES_SSLMODE"`
	PingSeconds int           `json:"ping_timeout_seconds" env:"POSTGRES_PING_TIMEOUT_SECONDS"`
	PingTimeout time.Duration `json:"-"`
}

func (c PostgresConfig) DSN() string {
	// Template: postgres://user:password@host:port/dbname?sslmode=disable
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

func (c *PostgresConfig) Init() error {
	c.PingTimeout = time.Duration(c.PingSeconds) * time.Second
	return nil
}

func (c *PostgresConfig) Validate() error {
	if c.User == "" {
		return errors.NewValidationErrorf("postgres user is required")
	}
	if c.DBName == "" {
		return errors.NewValidationErrorf("postgres dbname is required")
	}
	if c.Host == "" {
		return errors.NewValidationErrorf("postgres host is required")
	}
	if c.Port == 0 {
		return errors.NewValidationErrorf("postgres port is required")
	}
	if c.PingSeconds <= 0 {
		return errors.NewValidationErrorf("postgres ping timeout must be greater than zero")
	}
	return nil
}
