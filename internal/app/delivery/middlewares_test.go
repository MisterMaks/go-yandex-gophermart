package delivery

import (
	"errors"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app/delivery/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestAppHandler_AuthMiddleware(t *testing.T) {
	token := "token"
	userID := uint(1)

	invalidToken := "invalid_token"
	errInvalidToken := errors.New("invalid token")

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppUsecaseInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().GetUserID(token).Return(userID, nil)
	m.EXPECT().GetUserID(invalidToken).Return(uint(0), errInvalidToken)

	appHandler := NewAppHandler(m)

	handler := appHandler.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := getContextUserID(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Write([]byte(strconv.FormatUint(uint64(userID), 10)))
	}))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	t.Run("valid token", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, srv.URL, nil)
		r.AddCookie(&http.Cookie{Name: AccessTokenKey, Value: token, Path: "/"})
		r.RequestURI = ""

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		require.NoError(t, err)
		require.Equal(t, []byte(strconv.FormatUint(uint64(userID), 10)), body)
	})

	t.Run("invalid token", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, srv.URL, nil)
		r.AddCookie(&http.Cookie{Name: AccessTokenKey, Value: invalidToken, Path: "/"})
		r.RequestURI = ""

		resp, err := http.DefaultClient.Do(r)

		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		resp.Body.Close()
	})

	t.Run("no token", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, srv.URL, nil)
		r.RequestURI = ""

		resp, err := http.DefaultClient.Do(r)

		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		resp.Body.Close()
	})
}
