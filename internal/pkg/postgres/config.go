package postgres

import (
	"fmt"
	"os"
	"strconv"
)

type PostgresConfig struct {
	User     string `json:"user" env:"POSTGRES_USER"`
	Password string `json:"password" env:"POSTGRES_PASSWORD"`
	DBName   string `json:"dbname" env:"POSTGRES_DB"`
	Host     string `json:"host" env:"POSTGRES_HOST"`
	Port     int    `json:"port" env:"POSTGRES_PORT"`
	SSLMode  string `json:"ssl_mode" env:"POSTGRES_SSLMODE"`
}

func (c *PostgresConfig) OverrideFromEnv() {
	if val := os.Getenv("POSTGRES_USER"); val != "" {
		c.User = val
	}
	if val := os.Getenv("POSTGRES_PASSWORD"); val != "" {
		c.Password = val
	}
	if val := os.Getenv("POSTGRES_DB"); val != "" {
		c.DBName = val
	}
	if val := os.Getenv("POSTGRES_HOST"); val != "" {
		c.Host = val
	}
	if val := os.Getenv("POSTGRES_PORT"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			c.Port = p
		}
	}
	if val := os.Getenv("POSTGRES_SSLMODE"); val != "" {
		c.SSLMode = val
	}
}

func (c PostgresConfig) DSN() string {
	// Template: postgres://user:password@host:port/dbname?sslmode=disable
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}
