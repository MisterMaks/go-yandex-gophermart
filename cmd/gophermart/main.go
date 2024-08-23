package main

import (
	"github.com/MisterMaks/go-yandex-gophermart/internal/logger"
	"go.uber.org/zap"
	"log"
)

const (
	RunAddress = "127.0.0.1:8080"
	LogLevel   = "INFO"

	ConfigKey string = "config"
)

func main() {
	config, err := NewConfig()
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to get config. Error:", err)
	}

	err = logger.Initialize(config.LogLevel)
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to init logger. Error:", err)
	}

	logger.Log.Debug("Config data",
		zap.Any(ConfigKey, config),
	)
}
