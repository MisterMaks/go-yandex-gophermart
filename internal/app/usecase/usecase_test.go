package usecase

import (
	"context"
	"database/sql"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app/usecase/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAppUsecase_Register(t *testing.T) {
	login := "login"
	password := "password"
	invalidLogin := "invalid_login?"
	existedLogin := "existed_login"
	invalidPassword := "invalid_password?"

	user := &app.User{
		ID:           1,
		Login:        login,
		PasswordHash: "password_hash",
	}

	type args struct {
		login    string
		password string
	}

	type want struct {
		user *app.User
		err  error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid data",
			args: args{
				login:    login,
				password: password,
			},
			want: want{
				user: user,
				err:  nil,
			},
		},
		{
			name: "invalid login",
			args: args{
				login:    invalidLogin,
				password: password,
			},
			want: want{
				user: nil,
				err:  app.ErrInvalidLoginPasswordFormat,
			},
		},
		{
			name: "login taken",
			args: args{
				login:    existedLogin,
				password: password,
			},
			want: want{
				user: nil,
				err:  app.ErrLoginTaken,
			},
		},
		{
			name: "invalid password",
			args: args{
				login:    login,
				password: invalidPassword,
			},
			want: want{
				user: nil,
				err:  app.ErrInvalidLoginPasswordFormat,
			},
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppRepoInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().CreateUser(gomock.Any(), login, gomock.Any()).Return(user, nil).AnyTimes()

	m.EXPECT().CreateUser(gomock.Any(), existedLogin, gomock.Any()).Return(nil, &pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint \"user_login_key\""})

	appUsecase := &AppUsecase{
		AppRepo:                      m,
		AccrualSystemClient:          nil,
		passwordKey:                  "",
		tokenKey:                     "",
		tokenExp:                     0,
		processOrdersChan:            nil,
		processOrdersTicker:          nil,
		updateExistedNewOrdersTicker: nil,
		processOrdersCtx:             nil,
		processOrdersCtxCancel:       nil,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := appUsecase.Register(nil, tt.args.login, tt.args.password)
			assert.ErrorIs(t, err, tt.want.err)
			if err == nil {
				assert.Equal(t, tt.want.user, u)
			}
		})
	}
}

func TestAppUsecase_Login(t *testing.T) {
	login := "login"
	password := "password"
	invalidLogin := "invalid_login?"
	invalidPassword := "invalid_password?"
	incorrectPassword := "incorrect_password"

	user := &app.User{
		ID:           1,
		Login:        login,
		PasswordHash: "password_hash",
	}

	type args struct {
		login    string
		password string
	}

	type want struct {
		user *app.User
		err  error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid data",
			args: args{
				login:    login,
				password: password,
			},
			want: want{
				user: user,
				err:  nil,
			},
		},
		{
			name: "invalid login",
			args: args{
				login:    invalidLogin,
				password: password,
			},
			want: want{
				user: nil,
				err:  app.ErrInvalidLoginPasswordFormat,
			},
		},
		{
			name: "incorrect password",
			args: args{
				login:    login,
				password: incorrectPassword,
			},
			want: want{
				user: nil,
				err:  app.ErrInvalidLoginPassword,
			},
		},
		{
			name: "invalid password",
			args: args{
				login:    login,
				password: invalidPassword,
			},
			want: want{
				user: nil,
				err:  app.ErrInvalidLoginPasswordFormat,
			},
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppRepoInterface(ctrl)

	appUsecase := &AppUsecase{
		AppRepo:                      m,
		AccrualSystemClient:          nil,
		passwordKey:                  "",
		tokenKey:                     "",
		tokenExp:                     0,
		processOrdersChan:            nil,
		processOrdersTicker:          nil,
		updateExistedNewOrdersTicker: nil,
		processOrdersCtx:             nil,
		processOrdersCtxCancel:       nil,
	}

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().AuthUser(gomock.Any(), login, appUsecase.hashPassword(password)).Return(user, nil)

	m.EXPECT().AuthUser(gomock.Any(), login, gomock.Not(appUsecase.hashPassword(password))).Return(nil, sql.ErrNoRows)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := appUsecase.Login(nil, tt.args.login, tt.args.password)
			assert.ErrorIs(t, err, tt.want.err)
			if err == nil {
				assert.Equal(t, tt.want.user, u)
			}
		})
	}
}

func TestAppUsecase_BuildJWTString(t *testing.T) {
	userID := uint(1)

	type args struct {
		userID uint
	}

	type want struct {
		userID  uint
		wantErr bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid data",
			args: args{
				userID: userID,
			},
			want: want{
				userID:  userID,
				wantErr: false,
			},
		},
	}

	appUsecase := &AppUsecase{
		AppRepo:                      nil,
		AccrualSystemClient:          nil,
		passwordKey:                  "",
		tokenKey:                     "",
		tokenExp:                     1 * time.Second,
		processOrdersChan:            nil,
		processOrdersTicker:          nil,
		updateExistedNewOrdersTicker: nil,
		processOrdersCtx:             nil,
		processOrdersCtxCancel:       nil,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, errBuildJWTString := appUsecase.BuildJWTString(nil, tt.args.userID)
			if tt.want.wantErr {
				assert.Error(t, errBuildJWTString)
			} else {
				assert.NoError(t, errBuildJWTString)
			}
			userIDFromToken, errGetUserID := appUsecase.GetUserID(token)
			require.NoError(t, errGetUserID)
			if errBuildJWTString == nil && errGetUserID == nil {
				assert.Equal(t, tt.want.userID, userIDFromToken)
			}
		})
	}

	// Check expired token
	token, err := appUsecase.BuildJWTString(nil, userID)
	assert.NoError(t, err)
	time.Sleep(appUsecase.tokenExp + 1)
	_, err = appUsecase.GetUserID(token)
	assert.Error(t, err)
}

func TestAppUsecase_GetUserID(t *testing.T) {
	appUsecase := &AppUsecase{
		AppRepo:                      nil,
		AccrualSystemClient:          nil,
		passwordKey:                  "",
		tokenKey:                     "",
		tokenExp:                     1 * time.Second,
		processOrdersChan:            nil,
		processOrdersTicker:          nil,
		updateExistedNewOrdersTicker: nil,
		processOrdersCtx:             nil,
		processOrdersCtxCancel:       nil,
	}

	userID := uint(1)
	token, err := appUsecase.BuildJWTString(nil, userID)
	require.NoError(t, err)
	invalidToken := "invalid_token"

	type args struct {
		token string
	}

	type want struct {
		userID  uint
		wantErr bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid data",
			args: args{
				token: token,
			},
			want: want{
				userID:  userID,
				wantErr: false,
			},
		},
		{
			name: "invalid token",
			args: args{
				token: invalidToken,
			},
			want: want{
				userID:  userID,
				wantErr: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := appUsecase.GetUserID(tt.args.token)
			if tt.want.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if err == nil {
				assert.Equal(t, tt.want.userID, userID)
			}
		})
	}

	// Check expired token
	time.Sleep(appUsecase.tokenExp + 1)
	_, err = appUsecase.GetUserID(token)
	assert.Error(t, err)
}

func TestAppUsecase_CreateOrder(t *testing.T) {
	userID := uint(1)
	number := "4561261212345467"
	invalidNumber := "4561261212345464"
	uploadedNumber := "4561261212345566"
	uploadedByAnotherUserNumber := "4661261212345565"

	order := &app.Order{
		ID:         1,
		UserID:     userID,
		Number:     number,
		Status:     "NEW",
		Accrual:    nil,
		UploadedAt: time.Now(),
	}

	ctx, cancel := context.WithCancel(context.Background())

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppRepoInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().CreateOrder(gomock.Any(), userID, number).Return(order, nil)
	m.EXPECT().CreateOrder(gomock.Any(), userID, uploadedNumber).Return(nil, sql.ErrNoRows)
	m.EXPECT().CreateOrder(gomock.Any(), userID, uploadedByAnotherUserNumber).
		Return(
			nil,
			&pgconn.PgError{
				Code:    "23505",
				Message: "duplicate key value violates unique constraint \"order_number_key\"",
			},
		)

	appUsecase := &AppUsecase{
		AppRepo:                      m,
		AccrualSystemClient:          nil,
		passwordKey:                  "",
		tokenKey:                     "",
		tokenExp:                     0,
		processOrdersChan:            make(chan *app.Order, 1),
		processOrdersTicker:          nil,
		updateExistedNewOrdersTicker: nil,
		processOrdersCtx:             ctx,
		processOrdersCtxCancel:       cancel,
	}

	type args struct {
		userID uint
		number string
	}

	type want struct {
		order *app.Order
		err   error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid data",
			args: args{
				userID: userID,
				number: number,
			},
			want: want{
				order: order,
				err:   nil,
			},
		},
		{
			name: "invalid number",
			args: args{
				userID: userID,
				number: invalidNumber,
			},
			want: want{
				order: nil,
				err:   app.ErrInvalidOrderNumber,
			},
		},
		{
			name: "uploaded number",
			args: args{
				userID: userID,
				number: uploadedNumber,
			},
			want: want{
				order: nil,
				err:   app.ErrOrderUploaded,
			},
		},
		{
			name: "number uploaded by another user",
			args: args{
				userID: userID,
				number: uploadedByAnotherUserNumber,
			},
			want: want{
				order: nil,
				err:   app.ErrOrderUploadedByAnotherUser,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := appUsecase.CreateOrder(nil, tt.args.userID, tt.args.number)
			assert.ErrorIs(t, err, tt.want.err)
			if err == nil {
				assert.Equal(t, tt.want.order, o)
				oFromChan := <-appUsecase.processOrdersChan
				assert.Equal(t, tt.want.order, oFromChan)
			}
		})
	}
}

func TestAppUsecase_CreateWithdrawal(t *testing.T) {
	userID := uint(1)
	number := "4561261212345467"
	invalidNumber := "4561261212345464"
	uploadedNumber := "4561261212345566"
	uploadedByAnotherUserNumber := "4661261212345565"
	sum := float64(100)
	bigSum := float64(1000)

	withdrawal := &app.Withdrawal{
		ID:          1,
		UserID:      userID,
		OrderNumber: number,
		Sum:         sum,
		ProcessedAt: time.Now(),
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppRepoInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().CreateWithdrawal(gomock.Any(), userID, number, sum).Return(withdrawal, nil)
	m.EXPECT().CreateWithdrawal(gomock.Any(), userID, uploadedNumber, gomock.Any()).Return(nil, sql.ErrNoRows)
	m.EXPECT().CreateWithdrawal(gomock.Any(), userID, uploadedByAnotherUserNumber, gomock.Any()).
		Return(
			nil,
			&pgconn.PgError{
				Code:    "23505",
				Message: "duplicate key value violates unique constraint \"order_number_key\"",
			},
		)
	m.EXPECT().CreateWithdrawal(gomock.Any(), userID, number, bigSum).
		Return(
			nil,
			&pgconn.PgError{
				Code:    "23514",
				Message: "new row for relation \"balance\" violates check constraint \"balance_current_check\"",
			},
		)

	appUsecase := &AppUsecase{
		AppRepo:                      m,
		AccrualSystemClient:          nil,
		passwordKey:                  "",
		tokenKey:                     "",
		tokenExp:                     0,
		processOrdersChan:            nil,
		processOrdersTicker:          nil,
		updateExistedNewOrdersTicker: nil,
		processOrdersCtx:             nil,
		processOrdersCtxCancel:       nil,
	}

	type args struct {
		userID uint
		number string
		sum    float64
	}

	type want struct {
		withdrawal *app.Withdrawal
		err        error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid data",
			args: args{
				userID: userID,
				number: number,
				sum:    sum,
			},
			want: want{
				withdrawal: withdrawal,
				err:        nil,
			},
		},
		{
			name: "invalid number",
			args: args{
				userID: userID,
				number: invalidNumber,
				sum:    sum,
			},
			want: want{
				withdrawal: nil,
				err:        app.ErrInvalidOrderNumber,
			},
		},
		{
			name: "uploaded number",
			args: args{
				userID: userID,
				number: uploadedNumber,
				sum:    sum,
			},
			want: want{
				withdrawal: nil,
				err:        app.ErrOrderUploaded,
			},
		},
		{
			name: "number uploaded by another user",
			args: args{
				userID: userID,
				number: uploadedByAnotherUserNumber,
				sum:    sum,
			},
			want: want{
				withdrawal: nil,
				err:        app.ErrOrderUploadedByAnotherUser,
			},
		},
		{
			name: "insufficient funds",
			args: args{
				userID: userID,
				number: number,
				sum:    bigSum,
			},
			want: want{
				withdrawal: nil,
				err:        app.ErrInsufficientFunds,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := appUsecase.CreateWithdrawal(nil, tt.args.userID, tt.args.number, tt.args.sum)
			assert.ErrorIs(t, err, tt.want.err)
			if err == nil {
				assert.Equal(t, tt.want.withdrawal, w)
			}
		})
	}
}
