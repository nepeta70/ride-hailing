package redis

import (
	"time"
)

type RedisConfig struct {
	Host         string        `json:"host" env:"REDIS_HOST"`
	Port         int           `json:"port" env:"REDIS_PORT"`
	Password     string        `json:"password" env:"REDIS_PASSWORD"`
	DB           int           `json:"db" env:"REDIS_DB"`
	Address      string        `json:"address" env:"REDIS_ADDRESS"`
	PoolSize     int           `json:"pool_size" env:"REDIS_POOL_SIZE"`
	MinIdle      int           `json:"min_idle" env:"REDIS_MIN_IDLE"`
	PoolTimeout  time.Duration `json:"pool_timeout" env:"REDIS_POOL_TIMEOUT"`
	ReadTimeout  time.Duration `json:"read_timeout" env:"REDIS_READ_TIMEOUT"`
	WriteTimeout time.Duration `json:"write_timeout" env:"REDIS_WRITE_TIMEOUT"`
	DialTimeout  time.Duration `json:"dial_timeout" env:"REDIS_DIAL_TIMEOUT"`
}

func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Host:         "127.0.0.1",
		Port:         6379,
		Password:     "",
		DB:           0,
		Address:      "127.0.0.1:6379",
		PoolSize:     10,
		MinIdle:      2,
		PoolTimeout:  30 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		DialTimeout:  5 * time.Second,
	}
}
