package repo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var DSN = os.Getenv("TEST_DATABASE_URI")

func upMigrations(dsn string) error {
	db, err := goose.OpenDBWithDriver("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx := context.Background()
	return goose.RunContext(ctx, "up", db, "../../../migrations/")
}

func downMigrations(dsn string) error {
	db, err := goose.OpenDBWithDriver("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
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
		return nil, err
	}
	return db, nil
}

func newTestEnvironment(dsn string, t *testing.T) *TestEnvironment {
	if testing.Short() {
		t.Skip()
	}

	err := upMigrations(dsn)
	require.NoError(t, err, "Failed to apply migrations")

	db, err := connectPostgres(DSN)
	require.NoError(t, err, "Failed to connect to Postgres")

	return &TestEnvironment{
		DSN: dsn,
		T:   t,
		DB:  db,
	}
}

type TestEnvironment struct {
	DSN string
	T   *testing.T
	DB  *sql.DB
}

func (te *TestEnvironment) clean() {
	err := te.DB.Close()
	assert.NoError(te.T, err, "Failed to close DB")
	err = downMigrations(te.DSN)
	require.NoError(te.T, err, "Failed to roll back migrations")
}

func TestNewAppRepo(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	_, err := NewAppRepo(te.DB)
	assert.NoError(t, err, "Failed to run NewAppRepo()")
}

func TestAppRepo_CreateUser(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	appRepo, err := NewAppRepo(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepo()")

	ctx := context.Background()

	// valid data
	login := "login"
	passwordHash := "password_hash"
	user, err := appRepo.CreateUser(ctx, login, passwordHash)
	require.NoError(t, err)
	assert.Equal(t, user.Login, login)
	assert.Equal(t, user.PasswordHash, passwordHash)
	actualUser, err := appRepo.AuthUser(ctx, login, passwordHash)
	require.NoError(t, err)
	assert.Equal(t, user, actualUser)

	// login taken
	_, err = appRepo.CreateUser(ctx, login, passwordHash)
	require.Error(t, err)
	var pgErr *pgconn.PgError
	assert.True(t, errors.As(err, &pgErr))
	assert.Equal(t, "23505", pgErr.Code)
	assert.Equal(t, "duplicate key value violates unique constraint \"user_login_key\"", pgErr.Message)

	// invalid symbols
	_, err = appRepo.CreateUser(ctx, "$$$$", "!!!!")
	require.Error(t, err)
	assert.True(t, errors.As(err, &pgErr))
	assert.Equal(t, "23514", pgErr.Code)
	assert.Equal(t, "new row for relation \"user\" violates check constraint \"user_login_check\"", pgErr.Message)
}

func TestAppRepo_AuthUser(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	appRepo, err := NewAppRepo(te.DB)
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

func TestAppRepo_CreateOrder(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	appRepo, err := NewAppRepo(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepo()")

	ctx := context.Background()

	login := "login"
	passwordHash := "password_hash"
	user, err := appRepo.CreateUser(ctx, login, passwordHash)
	require.NoError(t, err)

	actualOrders, err := appRepo.GetOrders(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, []*app.Order{}, actualOrders)

	// valid data
	number := "12345"
	expectedStatus := "NEW"
	var expectedAccrual *float64
	order, err := appRepo.CreateOrder(ctx, user.ID, number)
	require.NoError(t, err)
	assert.Equal(t, number, order.Number)
	assert.Equal(t, user.ID, order.UserID)
	assert.Equal(t, expectedStatus, order.Status)
	assert.Equal(t, expectedAccrual, order.Accrual)

	actualOrders, err = appRepo.GetOrders(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, []*app.Order{order}, actualOrders)

	number2 := "67890"
	order2, err := appRepo.CreateOrder(ctx, user.ID, number2)
	require.NoError(t, err)

	actualOrders, err = appRepo.GetOrders(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, []*app.Order{order2, order}, actualOrders)

	// create existed order
	_, err = appRepo.CreateOrder(ctx, user.ID, number)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)

	// create existed order by another user
	login2 := "login2"
	user, err = appRepo.CreateUser(ctx, login2, passwordHash)
	require.NoError(t, err)

	_, err = appRepo.CreateOrder(ctx, user.ID, number)
	require.Error(t, err)
	var pgErr *pgconn.PgError
	require.True(t, errors.As(err, &pgErr))
	assert.Equal(t, "23505", pgErr.Code)
	assert.Equal(t, "duplicate key value violates unique constraint \"order_number_key\"", pgErr.Message)
}

func TestAppRepo_UpdateOrder(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	appRepo, err := NewAppRepo(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepo()")

	ctx := context.Background()

	login := "login"
	passwordHash := "password_hash"
	user, err := appRepo.CreateUser(ctx, login, passwordHash)
	require.NoError(t, err)

	number := "12345"
	order, err := appRepo.CreateOrder(ctx, user.ID, number)
	require.NoError(t, err)

	balance, err := appRepo.GetBalance(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, float64(0), balance.Current)

	actualOrders, err := appRepo.GetOrders(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, []*app.Order{order}, actualOrders)

	order.Status = "PROCESSED"
	accrual := float64(100)
	order.Accrual = &accrual

	err = appRepo.UpdateOrder(ctx, order)
	require.NoError(t, err)

	actualOrders, err = appRepo.GetOrders(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, []*app.Order{order}, actualOrders)

	balance, err = appRepo.GetBalance(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, accrual, balance.Current)
}

func TestAppRepo_GetOrders(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	appRepo, err := NewAppRepo(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepo()")

	ctx := context.Background()

	login := "login"
	passwordHash := "password_hash"
	user, err := appRepo.CreateUser(ctx, login, passwordHash)
	require.NoError(t, err)

	actualOrders, err := appRepo.GetOrders(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, []*app.Order{}, actualOrders)

	number := "12345"
	order, err := appRepo.CreateOrder(ctx, user.ID, number)
	require.NoError(t, err)

	actualOrders, err = appRepo.GetOrders(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, []*app.Order{order}, actualOrders)

	number2 := "67890"
	order2, err := appRepo.CreateOrder(ctx, user.ID, number2)
	require.NoError(t, err)

	order2.Status = "PROCESSED"
	accrual := float64(100)
	order2.Accrual = &accrual

	err = appRepo.UpdateOrder(ctx, order2)
	require.NoError(t, err)

	actualOrders, err = appRepo.GetOrders(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, []*app.Order{order2, order}, actualOrders)
}

func TestAppRepo_GetNewOrders(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	appRepo, err := NewAppRepo(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepo()")

	ctx := context.Background()

	login := "login"
	passwordHash := "password_hash"
	user, err := appRepo.CreateUser(ctx, login, passwordHash)
	require.NoError(t, err)

	number := "12345"
	order, err := appRepo.CreateOrder(ctx, user.ID, number)
	require.NoError(t, err)

	number2 := "67890"
	order2, err := appRepo.CreateOrder(ctx, user.ID, number2)
	require.NoError(t, err)

	number3 := "11111"
	order3, err := appRepo.CreateOrder(ctx, user.ID, number3)
	require.NoError(t, err)

	order3.Status = "PROCESSED"
	accrual := float64(100)
	order3.Accrual = &accrual

	err = appRepo.UpdateOrder(ctx, order3)
	require.NoError(t, err)

	actualOrders, err := appRepo.GetNewOrders(ctx)
	require.NoError(t, err)
	assert.Equal(t, []*app.Order{order, order2}, actualOrders)
}

func TestAppRepo_GetBalance(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	appRepo, err := NewAppRepo(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepo()")

	ctx := context.Background()

	login := "login"
	passwordHash := "password_hash"
	user, err := appRepo.CreateUser(ctx, login, passwordHash)
	require.NoError(t, err)

	number := "12345"
	order, err := appRepo.CreateOrder(ctx, user.ID, number)
	require.NoError(t, err)

	number2 := "67890"
	order2, err := appRepo.CreateOrder(ctx, user.ID, number2)
	require.NoError(t, err)

	balance, err := appRepo.GetBalance(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, float64(0), balance.Current)

	order.Status = "PROCESSED"
	accrual := float64(100)
	order.Accrual = &accrual

	err = appRepo.UpdateOrder(ctx, order)
	require.NoError(t, err)

	balance, err = appRepo.GetBalance(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, accrual, balance.Current)

	order2.Status = "PROCESSED"
	accrual2 := float64(200)
	order2.Accrual = &accrual2

	err = appRepo.UpdateOrder(ctx, order2)
	require.NoError(t, err)

	balance, err = appRepo.GetBalance(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, accrual+accrual2, balance.Current)
}

func TestAppRepo_CreateWithdrawal(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	appRepo, err := NewAppRepo(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepo()")

	ctx := context.Background()

	login := "login"
	passwordHash := "password_hash"
	user, err := appRepo.CreateUser(ctx, login, passwordHash)
	require.NoError(t, err)

	withdrawals, err := appRepo.GetWithdrawals(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, []*app.Withdrawal{}, withdrawals)

	balance, err := appRepo.GetBalance(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, float64(0), balance.Current)

	orderNumber := "12345"
	sum := float64(100)

	// There are insufficient funds in the account
	_, err = appRepo.CreateWithdrawal(ctx, user.ID, orderNumber, sum)
	require.Error(t, err)
	var pgErr *pgconn.PgError
	assert.True(t, errors.As(err, &pgErr))
	assert.Equal(t, "23514", pgErr.Code)
	assert.Equal(t, "new row for relation \"balance\" violates check constraint \"balance_current_check\"", pgErr.Message)

	// Valid data
	order, err := appRepo.CreateOrder(ctx, user.ID, "67890")
	require.NoError(t, err)

	order.Status = "PROCESSED"
	accrual := float64(200)
	order.Accrual = &accrual

	err = appRepo.UpdateOrder(ctx, order)
	require.NoError(t, err)

	balance, err = appRepo.GetBalance(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, accrual, balance.Current)

	withdrawal, err := appRepo.CreateWithdrawal(ctx, user.ID, orderNumber, sum)
	require.NoError(t, err)
	assert.Equal(t, orderNumber, withdrawal.OrderNumber)
	assert.Equal(t, sum, withdrawal.Sum)

	balance, err = appRepo.GetBalance(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, accrual-sum, balance.Current)

	_, err = appRepo.CreateWithdrawal(ctx, user.ID, orderNumber, sum)
	assert.ErrorIs(t, err, sql.ErrNoRows)

	balance, err = appRepo.GetBalance(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, accrual-sum, balance.Current)

	// Order number has already been uploaded by another user
	login2 := "login_2"
	passwordHash2 := "password_hash_2"
	user2, err := appRepo.CreateUser(ctx, login2, passwordHash2)
	require.NoError(t, err)

	order, err = appRepo.CreateOrder(ctx, user2.ID, "11111")
	require.NoError(t, err)

	order.Status = "PROCESSED"
	accrual = float64(200)
	order.Accrual = &accrual

	err = appRepo.UpdateOrder(ctx, order)
	require.NoError(t, err)

	_, err = appRepo.CreateWithdrawal(ctx, user2.ID, orderNumber, sum)
	require.True(t, errors.As(err, &pgErr))
	assert.Equal(t, "23505", pgErr.Code)
	assert.Equal(t, "duplicate key value violates unique constraint \"withdrawal_order_number_key\"", pgErr.Message)

	balance, err = appRepo.GetBalance(ctx, user2.ID)
	require.NoError(t, err)
	assert.Equal(t, accrual, balance.Current)
}

func TestAppRepo_GetWithdrawals(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	appRepo, err := NewAppRepo(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepo()")

	ctx := context.Background()

	login := "login"
	passwordHash := "password_hash"
	user, err := appRepo.CreateUser(ctx, login, passwordHash)
	require.NoError(t, err)

	// Empty
	withdrawals, err := appRepo.GetWithdrawals(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, []*app.Withdrawal{}, withdrawals)

	// 1 withdrawal
	order, err := appRepo.CreateOrder(ctx, user.ID, "11111")
	require.NoError(t, err)

	order.Status = "PROCESSED"
	accrual := float64(200)
	order.Accrual = &accrual

	err = appRepo.UpdateOrder(ctx, order)
	require.NoError(t, err)

	orderNumber := "12345"
	sum := float64(100)
	withdrawal, err := appRepo.CreateWithdrawal(ctx, user.ID, orderNumber, sum)
	require.NoError(t, err)

	withdrawals, err = appRepo.GetWithdrawals(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, []*app.Withdrawal{withdrawal}, withdrawals)

	// 2 withdrawals
	orderNumber2 := "67890"
	sum2 := float64(100)
	withdrawal2, err := appRepo.CreateWithdrawal(ctx, user.ID, orderNumber2, sum2)
	require.NoError(t, err)

	withdrawals, err = appRepo.GetWithdrawals(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, []*app.Withdrawal{withdrawal2, withdrawal}, withdrawals)
}
