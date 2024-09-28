package repo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/MisterMaks/go-yandex-gophermart/internal/logger"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"os"
	"testing"
)

var DSN = os.Getenv("TEST_DSN")

func upMigrations(dsn string) error {
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
	return goose.RunContext(ctx, "up", db, "../../../migrations/")
}

func downMigrations(dsn string) error {
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
	return goose.RunContext(ctx, "down", db, "../../../migrations/")
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

func TestNewAppRepo(t *testing.T) {
	db, err := connectPostgres(DSN)
	require.NoError(t, err, "Failed to connect to Postgres")
	defer db.Close()

	err = upMigrations(DSN)
	require.NoError(t, err, "Failed to apply migrations")

	_, err = NewAppRepo(db)
	assert.NoError(t, err, "Failed to run NewAppRepo()")

	err = downMigrations(DSN)
	require.NoError(t, err, "Failed to roll back migrations")
}

func TestAppRepo_CreateUser(t *testing.T) {
	db, err := connectPostgres(DSN)
	require.NoError(t, err, "Failed to connect to Postgres")
	defer db.Close()

	err = upMigrations(DSN)
	require.NoError(t, err, "Failed to apply migrations")
	defer downMigrations(DSN)

	appRepo, err := NewAppRepo(db)
	require.NoError(t, err, "Failed to run NewAppRepo()")

	ctx := context.Background()

	// valid data
	login := "login"
	passwordHash := "password_hash"
	user, err := appRepo.CreateUser(ctx, login, passwordHash)
	require.NoError(t, err)
	assert.Equal(t, user.Login, login)
	assert.Equal(t, user.PasswordHash, passwordHash)

	// login taken
	_, err = appRepo.CreateUser(ctx, login, passwordHash)
	require.Error(t, err)
	var pgErr *pgconn.PgError
	require.True(t, errors.As(err, &pgErr))
	require.Equal(t, pgErr.Code, "23505")
	require.Equal(t, pgErr.Message, "duplicate key value violates unique constraint \"user_login_key\"")

	// invalid symbols
	_, err = appRepo.CreateUser(ctx, "$$$$", "!!!!")
	require.Error(t, err)
	require.True(t, errors.As(err, &pgErr))
	require.Equal(t, pgErr.Code, "23514")
	require.Equal(t, pgErr.Message, "new row for relation \"user\" violates check constraint \"user_login_check\"")
}

func TestAppRepo_AuthUser(t *testing.T) {
	db, err := connectPostgres(DSN)
	require.NoError(t, err, "Failed to connect to Postgres")
	defer db.Close()

	err = upMigrations(DSN)
	require.NoError(t, err, "Failed to apply migrations")
	defer downMigrations(DSN)

	appRepo, err := NewAppRepo(db)
	require.NoError(t, err, "Failed to run NewAppRepo()")

	ctx := context.Background()

	// valid data
	login := "login"
	passwordHash := "password_hash"
	expectedUser, err := appRepo.CreateUser(ctx, login, passwordHash)
	require.NoError(t, err)
	actualUser, err := appRepo.AuthUser(ctx, login, passwordHash)
	require.NoError(t, err)
	assert.Equal(t, expectedUser, actualUser)

	// invalid password hash
	_, err = appRepo.AuthUser(ctx, login, "invalid_password_hash")
	assert.ErrorIs(t, err, sql.ErrNoRows)

	// invalid data
	_, err = appRepo.AuthUser(ctx, "invalid_login", "invalid_password_hash")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}
