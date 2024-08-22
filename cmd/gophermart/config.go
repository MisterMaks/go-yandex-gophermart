package main

import (
	"flag"
	"github.com/caarlos0/env"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewConfig() (*Config, error) {
	config := &Config{}

	flag.StringVar(&config.RunAddress, "a", "", "Run address")
	flag.StringVar(&config.DatabaseURI, "d", "", "Database URI")
	flag.StringVar(&config.AccrualSystemAddress, "r", "", "Accrual system address")
	flag.Parse()

	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	if config.RunAddress == "" {
		config.RunAddress = RunAddress
	}

	return config, nil
}
