package delivery

import (
	"bytes"
	"context"
	"fmt"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app/delivery/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	TestHost string = "http://example.com"
)

func TestAppHandler_Register(t *testing.T) {
	login := "login"
	password := "password"
	takenLogin := "taken_login"
	invalidLogin := "invalid_login"
	invalidPassword := "invalid_password"
	jwtString := "jwt_string"

	user := &app.User{ID: 1, Login: login, PasswordHash: "password_hash"}

	type request struct {
		contentType string
		body        []byte
	}

	type want struct {
		statusCode int
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "valid data",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(`{"login": "` + login + `", "password": "` + password + `"}`),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "invalid " + ContentTypeKey,
			request: request{
				contentType: "invalid " + ContentTypeKey,
				body:        []byte(`{"login": "` + login + `", "password": "` + password + `"}`),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "login already taken",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(`{"login": "` + takenLogin + `", "password": "` + password + `"}`),
			},
			want: want{
				statusCode: http.StatusConflict,
			},
		},
		{
			name: "invalid login",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(`{"login": "` + invalidLogin + `", "password": "` + password + `"}`),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid password",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(`{"login": "` + login + `", "password": "` + invalidPassword + `"}`),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid request body",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(`{"invalid_key": "invalid_value"}`),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid request body 2",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(`invalid request body`),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppUsecaseInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().Register(gomock.Any(), login, password).Return(user, nil).AnyTimes()
	m.EXPECT().Register(gomock.Any(), takenLogin, password).Return(nil, app.ErrLoginTaken).AnyTimes()
	m.EXPECT().Register(gomock.Any(), invalidLogin, password).Return(nil, app.ErrInvalidLoginPasswordFormat).AnyTimes()
	m.EXPECT().Register(gomock.Any(), login, invalidPassword).Return(nil, app.ErrInvalidLoginPasswordFormat).AnyTimes()
	m.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, app.ErrInvalidLoginPasswordFormat).AnyTimes()

	m.EXPECT().BuildJWTString(gomock.Any(), user.ID).Return(jwtString, nil).AnyTimes()

	appHandler := NewAppHandler(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyReader := bytes.NewReader(tt.request.body)
			req := httptest.NewRequest(http.MethodPost, TestHost+"/api/user/register", bodyReader)
			req.Header.Add(ContentTypeKey, tt.request.contentType)

			w := httptest.NewRecorder()

			appHandler.Register(w, req)

			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode, "Invalid status code")

			if res.StatusCode == http.StatusOK {
				accessTokenOk := false
				cookies := res.Cookies()
				for _, cookie := range cookies {
					if cookie.Name == AccessTokenKey && cookie.Value == jwtString && cookie.Path == "/" {
						accessTokenOk = true
					}
				}

				assert.True(t, accessTokenOk, "Invalid cookies")
			}

			res.Body.Close()
		})
	}
}

func TestAppHandler_Login(t *testing.T) {
	login := "login"
	password := "password"
	invalidLogin := "invalid_login"
	invalidPassword := "invalid_password"
	jwtString := "jwt_string"

	user := &app.User{ID: 1, Login: login, PasswordHash: "password_hash"}

	type request struct {
		contentType string
		body        []byte
	}

	type want struct {
		statusCode int
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "valid data",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(`{"login": "` + login + `", "password": "` + password + `"}`),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "invalid " + ContentTypeKey,
			request: request{
				contentType: "invalid " + ContentTypeKey,
				body:        []byte(`{"login": "` + login + `", "password": "` + password + `"}`),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid login",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(`{"login": "` + invalidLogin + `", "password": "` + password + `"}`),
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "invalid password",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(`{"login": "` + login + `", "password": "` + invalidPassword + `"}`),
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "invalid request body",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(`{"invalid_key": "invalid_value"}`),
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "invalid request body 2",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(`invalid request body`),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppUsecaseInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().Login(gomock.Any(), login, password).Return(user, nil).AnyTimes()
	m.EXPECT().Login(gomock.Any(), invalidLogin, password).Return(nil, app.ErrInvalidLoginPassword).AnyTimes()
	m.EXPECT().Login(gomock.Any(), login, invalidPassword).Return(nil, app.ErrInvalidLoginPassword).AnyTimes()
	m.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, app.ErrInvalidLoginPassword).AnyTimes()

	m.EXPECT().BuildJWTString(gomock.Any(), user.ID).Return(jwtString, nil).AnyTimes()

	appHandler := NewAppHandler(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyReader := bytes.NewReader(tt.request.body)
			req := httptest.NewRequest(http.MethodPost, TestHost+"/api/user/login", bodyReader)
			req.Header.Add(ContentTypeKey, tt.request.contentType)

			w := httptest.NewRecorder()

			appHandler.Login(w, req)

			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode, "Invalid status code")

			if res.StatusCode == http.StatusOK {
				accessTokenOk := false
				cookies := res.Cookies()
				for _, cookie := range cookies {
					if cookie.Name == AccessTokenKey && cookie.Value == jwtString && cookie.Path == "/" {
						accessTokenOk = true
					}
				}

				assert.True(t, accessTokenOk, "Invalid cookies")
			}

			res.Body.Close()
		})
	}
}

func TestAppHandler_CreateOrder(t *testing.T) {
	userID := uint(1)
	orderNumber := `12345`
	existedOrderNumber := `11111`
	someoneElsesOrderNumber := `22222`
	invalidOrderNumberFormat := `aaaaa`

	accrual := float64(100)

	order := &app.Order{
		ID:         1,
		UserID:     userID,
		Number:     orderNumber,
		Status:     "NEW",
		Accrual:    &accrual,
		UploadedAt: time.Now(),
	}

	type request struct {
		contentType string
		body        []byte
		ctx         context.Context
	}

	type want struct {
		statusCode int
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "valid data (new order)",
			request: request{
				contentType: TextPlainKey,
				body:        []byte(orderNumber),
				ctx:         context.WithValue(context.Background(), UserIDKey, userID),
			},
			want: want{
				statusCode: http.StatusCreated,
			},
		},
		{
			name: "valid data (existed order)",
			request: request{
				contentType: TextPlainKey,
				body:        []byte(existedOrderNumber),
				ctx:         context.WithValue(context.Background(), UserIDKey, userID),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "invalid " + ContentTypeKey,
			request: request{
				contentType: "invalid " + ContentTypeKey,
				body:        []byte(orderNumber),
				ctx:         context.WithValue(context.Background(), UserIDKey, userID),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "unauthorized",
			request: request{
				contentType: TextPlainKey,
				body:        []byte(orderNumber),
				ctx:         context.Background(),
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "someone else's order",
			request: request{
				contentType: TextPlainKey,
				body:        []byte(someoneElsesOrderNumber),
				ctx:         context.WithValue(context.Background(), UserIDKey, userID),
			},
			want: want{
				statusCode: http.StatusConflict,
			},
		},
		{
			name: "invalid order number format",
			request: request{
				contentType: TextPlainKey,
				body:        []byte(invalidOrderNumberFormat),
				ctx:         context.WithValue(context.Background(), UserIDKey, userID),
			},
			want: want{
				statusCode: http.StatusUnprocessableEntity,
			},
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppUsecaseInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().CreateOrder(gomock.Any(), userID, orderNumber).Return(order, nil).AnyTimes()
	m.EXPECT().CreateOrder(gomock.Any(), userID, existedOrderNumber).Return(nil, app.ErrOrderUploaded).AnyTimes()
	m.EXPECT().CreateOrder(gomock.Any(), userID, someoneElsesOrderNumber).Return(nil, app.ErrOrderUploadedByAnotherUser).AnyTimes()
	m.EXPECT().CreateOrder(gomock.Any(), userID, invalidOrderNumberFormat).Return(nil, app.ErrInvalidOrderNumber).AnyTimes()

	appHandler := NewAppHandler(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyReader := bytes.NewReader(tt.request.body)
			req := httptest.NewRequest(http.MethodPost, TestHost+"/api/user/orders", bodyReader)
			req.Header.Add(ContentTypeKey, tt.request.contentType)
			req = req.WithContext(tt.request.ctx)

			w := httptest.NewRecorder()

			appHandler.CreateOrder(w, req)

			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode, "Invalid status code")

			res.Body.Close()
		})
	}
}

func TestAppHandler_GetOrders(t *testing.T) {
	userID := uint(1)
	userIDWithoutOrders := uint(2)

	accrual := float64(100)

	order := &app.Order{
		ID:         1,
		UserID:     userID,
		Number:     `12345`,
		Status:     "NEW",
		Accrual:    &accrual,
		UploadedAt: time.Now(),
	}
	orderWithoutAccrual := &app.Order{
		ID:         2,
		UserID:     userID,
		Number:     `67890`,
		Status:     "NEW",
		Accrual:    nil,
		UploadedAt: time.Now(),
	}
	ordersStr := fmt.Sprintf(`[
	{
		"number": "%s", 
		"status": "%s", 
		"accrual": %f, 
		"uploaded_at": "%s"
	},
	{
		"number": "%s", 
		"status": "%s",
		"uploaded_at": "%s"
	}
]`,
		order.Number,
		order.Status,
		*order.Accrual,
		order.UploadedAt.Format(time.RFC3339),
		orderWithoutAccrual.Number,
		orderWithoutAccrual.Status,
		orderWithoutAccrual.UploadedAt.Format(time.RFC3339),
	)

	type request struct {
		ctx context.Context
	}

	type want struct {
		statusCode      int
		contentType     string
		responseBodyStr string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "user with orders",
			request: request{
				ctx: context.WithValue(context.Background(), UserIDKey, userID),
			},
			want: want{
				statusCode:      http.StatusOK,
				contentType:     ApplicationJSONKey,
				responseBodyStr: ordersStr,
			},
		},
		{
			name: "user without orders",
			request: request{
				ctx: context.WithValue(context.Background(), UserIDKey, userIDWithoutOrders),
			},
			want: want{
				statusCode:      http.StatusNoContent,
				contentType:     "",
				responseBodyStr: "",
			},
		},
		{
			name: "unauthorized",
			request: request{
				ctx: context.Background(),
			},
			want: want{
				statusCode:      http.StatusUnauthorized,
				contentType:     "",
				responseBodyStr: "",
			},
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppUsecaseInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().GetOrders(gomock.Any(), userID).Return([]*app.Order{order, orderWithoutAccrual}, nil).AnyTimes()
	m.EXPECT().GetOrders(gomock.Any(), userIDWithoutOrders).Return([]*app.Order{}, nil).AnyTimes()

	appHandler := NewAppHandler(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyReader := bytes.NewReader(nil)
			req := httptest.NewRequest(http.MethodGet, TestHost+"/api/user/orders", bodyReader)
			req = req.WithContext(tt.request.ctx)

			w := httptest.NewRecorder()

			appHandler.GetOrders(w, req)

			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode, "Invalid status code")

			if res.StatusCode == http.StatusOK {
				assert.Contains(t, res.Header.Values(ContentTypeKey), tt.want.contentType)

				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tt.want.responseBodyStr, string(resBody))
			}
		})
	}
}

func TestAppHandler_GetBalance(t *testing.T) {
	userID := uint(1)

	balance := &app.Balance{
		ID:        1,
		UserID:    userID,
		Current:   100,
		Withdrawn: 300,
	}
	balanceStr := fmt.Sprintf(`{"current": %f, "withdrawn": %f}`,
		balance.Current,
		balance.Withdrawn,
	)

	type request struct {
		ctx context.Context
	}

	type want struct {
		statusCode      int
		contentType     string
		responseBodyStr string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "ok",
			request: request{
				ctx: context.WithValue(context.Background(), UserIDKey, userID),
			},
			want: want{
				statusCode:      http.StatusOK,
				contentType:     ApplicationJSONKey,
				responseBodyStr: balanceStr,
			},
		},
		{
			name: "unauthorized",
			request: request{
				ctx: context.Background(),
			},
			want: want{
				statusCode:      http.StatusUnauthorized,
				contentType:     "",
				responseBodyStr: "",
			},
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppUsecaseInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().GetBalance(gomock.Any(), userID).Return(balance, nil).AnyTimes()

	appHandler := NewAppHandler(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyReader := bytes.NewReader(nil)
			req := httptest.NewRequest(http.MethodGet, TestHost+"/api/user/balance", bodyReader)
			req = req.WithContext(tt.request.ctx)

			w := httptest.NewRecorder()

			appHandler.GetBalance(w, req)

			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode, "Invalid status code")

			if res.StatusCode == http.StatusOK {
				assert.Contains(t, res.Header.Values(ContentTypeKey), tt.want.contentType)

				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tt.want.responseBodyStr, string(resBody))
			}
		})
	}
}

func TestAppHandler_CreateWithdraw(t *testing.T) {
	userID := uint(1)
	userIDWithInsufficientFunds := uint(2)
	orderNumber := `12345`
	invalidOrderNumber := `11111`
	sum := float64(100)

	withdrawal := &app.Withdrawal{
		ID:          1,
		UserID:      userID,
		OrderNumber: orderNumber,
		Sum:         sum,
		ProcessedAt: time.Now(),
	}

	type request struct {
		contentType string
		body        []byte
		ctx         context.Context
	}

	type want struct {
		statusCode int
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "valid data",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(fmt.Sprintf(`{"order": "%s", "sum": %f}`, orderNumber, sum)),
				ctx:         context.WithValue(context.Background(), UserIDKey, userID),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "unauthorized",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(fmt.Sprintf(`{"order": "%s", "sum": %f}`, orderNumber, sum)),
				ctx:         context.Background(),
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "invalid order",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(fmt.Sprintf(`{"order": "%s", "sum": %f}`, invalidOrderNumber, sum)),
				ctx:         context.WithValue(context.Background(), UserIDKey, userID),
			},
			want: want{
				statusCode: http.StatusUnprocessableEntity,
			},
		},
		{
			name: "insufficient funds",
			request: request{
				contentType: ApplicationJSONKey,
				body:        []byte(fmt.Sprintf(`{"order": "%s", "sum": %f}`, orderNumber, sum)),
				ctx:         context.WithValue(context.Background(), UserIDKey, userIDWithInsufficientFunds),
			},
			want: want{
				statusCode: http.StatusPaymentRequired,
			},
		},
		{
			name: "invalid " + ContentTypeKey,
			request: request{
				contentType: "invalid " + ContentTypeKey,
				body:        []byte(fmt.Sprintf(`{"order": "%s", "sum": %f}`, orderNumber, sum)),
				ctx:         context.WithValue(context.Background(), UserIDKey, userID),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppUsecaseInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().CreateWithdrawal(gomock.Any(), userID, orderNumber, sum).Return(withdrawal, nil).AnyTimes()
	m.EXPECT().CreateWithdrawal(gomock.Any(), userIDWithInsufficientFunds, orderNumber, sum).Return(nil, app.ErrInsufficientFunds).AnyTimes()
	m.EXPECT().CreateWithdrawal(gomock.Any(), userID, invalidOrderNumber, sum).Return(nil, app.ErrInvalidOrderNumber).AnyTimes()

	appHandler := NewAppHandler(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyReader := bytes.NewReader(tt.request.body)
			req := httptest.NewRequest(http.MethodPost, TestHost+"/api/user/balance/withdraw", bodyReader)
			req.Header.Add(ContentTypeKey, tt.request.contentType)
			req = req.WithContext(tt.request.ctx)

			w := httptest.NewRecorder()

			appHandler.CreateWithdrawal(w, req)

			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode, "Invalid status code")

			res.Body.Close()
		})
	}
}

func TestAppHandler_GetWithdrawals(t *testing.T) {
	userID := uint(1)
	userIDWithoutWithdrawals := uint(2)

	withdrawal := &app.Withdrawal{
		ID:          1,
		UserID:      userID,
		OrderNumber: "12345",
		Sum:         100,
		ProcessedAt: time.Now(),
	}
	withdrawalsStr := fmt.Sprintf(`[{"order": "%s", "sum": %f, "processed_at": "%s"}]`,
		withdrawal.OrderNumber,
		withdrawal.Sum,
		withdrawal.ProcessedAt.Format(time.RFC3339),
	)

	type request struct {
		ctx context.Context
	}

	type want struct {
		statusCode      int
		contentType     string
		responseBodyStr string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "user with withdrawals",
			request: request{
				ctx: context.WithValue(context.Background(), UserIDKey, userID),
			},
			want: want{
				statusCode:      http.StatusOK,
				contentType:     ApplicationJSONKey,
				responseBodyStr: withdrawalsStr,
			},
		},
		{
			name: "user without withdrawals",
			request: request{
				ctx: context.WithValue(context.Background(), UserIDKey, userIDWithoutWithdrawals),
			},
			want: want{
				statusCode:      http.StatusNoContent,
				contentType:     "",
				responseBodyStr: "",
			},
		},
		{
			name: "unauthorized",
			request: request{
				ctx: context.Background(),
			},
			want: want{
				statusCode:      http.StatusUnauthorized,
				contentType:     "",
				responseBodyStr: "",
			},
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppUsecaseInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().GetWithdrawals(gomock.Any(), userID).Return([]*app.Withdrawal{withdrawal}, nil).AnyTimes()
	m.EXPECT().GetWithdrawals(gomock.Any(), userIDWithoutWithdrawals).Return([]*app.Withdrawal{}, nil).AnyTimes()

	appHandler := NewAppHandler(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyReader := bytes.NewReader(nil)
			req := httptest.NewRequest(http.MethodGet, TestHost+"/api/user/withdrawals", bodyReader)
			req = req.WithContext(tt.request.ctx)

			w := httptest.NewRecorder()

			appHandler.GetWithdrawals(w, req)

			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode, "Invalid status code")

			if res.StatusCode == http.StatusOK {
				assert.Contains(t, res.Header.Values(ContentTypeKey), tt.want.contentType)

				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tt.want.responseBodyStr, string(resBody))
			}
		})
	}
}
