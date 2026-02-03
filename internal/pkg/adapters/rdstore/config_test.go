package rdstore_test

import (
	"testing"
	"time"

	. "github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	"github.com/stretchr/testify/assert"
)

func TestRedisConfig_Validate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     RedisConfig
		wantErr bool
	}{
		{
			name: "valid config with address",
			cfg: RedisConfig{
				Address:      "127.0.0.1:6379",
				DialSeconds:  5,
				ReadSeconds:  1,
				WriteSeconds: 1,
			},
			wantErr: false,
		},
		{
			name: "valid config with host/port",
			cfg: RedisConfig{
				Host:         "127.0.0.1",
				Port:         6379,
				DialSeconds:  5,
				ReadSeconds:  1,
				WriteSeconds: 1,
			},
			wantErr: false,
		},
		{
			name: "missing address and host/port",
			cfg: RedisConfig{
				DialSeconds:  5,
				ReadSeconds:  1,
				WriteSeconds: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid dial timeout",
			cfg: RedisConfig{
				Address:      "127.0.0.1:6379",
				DialSeconds:  0,
				ReadSeconds:  1,
				WriteSeconds: 1,
			},
			wantErr: true,
		},
		{
			name: "negative read timeout",
			cfg: RedisConfig{
				Address:      "127.0.0.1:6379",
				DialSeconds:  5,
				ReadSeconds:  -1,
				WriteSeconds: 1,
			},
			wantErr: true,
		},
		{
			name: "negative write timeout",
			cfg: RedisConfig{
				Address:      "127.0.0.1:6379",
				DialSeconds:  5,
				ReadSeconds:  1,
				WriteSeconds: -1,
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

func TestRedisConfig_Init(t *testing.T) {
	cases := []struct {
		name        string
		cfg         RedisConfig
		wantAddress string
		wantPool    time.Duration
		wantRead    time.Duration
		wantWrite   time.Duration
		wantDial    time.Duration
	}{
		{
			name:        "init with host/port",
			cfg:         RedisConfig{Host: "127.0.0.1", Port: 6379, PoolSeconds: 10, ReadSeconds: 2, WriteSeconds: 3, DialSeconds: 4},
			wantAddress: "127.0.0.1:6379",
			wantPool:    10 * time.Second,
			wantRead:    2 * time.Second,
			wantWrite:   3 * time.Second,
			wantDial:    4 * time.Second,
		},
		{
			name:        "init with address",
			cfg:         RedisConfig{Address: "localhost:6380", PoolSeconds: 5, ReadSeconds: 1, WriteSeconds: 2, DialSeconds: 3},
			wantAddress: "localhost:6380",
			wantPool:    5 * time.Second,
			wantRead:    1 * time.Second,
			wantWrite:   2 * time.Second,
			wantDial:    3 * time.Second,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.cfg.Init()
			assert.Equal(t, tc.wantAddress, tc.cfg.Address)
			assert.Equal(t, tc.wantPool, tc.cfg.PoolTimeout)
			assert.Equal(t, tc.wantRead, tc.cfg.ReadTimeout)
			assert.Equal(t, tc.wantWrite, tc.cfg.WriteTimeout)
			assert.Equal(t, tc.wantDial, tc.cfg.DialTimeout)
		})
	}
}
