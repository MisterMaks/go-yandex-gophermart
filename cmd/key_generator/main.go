package main

import (
	"errors"
	"fmt"
	"github.com/MisterMaks/go-yandex-gophermart/internal/logger"
	"go.uber.org/zap"
	"log"
	"math/rand"
)

const (
	Size     = 16
	LogLevel = "INFO"

	ConfigKey string = "config"

	Symbols      string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 -_"
	CountSymbols        = len(Symbols)
)

var ErrZeroSize = errors.New("key size <= 0")

func generateRandomKey(size int) (string, error) {
	if size <= 0 {
		return "", ErrZeroSize
	}

	b := make([]byte, size)
	for i := range b {
		b[i] = Symbols[rand.Intn(CountSymbols)]
	}

	return string(b), nil
}

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

	// создаём случайный ключ
	key, err := generateRandomKey(config.Size)
	if err != nil {
		logger.Log.Fatal("Failed to generate key", zap.Error(err))
	}

	fmt.Println("Key:", string(key))
}
