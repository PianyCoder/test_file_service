package config

type ServerConfig struct {
	Addr string `env:"SERVER_ADDR" envDefault:"0.0.0.0:8000"`
}
