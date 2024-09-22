package delivery

import (
	"context"
	"github.com/MisterMaks/go-yandex-gophermart/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

func (ah *AppHandler) AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID uint
		var err error

		header := r.Header.Get(AuthorizationKey)
		value := strings.TrimPrefix(header, BearerKey)
		if value != "" {
			userID, err = ah.AppUsecase.GetUserID(value)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		cookie, err := r.Cookie(AccessTokenKey)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		value = cookie.Value
		userID, err = ah.AppUsecase.GetUserID(value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		ctxLogger := logger.GetContextLogger(r.Context())
		ctxLogger = ctxLogger.With(zap.Uint(string(UserIDKey), userID))
		ctx = context.WithValue(ctx, logger.LoggerKey, ctxLogger)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
