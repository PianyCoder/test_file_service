package config

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	ServerConfig  ServerConfig  `envPrefix:"SERVER_"`
	ServiceConfig ServiceConfig `envPrefix:"SERVICE_"`
	MinioConfig   MinioConfig   `envPrefix:"MINIO_"`
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
