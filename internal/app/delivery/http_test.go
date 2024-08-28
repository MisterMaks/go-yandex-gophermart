package delivery

import (
	"bytes"
	"context"
	"fmt"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app/mock"
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
	m := mock.NewMockAppUsecaseInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().Register(login, password).Return(user, nil)
	m.EXPECT().Register(takenLogin, password).Return(nil, app.ErrLoginTaken)
	m.EXPECT().Register(invalidLogin, password).Return(nil, app.ErrInvalidLoginPasswordFormat)
	m.EXPECT().Register(login, invalidPassword).Return(nil, app.ErrInvalidLoginPasswordFormat)
	m.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil, app.ErrInvalidLoginPasswordFormat)

	m.EXPECT().BuildJWTString(user.ID).Return(jwtString, nil)

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
	m := mock.NewMockAppUsecaseInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().Login(login, password).Return(user, nil)
	m.EXPECT().Login(invalidLogin, password).Return(nil, app.ErrInvalidLoginPassword)
	m.EXPECT().Login(login, invalidPassword).Return(nil, app.ErrInvalidLoginPassword)
	m.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, app.ErrInvalidLoginPassword)

	m.EXPECT().BuildJWTString(user.ID).Return(jwtString, nil)

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
		})
	}
}

func TestAppHandler_CreateOrder(t *testing.T) {
	userID := uint(1)
	newOrder := `12345`
	existedOrder := `11111`
	someoneElsesOrder := `22222`
	invalidOrderNumberFormat := `aaaaa`

	accrual := float64(100)

	order := &app.Order{
		ID:         1,
		UserID:     userID,
		Number:     newOrder,
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
				body:        []byte(newOrder),
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
				body:        []byte(existedOrder),
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
				body:        []byte(newOrder),
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
				body:        []byte(newOrder),
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
				body:        []byte(someoneElsesOrder),
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
	m := mock.NewMockAppUsecaseInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().CreateOrder(userID, newOrder).Return(order, nil)
	m.EXPECT().CreateOrder(userID, existedOrder).Return(nil, app.ErrOrderUploaded)
	m.EXPECT().CreateOrder(userID, someoneElsesOrder).Return(nil, app.ErrOrderUploadedByAnotherUser)
	m.EXPECT().CreateOrder(userID, invalidOrderNumberFormat).Return(nil, app.ErrInvalidOrderNumberFormat)

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
		})
	}
}

func TestAppHandler_GetOrders(t *testing.T) {
	userIDWithOrders := uint(1)
	userIDWithoutOrders := uint(2)

	orderNumber := `12345`
	status := "NEW"
	accrual := float64(100)
	uploadedAt := time.Now()

	order := &app.Order{
		ID:         1,
		UserID:     userIDWithOrders,
		Number:     orderNumber,
		Status:     status,
		Accrual:    &accrual,
		UploadedAt: time.Now(),
	}
	orderStr := fmt.Sprintf(`[{"number": "%s", "status": "%s", "accrual": %f, "uploaded_at": "%s"}]`,
		orderNumber,
		status,
		accrual,
		uploadedAt.Format(time.RFC3339),
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
				ctx: context.WithValue(context.Background(), UserIDKey, userIDWithOrders),
			},
			want: want{
				statusCode:      http.StatusOK,
				contentType:     ApplicationJSONKey,
				responseBodyStr: orderStr,
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
	m := mock.NewMockAppUsecaseInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().GetOrders(userIDWithOrders).Return([]*app.Order{order}, nil)
	m.EXPECT().GetOrders(userIDWithoutOrders).Return([]*app.Order{}, nil)

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
