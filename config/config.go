package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Http  `yaml:"http"`
		Redis `yaml:"redis"`
	}
	Http struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}
	Redis struct {
		Url string `env-required:"true" yaml:"url" env:"REDIS_URL"`
	}
)

func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("./config/config.yml", cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
