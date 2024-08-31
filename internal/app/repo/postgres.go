package repo

import (
	"context"
	"database/sql"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app"
	"time"
)

const (
	IDKey           = "id"
	LoginKey        = "login"
	PasswordHashKey = "password_hash"
	UserIDKey       = "user_id"
	NumberKey       = "number"
	StatusKey       = "status"
	AccrualKey      = "accrual"
	OrderNumberKey  = "order_number"
	SumKey          = "sum"
	WithdrawKey     = "withdraw"

	CreateUserQuery     = `INSERT INTO "user" (login, password_hash) VALUES (@login, @password_hash) RETURNING id;`
	CreateBalanceQuery  = `INSERT INTO balance (user_id) VALUES (@user_id);`
	GetUserQuery        = `SELECT id FROM "user" WHERE login = @login AND password_hash = @password_hash;`
	CreateOrderQuery    = `INSERT INTO "order" (user_id, number) VALUES (@user_id, @number) RETURNING id, status, uploaded_at;`
	UpdateOrderQuery    = `UPDATE "order" SET status = @status, accrual = @accrual WHERE id = @id;`
	GetOrdersQuery      = `SELECT id, number, status, accrual, uploaded_at FROM "order" WHERE user_id = @user_id;`
	GetNewOrdersQuery   = `SELECT id, user_id, number, status, accrual, uploaded_at FROM "order" WHERE status = 'NEW';`
	GetBalanceQuery     = `SELECT id, current, withdrawn FROM balance WHERE user_id = @user_id;`
	UpdateBalanceQuery  = `UPDATE balance SET current = current - @withdraw, withdrawn = withdrawn + @withdraw WHERE user_id = @user_id;`
	CreateWithdrawQuery = `INSERT INTO withdrawal (user_id, order_number, sum) VALUES (@user_id, @order_number, @sum) RETURNING id, processed_at;`
	GetWithdrawalsQuery = `SELECT id, order_number, sum, processed_at FROM withdrawal WHERE user_id = @user_id;`
)

type AppRepo struct {
	db *sql.DB
}

func NewAppRepo(db *sql.DB) (*AppRepo, error) {
	return &AppRepo{db: db}, nil
}

func (ar *AppRepo) CreateUser(ctx context.Context, login, passwordHash string) (*app.User, error) {
	// запускаем транзакцию
	tx, err := ar.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	// в случае неуспешного коммита все изменения транзакции будут отменены
	defer tx.Rollback()

	row := tx.QueryRowContext(
		ctx,
		CreateUserQuery,
		sql.Named(LoginKey, login),
		sql.Named(PasswordHashKey, passwordHash),
	)

	var id uint
	err = row.Scan(&id)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, CreateBalanceQuery, sql.Named(UserIDKey, id))
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &app.User{
		ID:           id,
		Login:        login,
		PasswordHash: passwordHash,
	}, nil
}

func (ar *AppRepo) AuthUser(ctx context.Context, login, passwordHash string) (*app.User, error) {
	row := ar.db.QueryRowContext(
		ctx,
		GetUserQuery,
		sql.Named(LoginKey, login),
		sql.Named(PasswordHashKey, passwordHash),
	)

	var id uint
	err := row.Scan(&id)
	if err != nil {
		return nil, err
	}

	return &app.User{
		ID:           id,
		Login:        login,
		PasswordHash: passwordHash,
	}, nil
}

func (ar *AppRepo) CreateOrder(ctx context.Context, userID uint, number string) (*app.Order, error) {
	row := ar.db.QueryRowContext(
		ctx,
		CreateOrderQuery,
		sql.Named(UserIDKey, userID),
		sql.Named(NumberKey, number),
	)

	var id uint
	var status string
	var uploadedAt time.Time
	err := row.Scan(&id, &status, &uploadedAt)
	if err != nil {
		return nil, err
	}

	return &app.Order{
		ID:         id,
		UserID:     userID,
		Number:     number,
		Status:     status,
		Accrual:    nil,
		UploadedAt: uploadedAt,
	}, nil
}

func (ar *AppRepo) UpdateOrder(ctx context.Context, order *app.Order) error {
	_, err := ar.db.ExecContext(
		ctx,
		UpdateOrderQuery,
		sql.Named(StatusKey, order.Status),
		sql.Named(AccrualKey, order.Accrual),
		sql.Named(IDKey, order.ID),
	)

	return err
}

func (ar *AppRepo) GetOrders(ctx context.Context, userID uint) ([]*app.Order, error) {
	rows, err := ar.db.QueryContext(
		ctx,
		GetOrdersQuery,
		sql.Named(UserIDKey, userID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []*app.Order{}
	for rows.Next() {
		var (
			id         uint
			number     string
			status     string
			accrual    *float64
			uploadedAt time.Time
		)

		err = rows.Scan(
			&id,
			&number,
			&status,
			&accrual,
			&uploadedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &app.Order{
			ID:         id,
			UserID:     userID,
			Number:     number,
			Status:     status,
			Accrual:    accrual,
			UploadedAt: uploadedAt,
		})
	}

	return orders, nil
}

func (ar *AppRepo) GetNewOrders(ctx context.Context) ([]*app.Order, error) {
	rows, err := ar.db.QueryContext(
		ctx,
		GetNewOrdersQuery,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []*app.Order{}
	for rows.Next() {
		var (
			id         uint
			userID     uint
			number     string
			status     string
			accrual    *float64
			uploadedAt time.Time
		)

		err = rows.Scan(
			&id,
			&userID,
			&number,
			&status,
			&accrual,
			&uploadedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &app.Order{
			ID:         id,
			UserID:     userID,
			Number:     number,
			Status:     status,
			Accrual:    accrual,
			UploadedAt: uploadedAt,
		})
	}

	return orders, nil
}

func (ar *AppRepo) GetBalance(ctx context.Context, userID uint) (*app.Balance, error) {
	row := ar.db.QueryRowContext(
		ctx,
		GetBalanceQuery,
		sql.Named(UserIDKey, userID),
	)

	var id uint
	var current float64
	var withdrawn float64
	err := row.Scan(&id, &current, &withdrawn)
	if err != nil {
		return nil, err
	}

	return &app.Balance{
		ID:        id,
		UserID:    userID,
		Current:   current,
		Withdrawn: withdrawn,
	}, nil
}

func (ar *AppRepo) CreateWithdraw(ctx context.Context, userID uint, orderNumber string, sum float64) (*app.Withdrawal, error) {
	// запускаем транзакцию
	tx, err := ar.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	// в случае неуспешного коммита все изменения транзакции будут отменены
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, UpdateBalanceQuery, sql.Named(UserIDKey, userID), sql.Named(WithdrawKey, sum))
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(
		ctx,
		CreateWithdrawQuery,
		sql.Named(UserIDKey, userID),
		sql.Named(OrderNumberKey, orderNumber),
		sql.Named(SumKey, sum),
	)

	var id uint
	var processedAt time.Time
	err = row.Scan(&id, &processedAt)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &app.Withdrawal{
		ID:          id,
		UserID:      userID,
		OrderNumber: orderNumber,
		Sum:         sum,
		ProcessedAt: processedAt,
	}, nil
}

func (ar *AppRepo) GetWithdrawals(ctx context.Context, userID uint) ([]*app.Withdrawal, error) {
	rows, err := ar.db.QueryContext(
		ctx,
		GetWithdrawalsQuery,
		sql.Named(UserIDKey, userID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawals := []*app.Withdrawal{}
	for rows.Next() {
		var (
			id          uint
			orderNumber string
			sum         float64
			processedAt time.Time
		)

		err = rows.Scan(
			&id,
			&orderNumber,
			&sum,
			&processedAt,
		)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, &app.Withdrawal{
			ID:          id,
			UserID:      userID,
			OrderNumber: orderNumber,
			Sum:         sum,
			ProcessedAt: processedAt,
		})
	}

	return withdrawals, nil
}
