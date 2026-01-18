package redis

import (
	"os"
	"time"
)

type RedisConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Password     string        `json:"password"`
	DB           int           `json:"db"`
	Address      string        `json:"address"`
	PoolSize     int           `json:"pool_size"`
	MinIdle      int           `json:"min_idle"`
	PoolTimeout  time.Duration `json:"pool_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	DialTimeout  time.Duration `json:"dial_timeout"`
}

// OverrideFromEnv sets RedisConfig fields from environment variables if present.
func (c *RedisConfig) OverrideFromEnv() {
	if redisAddr := os.Getenv("REDIS_ADDRESS"); redisAddr != "" {
		c.Address = redisAddr
	}
	if redisPass := os.Getenv("REDIS_PASSWORD"); redisPass != "" {
		c.Password = redisPass
	}
}
