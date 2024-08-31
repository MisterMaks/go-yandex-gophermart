package main

import (
	"flag"
	"github.com/caarlos0/env"
	"time"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`

	LogLevel                          string        `env:"LOG_LEVEL"`
	PasswordKey                       string        `env:"PASSWORD_KEY"`
	TokenKey                          string        `env:"TOKEN_KEY"`
	TokenExpiration                   time.Duration `env:"TOKEN_EXPIRATION"`
	ProcessOrderChanSize              uint          `env:"PROCESS_ORDER_CHAN_SIZE"`
	ProcessOrderWaitingTime           time.Duration `env:"PROCESS_ORDER_WAITING_TIME"`
	UpdateExistedNewOrdersWaitingTime time.Duration `env:"UPDATE_EXISTED_ORDERS_WAITING_TIME"`
}

func NewConfig() (*Config, error) {
	config := &Config{}

	flag.StringVar(&config.RunAddress, "a", "", "Run address")
	flag.StringVar(&config.DatabaseURI, "d", "", "Database URI")
	flag.StringVar(&config.AccrualSystemAddress, "r", "", "Accrual system address")

	flag.StringVar(&config.LogLevel, "l", "", "Log level")
	flag.StringVar(&config.PasswordKey, "pk", "", "Password key")
	flag.StringVar(&config.TokenKey, "tk", "", "Token key")
	flag.DurationVar(&config.TokenExpiration, "te", 15*time.Minute, "Token expiration")
	flag.UintVar(&config.ProcessOrderChanSize, "pocs", 256, "Process order chan size")
	flag.DurationVar(&config.ProcessOrderWaitingTime, "powt", 2*time.Minute, "Process order waiting time")
	flag.DurationVar(&config.UpdateExistedNewOrdersWaitingTime, "uenowt", 2*time.Minute, "Update existed new orders waiting Time")

	flag.Parse()

	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	if config.RunAddress == "" {
		config.RunAddress = RunAddress
	}
	if config.LogLevel == "" {
		config.LogLevel = LogLevel
	}

	return config, nil
}
