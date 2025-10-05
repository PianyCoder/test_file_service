package config

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"time"
)

type Config struct {
	Server  ServerConfig  `envPrefix:"SERVER_"`
	Storage StorageConfig `envPrefix:"STORAGE_"`
}

type ServerConfig struct {
	ListenAddress string        `env:"LISTEN_ADDRESS,:"`
	ReadTimeout   time.Duration `env:"READ_TIMEOUT,5s"`
	WriteTimeout  time.Duration `env:"WRITE_TIMEOUT,10s"`
}

type StorageConfig struct {
	Directory     string `env:"DIRECTORY,./storage_data"`
	MaxUploadSize int    `env:"MAX_UPLOAD_SIZE,20971520"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf(".env file not found or could not be loaded: %w", err)
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return &cfg, nil
}
