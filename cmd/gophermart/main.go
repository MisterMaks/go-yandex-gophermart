package main

import (
	"context"
	"database/sql"
	"github.com/MisterMaks/go-yandex-gophermart/internal/accrual_system"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app/delivery"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app/gzip"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app/repo"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app/usecase"
	"github.com/MisterMaks/go-yandex-gophermart/internal/logger"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	RunAddress = ":8080"
	LogLevel   = "INFO"

	ConfigKey string = "config"

	RunAddressKey string = "run_address"
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

type Middlewares struct {
	RequestLogger  func(http.Handler) http.Handler
	AuthMiddleware func(http.Handler) http.Handler
	GzipMiddleware func(http.Handler) http.Handler
}

type AppHandlerInterface interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	CreateOrder(w http.ResponseWriter, r *http.Request)
	GetOrders(w http.ResponseWriter, r *http.Request)
	GetBalance(w http.ResponseWriter, r *http.Request)
	CreateWithdrawal(w http.ResponseWriter, r *http.Request)
	GetWithdrawals(w http.ResponseWriter, r *http.Request)
}

func router(
	appHandler AppHandlerInterface,
	middlewares *Middlewares,
) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.RequestLogger, middlewares.GzipMiddleware)
	r.Post(`/api/user/register`, appHandler.Register)
	r.Post(`/api/user/login`, appHandler.Login)
	r.Route(`/api/user`, func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware)
		r.Post(`/orders`, appHandler.CreateOrder)
		r.Get(`/orders`, appHandler.GetOrders)
		r.Get(`/balance`, appHandler.GetBalance)
		r.Post(`/balance/withdraw`, appHandler.CreateWithdrawal)
		r.Get(`/withdrawals`, appHandler.GetWithdrawals)
	})

	return r
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

	appRepo, err := repo.NewAppRepo(db)
	if err != nil {
		logger.Log.Fatal("Failed to create AppRepo",
			zap.Error(err),
		)
	}

	accrualSystemClient := accrual_system.NewAccrualSystemClient(config.AccrualSystemAddress, config.AccrualSystemRequestTimeout)
	appUsecase, err := usecase.NewAppUsecase(
		appRepo,
		accrualSystemClient,
		config.PasswordKey,
		config.TokenKey,
		config.TokenExpiration,
		config.ProcessOrderChanSize,
		config.ProcessOrderWaitingTime,
		config.UpdateExistedNewOrdersWaitingTime,
	)
	if err != nil {
		logger.Log.Fatal("Failed to create AppUsecase",
			zap.Error(err),
		)
	}
	defer appUsecase.Close()

	appHandler := delivery.NewAppHandler(appUsecase)

	middlewares := &Middlewares{
		RequestLogger:  logger.RequestLoggerMiddleware,
		AuthMiddleware: appHandler.AuthMiddleware,
		GzipMiddleware: gzip.GzipMiddleware,
	}

	r := router(appHandler, middlewares)

	logger.Log.Info("Server running",
		zap.String(RunAddressKey, config.RunAddress),
	)
	go func() {
		err = http.ListenAndServe(config.RunAddress, r)

		if err != nil {
			logger.Log.Fatal("Failed to start server",
				zap.Error(err),
			)
		}
	}()

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)

Loop:
	for {
		select {
		case exitSyg := <-exitChan:
			logger.Log.Info("terminating: via signal", zap.Any("signal", exitSyg))
			break Loop
		}
	}

	return
}
