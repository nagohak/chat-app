package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Http     `yaml:"http"`
		Redis    `yaml:"redis"`
		Postgres `yaml:"postgres"`
	}
	Http struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}
	Redis struct {
		Host string `env-required:"true" yaml:"host" env:"REDIS_HOST"`
		Port string `env-required:"true" yaml:"port" env:"REDIS_PORT"`
	}
	Postgres struct {
		Host     string `env-required:"true" yaml:"host" env:"POSTGRES_HOST"`
		Port     string `env-required:"true" yaml:"port" env:"POSTGRES_PORT"`
		Db       string `env-required:"true" yaml:"db" env:"POSTGRES_DB"`
		User     string `env-required:"true" yaml:"user" env:"POSTGRES_USER"`
		Password string `env-required:"true" yaml:"password" env:"POSTGRES_PASSWORD"`
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
