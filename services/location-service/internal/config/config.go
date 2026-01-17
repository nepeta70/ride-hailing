package config

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Redis  RedisConfig  `mapstructure:"redis"`
	Logic  LogicConfig  `mapstructure:"logic"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type LogicConfig struct {
	GeohashPrecision   int `mapstructure:"geohash_precision"`
	LocationTTLSeconds int `mapstructure:"location_ttl_seconds"`
}
