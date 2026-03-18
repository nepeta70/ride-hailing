package pgstore_test

import (
	"testing"
	"time"

	. "github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
	"github.com/stretchr/testify/assert"
)

func TestPostgresConfig_Validate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     PostgresConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: PostgresConfig{
				User:         "user",
				Password:     "pass",
				DBName:       "db",
				Host:         "localhost",
				Port:         5432,
				SSLMode:      "disable",
				PingSeconds:  5,
				QuerySeconds: 1,
			},
			wantErr: false,
		},
		{
			name: "missing user",
			cfg: PostgresConfig{
				DBName:       "db",
				Host:         "localhost",
				Port:         5432,
				SSLMode:      "disable",
				PingSeconds:  5,
				QuerySeconds: 1,
			},
			wantErr: true,
		},
		{
			name: "missing dbname",
			cfg: PostgresConfig{
				User:         "user",
				Host:         "localhost",
				Port:         5432,
				SSLMode:      "disable",
				PingSeconds:  5,
				QuerySeconds: 1,
			},
			wantErr: true,
		},
		{
			name: "missing host",
			cfg: PostgresConfig{
				User:         "user",
				DBName:       "db",
				Port:         5432,
				SSLMode:      "disable",
				PingSeconds:  5,
				QuerySeconds: 1,
			},
			wantErr: true,
		},
		{
			name: "missing port",
			cfg: PostgresConfig{
				User:         "user",
				DBName:       "db",
				Host:         "localhost",
				SSLMode:      "disable",
				PingSeconds:  5,
				QuerySeconds: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid ping timeout",
			cfg: PostgresConfig{
				User:         "user",
				DBName:       "db",
				Host:         "localhost",
				Port:         5432,
				SSLMode:      "disable",
				PingSeconds:  0,
				QuerySeconds: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid query timeout",
			cfg: PostgresConfig{
				User:         "user",
				DBName:       "db",
				Host:         "localhost",
				Port:         5432,
				SSLMode:      "disable",
				PingSeconds:  5,
				QuerySeconds: 0,
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostgresConfig_Init(t *testing.T) {
	cases := []struct {
		name        string
		cfg         PostgresConfig
		wantTimeout time.Duration
	}{
		{
			name:        "init with 10 seconds",
			cfg:         PostgresConfig{PingSeconds: 10, QuerySeconds: 10},
			wantTimeout: 10 * time.Second,
		},
		{
			name:        "init with 0 seconds",
			cfg:         PostgresConfig{PingSeconds: 0, QuerySeconds: 0},
			wantTimeout: 0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.cfg.Init()
			assert.Equal(t, tc.wantTimeout, tc.cfg.PingTimeout)
			assert.Equal(t, tc.wantTimeout, tc.cfg.QueryTimeout)
		})
	}
}

func TestPostgresConfig_DSN(t *testing.T) {
	cases := []struct {
		name     string
		cfg      PostgresConfig
		expected string
	}{
		{
			name: "standard dsn",
			cfg: PostgresConfig{
				User:     "user",
				Password: "pass",
				DBName:   "db",
				Host:     "localhost",
				Port:     5432,
				SSLMode:  "disable",
			},
			expected: "postgres://user:pass@localhost:5432/db?sslmode=disable",
		},
		{
			name:     "empty fields",
			cfg:      PostgresConfig{},
			expected: "postgres://:@:0?sslmode=",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cfg.DSN()
			assert.Equal(t, tc.expected, got)
		})
	}
}
