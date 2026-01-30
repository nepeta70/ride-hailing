package postgres

import (
	"fmt"
)

type PostgresConfig struct {
	User     string `json:"user" env:"POSTGRES_USER"`
	Password string `json:"password" env:"POSTGRES_PASSWORD"`
	DBName   string `json:"dbname" env:"POSTGRES_DB"`
	Host     string `json:"host" env:"POSTGRES_HOST"`
	Port     int    `json:"port" env:"POSTGRES_PORT"`
	SSLMode  string `json:"ssl_mode" env:"POSTGRES_SSLMODE"`
}

func (c PostgresConfig) DSN() string {
	// Template: postgres://user:password@host:port/dbname?sslmode=disable
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}
