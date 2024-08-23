package main

import (
	"context"
	"database/sql"
	"github.com/MisterMaks/go-yandex-gophermart/internal/logger"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	RunAddress = "127.0.0.1:8080"
	LogLevel   = "INFO"

	ConfigKey string = "config"
)

func migrate(dsn string) error {
	db, err := goose.OpenDBWithDriver("postgres", dsn)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Log.Fatal("Failed to close DB",
				zap.Error(err),
			)
		}
	}()
	ctx := context.Background()
	return goose.RunContext(ctx, "up", db, "./migrations/")
}

func connectPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		logger.Log.Error("Failed to ping DB Postgres",
			zap.Error(err),
		)
	}
	return db, nil
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

	db, err := connectPostgres(config.DatabaseURI)
	if err != nil {
		logger.Log.Fatal("Failed to connect to Postgres",
			zap.Error(err),
		)
	}
	defer db.Close()

	logger.Log.Info("Applying migrations")
	err = migrate(config.DatabaseURI)
	if err != nil {
		logger.Log.Fatal("Failed to apply migrations",
			zap.Error(err),
		)
	}
}
