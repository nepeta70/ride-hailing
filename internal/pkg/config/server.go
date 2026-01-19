package config

type ServerConfig struct {
	Port int    `json:"port" env:"SERVER_PORT" envDefault:"50051"`
	Host string `json:"host" env:"SERVER_HOST" envDefault:"127.0.0.1"`
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Port: 50051,
		Host: "127.0.0.1",
	}
}
