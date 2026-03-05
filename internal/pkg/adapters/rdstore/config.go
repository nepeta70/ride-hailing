package rdstore

import (
	"net"
	"strconv"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type RedisConfig struct {
	Host     string `json:"host" env:"REDIS_HOST"`
	Port     int    `json:"port" env:"REDIS_PORT"`
	Password string `json:"password" env:"REDIS_PASSWORD"`
	DB       int    `json:"db" env:"REDIS_DB"`
	Address  string `json:"address" env:"REDIS_ADDRESS"`
	PoolSize int    `json:"pool_size" env:"REDIS_POOL_SIZE"`
	MinIdle  int    `json:"min_idle" env:"REDIS_MIN_IDLE"`

	PoolSeconds  int `json:"pool_timeout_seconds" env:"REDIS_POOL_TIMEOUT_SECONDS"`
	ReadSeconds  int `json:"read_timeout_seconds" env:"REDIS_READ_TIMEOUT_SECONDS"`
	WriteSeconds int `json:"write_timeout_seconds" env:"REDIS_WRITE_TIMEOUT_SECONDS"`
	DialSeconds  int `json:"dial_timeout_seconds" env:"REDIS_DIAL_TIMEOUT_SECONDS"`

	PoolTimeout  time.Duration `json:"-"`
	ReadTimeout  time.Duration `json:"-"`
	WriteTimeout time.Duration `json:"-"`
	DialTimeout  time.Duration `json:"-"`
}

func DefaultRedisConfig() RedisConfig {
	cfg := RedisConfig{
		Host:         "127.0.0.1",
		Port:         6379,
		Password:     "",
		DB:           0,
		Address:      "127.0.0.1:6379",
		PoolSize:     10,
		MinIdle:      2,
		PoolSeconds:  30,
		ReadSeconds:  3,
		WriteSeconds: 3,
		DialSeconds:  5,
	}
	cfg.Init()
	return cfg
}

func (c *RedisConfig) Validate() error {
	if c.Address == "" {
		if c.Host == "" || c.Port == 0 {
			return errors.NewValidationErrorf("redis address or host/port are required")
		}
	}

	if c.DialSeconds <= 0 {
		return errors.NewValidationErrorf("redis dial timeout must be greater than zero")
	}

	if c.ReadSeconds < 0 || c.WriteSeconds < 0 || c.PoolSeconds < 0 {
		return errors.NewValidationErrorf("redis read/write/pool timeouts cannot be negative")
	}
	return nil
}

func (c *RedisConfig) Init() error {
	if c.Address == "" && c.Host != "" {
		c.Address = net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
	}
	c.PoolTimeout = time.Duration(c.PoolSeconds) * time.Second
	c.ReadTimeout = time.Duration(c.ReadSeconds) * time.Second
	c.WriteTimeout = time.Duration(c.WriteSeconds) * time.Second
	c.DialTimeout = time.Duration(c.DialSeconds) * time.Second
	return nil
}

var _ config.Initializer = (*RedisConfig)(nil)
var _ config.Validator = (*RedisConfig)(nil)
