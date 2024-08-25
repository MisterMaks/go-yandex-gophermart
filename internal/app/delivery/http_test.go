package delivery

import (
	"bytes"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
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
			req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/register", bodyReader)
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
			req := httptest.NewRequest(http.MethodPost, "http://example.com/api/user/login", bodyReader)
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
