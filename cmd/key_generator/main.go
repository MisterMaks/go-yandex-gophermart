package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/MisterMaks/go-yandex-gophermart/internal/logger"
	"go.uber.org/zap"
	"log"
)

const (
	Size     = 16
	LogLevel = "INFO"

	ConfigKey string = "config"
)

var ErrZeroSize = errors.New("key size <= 0")

func generateRandomKey(size int) ([]byte, error) {
	if size <= 0 {
		return nil, ErrZeroSize
	}

	// генерируем случайную последовательность байт
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
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

	fmt.Printf("Key: %x\n", key)
}
