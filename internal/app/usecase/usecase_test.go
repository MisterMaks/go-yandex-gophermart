package usecase

import (
	"context"
	"database/sql"
	"github.com/MisterMaks/go-yandex-gophermart/internal/accrual"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app/usecase/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewAppUsecase(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	mockARI := mocks.NewMockAppRepoInterface(ctrl)
	mockARI.EXPECT().GetNewOrders(gomock.Any()).Return(nil, nil).AnyTimes()

	mockASCI := mocks.NewMockAccrualSystemClientInterface(ctrl)
	mockASCI.EXPECT().GetOrderInfo(gomock.Any(), gomock.Any()).Return(accrual.OrderInfo{}, nil).AnyTimes()

	passwordKey := "12345"
	tokenKey := "00000"
	tokenExp := 10 * time.Second
	processOrderChanSize := uint(1)
	processOrderWaitingTime := time.Second
	updateExistedNewOrdersWaitingTime := time.Second

	processOrderCtx, processOrderCtxCancel := context.WithCancel(context.Background())

	appUsecase := &AppUsecase{
		AppRepo:                      mockARI,
		AccrualSystemClient:          mockASCI,
		passwordKey:                  passwordKey,
		tokenKey:                     tokenKey,
		tokenExp:                     tokenExp,
		processOrdersChan:            make(chan *app.Order, processOrderChanSize),
		processOrdersTicker:          time.NewTicker(processOrderWaitingTime),
		updateExistedNewOrdersTicker: time.NewTicker(updateExistedNewOrdersWaitingTime),
		processOrdersCtx:             processOrderCtx,
		processOrdersCtxCancel:       processOrderCtxCancel,
	}

	type args struct {
		minLoginLen                       uint
		passwordKey                       string
		minPasswordLen                    uint
		tokenKey                          string
		tokenExp                          time.Duration
		processOrderChanSize              uint
		processOrderWaitingTime           time.Duration
		updateExistedNewOrdersWaitingTime time.Duration
	}

	type want struct {
		appUsecase *AppUsecase
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
				minLoginLen:                       6,
				passwordKey:                       passwordKey,
				minPasswordLen:                    6,
				tokenKey:                          tokenKey,
				tokenExp:                          tokenExp,
				processOrderChanSize:              processOrderChanSize,
				processOrderWaitingTime:           processOrderWaitingTime,
				updateExistedNewOrdersWaitingTime: updateExistedNewOrdersWaitingTime,
			},
			want: want{
				appUsecase: appUsecase,
				err:        nil,
			},
		},
		{
			name: "empty password key",
			args: args{
				minLoginLen:                       6,
				passwordKey:                       "",
				minPasswordLen:                    6,
				tokenKey:                          tokenKey,
				tokenExp:                          tokenExp,
				processOrderChanSize:              processOrderChanSize,
				processOrderWaitingTime:           processOrderWaitingTime,
				updateExistedNewOrdersWaitingTime: updateExistedNewOrdersWaitingTime,
			},
			want: want{
				appUsecase: nil,
				err:        ErrEmptyPasswordKey,
			},
		},
		{
			name: "empty token key",
			args: args{
				minLoginLen:                       6,
				passwordKey:                       passwordKey,
				minPasswordLen:                    6,
				tokenKey:                          "",
				tokenExp:                          tokenExp,
				processOrderChanSize:              processOrderChanSize,
				processOrderWaitingTime:           processOrderWaitingTime,
				updateExistedNewOrdersWaitingTime: updateExistedNewOrdersWaitingTime,
			},
			want: want{
				appUsecase: nil,
				err:        ErrEmptyTokenKey,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			au, err := NewAppUsecase(
				mockARI,
				mockASCI,
				tt.args.minLoginLen,
				tt.args.passwordKey,
				tt.args.minPasswordLen,
				tt.args.tokenKey,
				tt.args.tokenExp,
				tt.args.processOrderChanSize,
				tt.args.processOrderWaitingTime,
				tt.args.updateExistedNewOrdersWaitingTime,
			)
			assert.ErrorIs(t, err, tt.want.err)
			if err == nil {
				assert.Equal(t, tt.want.appUsecase.passwordKey, au.passwordKey)
				assert.Equal(t, tt.want.appUsecase.tokenKey, au.tokenKey)
				assert.Equal(t, tt.want.appUsecase.tokenExp, au.tokenExp)
				assert.Equal(t, len(tt.want.appUsecase.processOrdersChan), len(au.processOrdersChan))
				assert.NotNil(t, au.processOrdersTicker)
				assert.NotNil(t, au.updateExistedNewOrdersTicker)
				assert.NotNil(t, au.processOrdersCtx)
				assert.NotNil(t, au.processOrdersCtxCancel)
				au.Close()
			}
		})
	}
}

func TestAppUsecase_Close(t *testing.T) {
	processOrderCtx, processOrderCtxCancel := context.WithCancel(context.Background())

	appUsecase := &AppUsecase{
		AppRepo:                      nil,
		AccrualSystemClient:          nil,
		passwordKey:                  "",
		tokenKey:                     "",
		tokenExp:                     0,
		processOrdersChan:            make(chan *app.Order, 1),
		processOrdersTicker:          nil,
		updateExistedNewOrdersTicker: time.NewTicker(time.Millisecond),
		processOrdersCtx:             processOrderCtx,
		processOrdersCtxCancel:       processOrderCtxCancel,
	}

	appUsecase.Close()

	_, ok := <-appUsecase.processOrdersChan
	assert.False(t, ok)

	select {
	case <-appUsecase.processOrdersCtx.Done():
		return
	case <-time.NewTicker(time.Second).C:
		assert.Fail(t, "context not closed")
	}
}

func TestAppUsecase_processOrders(t *testing.T) {
	orderRegisteredNumber := "1"
	orderInvalidNumber := "2"
	orderProcessingNumber := "3"
	orderProcessedNumber := "4"

	accrualSum := float64(100)

	uploadedAt := time.Now()

	type args struct {
		orders []*app.Order
	}

	type want struct {
		orders []*app.Order
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "registered order",
			args: args{
				orders: []*app.Order{
					{
						ID:         1,
						UserID:     1,
						Number:     orderRegisteredNumber,
						Status:     "NEW",
						Accrual:    nil,
						UploadedAt: uploadedAt,
					},
				},
			},
			want: want{
				orders: nil,
			},
		},
		{
			name: "invalid order",
			args: args{
				orders: []*app.Order{
					{
						ID:         2,
						UserID:     2,
						Number:     orderInvalidNumber,
						Status:     "NEW",
						Accrual:    nil,
						UploadedAt: uploadedAt,
					},
				},
			},
			want: want{
				orders: []*app.Order{
					{
						ID:         2,
						UserID:     2,
						Number:     orderInvalidNumber,
						Status:     "INVALID",
						Accrual:    nil,
						UploadedAt: uploadedAt,
					},
				},
			},
		},
		{
			name: "processing order",
			args: args{
				orders: []*app.Order{
					{
						ID:         3,
						UserID:     3,
						Number:     orderProcessingNumber,
						Status:     "NEW",
						Accrual:    nil,
						UploadedAt: uploadedAt,
					},
				},
			},
			want: want{
				orders: []*app.Order{
					{
						ID:         3,
						UserID:     3,
						Number:     orderProcessingNumber,
						Status:     "PROCESSING",
						Accrual:    nil,
						UploadedAt: uploadedAt,
					},
				},
			},
		},
		{
			name: "processed order",
			args: args{
				orders: []*app.Order{
					{
						ID:         4,
						UserID:     4,
						Number:     orderProcessedNumber,
						Status:     "NEW",
						Accrual:    nil,
						UploadedAt: uploadedAt,
					},
				},
			},
			want: want{
				orders: []*app.Order{
					{
						ID:         4,
						UserID:     4,
						Number:     orderProcessedNumber,
						Status:     "PROCESSED",
						Accrual:    &accrualSum,
						UploadedAt: uploadedAt,
					},
				},
			},
		},
		{
			name: "some orders",
			args: args{
				orders: []*app.Order{
					{
						ID:         3,
						UserID:     3,
						Number:     orderProcessingNumber,
						Status:     "NEW",
						Accrual:    nil,
						UploadedAt: uploadedAt,
					},
					{
						ID:         4,
						UserID:     4,
						Number:     orderProcessedNumber,
						Status:     "NEW",
						Accrual:    nil,
						UploadedAt: uploadedAt,
					},
				},
			},
			want: want{
				orders: []*app.Order{
					{
						ID:         3,
						UserID:     3,
						Number:     orderProcessingNumber,
						Status:     "PROCESSING",
						Accrual:    nil,
						UploadedAt: uploadedAt,
					},
					{
						ID:         4,
						UserID:     4,
						Number:     orderProcessedNumber,
						Status:     "PROCESSED",
						Accrual:    &accrualSum,
						UploadedAt: uploadedAt,
					},
				},
			},
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	mockARI := mocks.NewMockAppRepoInterface(ctrl)

	actualOrders := []*app.Order{}
	mockARI.EXPECT().UpdateOrder(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, order *app.Order) error {
			actualOrders = append(actualOrders, order)
			return nil
		},
	).AnyTimes()

	mockASCI := mocks.NewMockAccrualSystemClientInterface(ctrl)

	mockASCI.EXPECT().GetOrderInfo(
		gomock.Any(),
		orderRegisteredNumber,
	).Return(
		accrual.OrderInfo{
			Number:  orderRegisteredNumber,
			Status:  "REGISTERED",
			Accrual: nil,
		},
		nil,
	).AnyTimes()

	mockASCI.EXPECT().GetOrderInfo(
		gomock.Any(),
		orderInvalidNumber,
	).Return(
		accrual.OrderInfo{
			Number:  orderInvalidNumber,
			Status:  "INVALID",
			Accrual: nil,
		},
		nil,
	).AnyTimes()

	mockASCI.EXPECT().GetOrderInfo(
		gomock.Any(),
		orderProcessingNumber,
	).Return(
		accrual.OrderInfo{
			Number:  orderProcessingNumber,
			Status:  "PROCESSING",
			Accrual: nil,
		},
		nil,
	).AnyTimes()

	mockASCI.EXPECT().GetOrderInfo(
		gomock.Any(),
		orderProcessedNumber,
	).Return(
		accrual.OrderInfo{
			Number:  orderProcessedNumber,
			Status:  "PROCESSED",
			Accrual: &accrualSum,
		},
		nil,
	).AnyTimes()

	processOrderCtx, processOrderCtxCancel := context.WithCancel(context.Background())

	appUsecase := &AppUsecase{
		AppRepo:                      mockARI,
		AccrualSystemClient:          mockASCI,
		passwordKey:                  "",
		tokenKey:                     "",
		tokenExp:                     0,
		processOrdersChan:            make(chan *app.Order, 1),
		processOrdersTicker:          time.NewTicker(time.Millisecond),
		updateExistedNewOrdersTicker: time.NewTicker(time.Millisecond),
		processOrdersCtx:             processOrderCtx,
		processOrdersCtxCancel:       processOrderCtxCancel,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appUsecase.processOrders(context.Background(), tt.args.orders)
			if tt.want.orders == nil {
				assert.Len(t, actualOrders, 0)
				return
			}
			assert.Equal(t, tt.want.orders, actualOrders)
			actualOrders = []*app.Order{}
		})
	}

	appUsecase.Close()
}

func TestAppUsecase_worker(t *testing.T) {
	orderRegisteredNumber := "1"
	orderInvalidNumber := "2"
	orderProcessingNumber := "3"
	orderProcessedNumber := "4"

	accrualSum := float64(100)

	uploadedAt := time.Now()

	type args struct {
		order *app.Order
	}

	type want struct {
		order *app.Order
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "registered order",
			args: args{
				order: &app.Order{
					ID:         1,
					UserID:     1,
					Number:     orderRegisteredNumber,
					Status:     "NEW",
					Accrual:    nil,
					UploadedAt: uploadedAt,
				},
			},
			want: want{
				order: nil,
			},
		},
		{
			name: "invalid order",
			args: args{
				order: &app.Order{
					ID:         2,
					UserID:     2,
					Number:     orderInvalidNumber,
					Status:     "NEW",
					Accrual:    nil,
					UploadedAt: uploadedAt,
				},
			},
			want: want{
				order: &app.Order{
					ID:         2,
					UserID:     2,
					Number:     orderInvalidNumber,
					Status:     "INVALID",
					Accrual:    nil,
					UploadedAt: uploadedAt,
				},
			},
		},
		{
			name: "processing order",
			args: args{
				order: &app.Order{
					ID:         3,
					UserID:     3,
					Number:     orderProcessingNumber,
					Status:     "NEW",
					Accrual:    nil,
					UploadedAt: uploadedAt,
				},
			},
			want: want{
				order: &app.Order{
					ID:         3,
					UserID:     3,
					Number:     orderProcessingNumber,
					Status:     "PROCESSING",
					Accrual:    nil,
					UploadedAt: uploadedAt,
				},
			},
		},
		{
			name: "processed order",
			args: args{
				order: &app.Order{
					ID:         4,
					UserID:     4,
					Number:     orderProcessedNumber,
					Status:     "NEW",
					Accrual:    nil,
					UploadedAt: uploadedAt,
				},
			},
			want: want{
				order: &app.Order{
					ID:         4,
					UserID:     4,
					Number:     orderProcessedNumber,
					Status:     "PROCESSED",
					Accrual:    &accrualSum,
					UploadedAt: uploadedAt,
				},
			},
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	mockARI := mocks.NewMockAppRepoInterface(ctrl)

	testOrderChan := make(chan *app.Order, 1)
	mockARI.EXPECT().UpdateOrder(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, order *app.Order) error {
			testOrderChan <- order
			return nil
		},
	).AnyTimes()

	mockASCI := mocks.NewMockAccrualSystemClientInterface(ctrl)

	mockASCI.EXPECT().GetOrderInfo(
		gomock.Any(),
		orderRegisteredNumber,
	).Return(
		accrual.OrderInfo{
			Number:  orderRegisteredNumber,
			Status:  "REGISTERED",
			Accrual: nil,
		},
		nil,
	).AnyTimes()

	mockASCI.EXPECT().GetOrderInfo(
		gomock.Any(),
		orderInvalidNumber,
	).Return(
		accrual.OrderInfo{
			Number:  orderInvalidNumber,
			Status:  "INVALID",
			Accrual: nil,
		},
		nil,
	).AnyTimes()

	mockASCI.EXPECT().GetOrderInfo(
		gomock.Any(),
		orderProcessingNumber,
	).Return(
		accrual.OrderInfo{
			Number:  orderProcessingNumber,
			Status:  "PROCESSING",
			Accrual: nil,
		},
		nil,
	).AnyTimes()

	mockASCI.EXPECT().GetOrderInfo(
		gomock.Any(),
		orderProcessedNumber,
	).Return(
		accrual.OrderInfo{
			Number:  orderProcessedNumber,
			Status:  "PROCESSED",
			Accrual: &accrualSum,
		},
		nil,
	).AnyTimes()

	processOrderCtx, processOrderCtxCancel := context.WithCancel(context.Background())

	appUsecase := &AppUsecase{
		AppRepo:                      mockARI,
		AccrualSystemClient:          mockASCI,
		passwordKey:                  "",
		tokenKey:                     "",
		tokenExp:                     0,
		processOrdersChan:            make(chan *app.Order, 1),
		processOrdersTicker:          time.NewTicker(time.Millisecond),
		updateExistedNewOrdersTicker: time.NewTicker(time.Millisecond),
		processOrdersCtx:             processOrderCtx,
		processOrdersCtxCancel:       processOrderCtxCancel,
	}

	go appUsecase.worker()

	ticker := time.NewTicker(time.Second)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appUsecase.processOrdersChan <- tt.args.order
			select {
			case order := <-testOrderChan:
				assert.Equal(t, tt.want.order, order)
			case <-ticker.C:
				if tt.args.order.Number == orderRegisteredNumber {
					return
				}
				assert.Fail(t, "no order in chan after timeout")
			}
		})
	}

	appUsecase.Close()
}

func TestAppUsecase_deferredWorker(t *testing.T) {
	orderNumber := "12345"
	order := &app.Order{
		ID:         1,
		UserID:     1,
		Number:     orderNumber,
		Status:     "NEW",
		Accrual:    nil,
		UploadedAt: time.Now(),
	}
	orders := []*app.Order{order}

	processedStatus := "PROCESSED"
	accrualSum := float64(100)

	expectedOrder := order
	expectedOrder.Status = processedStatus
	expectedOrder.Accrual = &accrualSum

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	mockARI := mocks.NewMockAppRepoInterface(ctrl)

	mockARI.EXPECT().GetNewOrders(gomock.Any()).Return(orders, nil).AnyTimes()

	testOrderChan := make(chan *app.Order, 1)
	mockARI.EXPECT().UpdateOrder(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, order *app.Order) error {
			testOrderChan <- order
			return nil
		},
	).AnyTimes()

	mockASCI := mocks.NewMockAccrualSystemClientInterface(ctrl)

	mockASCI.EXPECT().GetOrderInfo(
		gomock.Any(),
		orderNumber,
	).Return(
		accrual.OrderInfo{
			Number:  orderNumber,
			Status:  processedStatus,
			Accrual: &accrualSum,
		},
		nil,
	).AnyTimes()

	processOrderCtx, processOrderCtxCancel := context.WithCancel(context.Background())

	appUsecase := &AppUsecase{
		AppRepo:                      mockARI,
		AccrualSystemClient:          mockASCI,
		passwordKey:                  "",
		tokenKey:                     "",
		tokenExp:                     0,
		processOrdersChan:            make(chan *app.Order, 1),
		processOrdersTicker:          time.NewTicker(time.Nanosecond),
		updateExistedNewOrdersTicker: time.NewTicker(time.Nanosecond),
		processOrdersCtx:             processOrderCtx,
		processOrdersCtxCancel:       processOrderCtxCancel,
	}

	go appUsecase.deferredWorker()

	select {
	case actualOrder := <-testOrderChan:
		assert.Equal(t, expectedOrder, actualOrder)
	case <-time.NewTicker(time.Second).C:
		assert.Fail(t, "no order in chan after timeout")
	}

	appUsecase.Close()
}

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

	m.EXPECT().CreateUser(
		gomock.Any(),
		existedLogin,
		gomock.Any(),
	).Return(
		nil,
		&pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint \"user_login_key\""},
	).AnyTimes()

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
			u, err := appUsecase.Register(context.Background(), tt.args.login, tt.args.password)
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
	m.EXPECT().AuthUser(gomock.Any(), login, appUsecase.hashPassword(password)).Return(user, nil).AnyTimes()

	m.EXPECT().AuthUser(gomock.Any(), login, gomock.Not(appUsecase.hashPassword(password))).Return(nil, sql.ErrNoRows).AnyTimes()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := appUsecase.Login(context.Background(), tt.args.login, tt.args.password)
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
			token, errBuildJWTString := appUsecase.BuildJWTString(context.Background(), tt.args.userID)
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
	token, err := appUsecase.BuildJWTString(context.Background(), userID)
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
	token, err := appUsecase.BuildJWTString(context.Background(), userID)
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
	m.EXPECT().CreateOrder(gomock.Any(), userID, number).Return(order, nil).AnyTimes()
	m.EXPECT().CreateOrder(gomock.Any(), userID, uploadedNumber).Return(nil, sql.ErrNoRows).AnyTimes()
	m.EXPECT().CreateOrder(
		gomock.Any(),
		userID,
		uploadedByAnotherUserNumber,
	).Return(
		nil,
		&pgconn.PgError{
			Code:    "23505",
			Message: "duplicate key value violates unique constraint \"order_number_key\"",
		},
	).AnyTimes()

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
			o, err := appUsecase.CreateOrder(context.Background(), tt.args.userID, tt.args.number)
			assert.ErrorIs(t, err, tt.want.err)
			if err == nil {
				assert.Equal(t, tt.want.order, o)
				oFromChan := <-appUsecase.processOrdersChan
				assert.Equal(t, tt.want.order, oFromChan)
			}
		})
	}

	close(appUsecase.processOrdersChan)
	appUsecase.processOrdersCtxCancel()
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
	m.EXPECT().CreateWithdrawal(gomock.Any(), userID, number, sum).Return(withdrawal, nil).AnyTimes()
	m.EXPECT().CreateWithdrawal(
		gomock.Any(),
		userID,
		uploadedNumber,
		gomock.Any(),
	).Return(nil, sql.ErrNoRows).AnyTimes()
	m.EXPECT().CreateWithdrawal(gomock.Any(), userID, uploadedByAnotherUserNumber, gomock.Any()).
		Return(
			nil,
			&pgconn.PgError{
				Code:    "23505",
				Message: "duplicate key value violates unique constraint \"order_number_key\"",
			},
		).AnyTimes()
	m.EXPECT().CreateWithdrawal(gomock.Any(), userID, number, bigSum).
		Return(
			nil,
			&pgconn.PgError{
				Code:    "23514",
				Message: "new row for relation \"balance\" violates check constraint \"balance_current_check\"",
			},
		).AnyTimes()

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
			w, err := appUsecase.CreateWithdrawal(context.Background(), tt.args.userID, tt.args.number, tt.args.sum)
			assert.ErrorIs(t, err, tt.want.err)
			if err == nil {
				assert.Equal(t, tt.want.withdrawal, w)
			}
		})
	}
}

func TestLuhnAlgorithm(t *testing.T) {
	evenLenNumber := "4561261212345467"
	invalidEvenLenNumber := "4561261212345464"
	oddLenNumber := "456126121234548"
	invalidOddLenNumber := "456126121234546"
	badNumber := "bad_number"

	type args struct {
		number string
	}

	type want struct {
		ok      bool
		wantErr bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid even number",
			args: args{
				number: evenLenNumber,
			},
			want: want{
				ok:      true,
				wantErr: false,
			},
		},
		{
			name: "invalid even number",
			args: args{
				number: invalidEvenLenNumber,
			},
			want: want{
				ok:      false,
				wantErr: false,
			},
		},
		{
			name: "valid odd number",
			args: args{
				number: oddLenNumber,
			},
			want: want{
				ok:      true,
				wantErr: false,
			},
		},
		{
			name: "invalid odd number",
			args: args{
				number: invalidOddLenNumber,
			},
			want: want{
				ok:      false,
				wantErr: false,
			},
		},
		{
			name: "bad number",
			args: args{
				number: badNumber,
			},
			want: want{
				ok:      false,
				wantErr: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := luhnAlgorithm(tt.args.number)
			if tt.want.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if err == nil {
				assert.Equal(t, tt.want.ok, ok)
			}
		})
	}
}
