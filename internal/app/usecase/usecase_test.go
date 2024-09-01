package usecase

import (
	"context"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app/usecase/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"testing"
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
		ctx      context.Context
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
				ctx:      nil,
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
				ctx:      nil,
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
				ctx:      nil,
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
				ctx:      nil,
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

	//m.EXPECT().CreateUser(gomock.Any(), invalidLogin, gomock.Any()).Return(nil, &pgconn.PgError{Code: "23514", Message: "new row for relation \"user\" violates check constraint \"user_login_check\""})
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
			u, err := appUsecase.Register(tt.args.ctx, tt.args.login, tt.args.password)
			assert.ErrorIs(t, err, tt.want.err)
			if err != nil {
				assert.Equal(t, tt.want.user, u)
			}
		})
	}
}
