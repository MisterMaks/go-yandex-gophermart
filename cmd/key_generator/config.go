package main

import (
	"flag"
	"github.com/caarlos0/env"
)

type Config struct {
	Size     int    `env:"SIZE"`
	LogLevel string `env:"LOG_LEVEL"`
}

func NewConfig() (*Config, error) {
	config := &Config{}

	flag.IntVar(&config.Size, "s", Size, "Key size")
	flag.StringVar(&config.LogLevel, "l", "", "Log level")
	flag.Parse()

	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	if config.LogLevel == "" {
		config.LogLevel = LogLevel
	}

	return config, nil
}
