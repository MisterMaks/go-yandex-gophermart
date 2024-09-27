package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app"
	loggerInternal "github.com/MisterMaks/go-yandex-gophermart/internal/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

type UserIDKeyType string

const (
	ContentTypeKey                   = "Content-Type"
	ApplicationJSONKey               = "application/json"
	TextPlainKey                     = "text/plain"
	HeaderKey                        = "header"
	RequestBodyKey                   = "request_body"
	AccessTokenKey                   = "accessToken"
	UserIDKey          UserIDKeyType = "user_id"
	AuthorizationKey                 = "Authorization"
	BearerKey                        = "Bearer "
)

func getContextUserID(ctx context.Context) (uint, error) {
	if ctx == nil {
		return 0, fmt.Errorf("no context")
	}
	userID, ok := ctx.Value(UserIDKey).(uint)
	if !ok {
		return 0, fmt.Errorf("no %v", UserIDKey)
	}
	return userID, nil
}

type AppUsecaseInterface interface {
	Register(ctx context.Context, login, password string) (*app.User, error)
	Login(ctx context.Context, login, password string) (*app.User, error)
	BuildJWTString(ctx context.Context, userID uint) (string, error)
	GetUserID(tokenString string) (uint, error)
	CreateOrder(ctx context.Context, userID uint, number string) (*app.Order, error)
	GetOrders(ctx context.Context, userID uint) ([]*app.Order, error)
	GetBalance(ctx context.Context, userID uint) (*app.Balance, error)
	CreateWithdrawal(ctx context.Context, userID uint, orderNumber string, sum float64) (*app.Withdrawal, error)
	GetWithdrawals(ctx context.Context, userID uint) ([]*app.Withdrawal, error)
}

type AppHandler struct {
	AppUsecase AppUsecaseInterface
}

func NewAppHandler(appUsecase AppUsecaseInterface) *AppHandler {
	return &AppHandler{AppUsecase: appUsecase}
}

func (ah *AppHandler) Register(w http.ResponseWriter, r *http.Request) {
	logger := loggerInternal.GetContextLogger(r.Context())

	logger.Info("Registration")

	contentType := r.Header.Get(ContentTypeKey)
	if !strings.Contains(contentType, ApplicationJSONKey) {
		logger.Warn("Request header \"Content-Type\" does not contain \"application/json\"",
			zap.Any(HeaderKey, r.Header),
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Header '%s' is not contain '%s'", ContentTypeKey, ApplicationJSONKey)))
		return
	}

	type Request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var req Request
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&req)
	if err != nil {
		logger.Warn("Failed to decode request body",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Request body could not be deserialized"))
		return
	}

	user, err := ah.AppUsecase.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		logger.Warn("Invalid login/password",
			zap.Error(err),
		)
		switch err {
		case app.ErrLoginTaken:
			w.WriteHeader(http.StatusConflict)
		case app.ErrInvalidLoginPasswordFormat:
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}

	accessToken, err := ah.AppUsecase.BuildJWTString(r.Context(), user.ID)
	if err != nil {
		logger.Error("Failed to build JWT",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set(AuthorizationKey, BearerKey+accessToken)

	http.SetCookie(w, &http.Cookie{Name: AccessTokenKey, Value: accessToken, Path: "/"})
}

func (ah *AppHandler) Login(w http.ResponseWriter, r *http.Request) {
	logger := loggerInternal.GetContextLogger(r.Context())

	logger.Info("Login")

	contentType := r.Header.Get(ContentTypeKey)
	if !strings.Contains(contentType, ApplicationJSONKey) {
		logger.Warn("Request header \"Content-Type\" does not contain \"application/json\"",
			zap.Any(HeaderKey, r.Header),
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Header '%s' is not contain '%s'", ContentTypeKey, ApplicationJSONKey)))
		return
	}

	type Request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var req Request
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&req)
	if err != nil {
		logger.Warn("Failed to decode request body",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Request body could not be deserialized"))
		return
	}

	user, err := ah.AppUsecase.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		logger.Warn("Invalid login/password",
			zap.Error(err),
		)
		switch err {
		case app.ErrInvalidLoginPassword:
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}

	accessToken, err := ah.AppUsecase.BuildJWTString(r.Context(), user.ID)
	if err != nil {
		logger.Error("Failed to build JWT",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set(AuthorizationKey, BearerKey+accessToken)

	http.SetCookie(w, &http.Cookie{Name: AccessTokenKey, Value: accessToken, Path: "/"})
}

func (ah *AppHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	logger := loggerInternal.GetContextLogger(r.Context())

	logger.Info("Create order")

	contentType := r.Header.Get(ContentTypeKey)
	if !strings.Contains(contentType, TextPlainKey) {
		logger.Warn("Request header \"Content-Type\" does not contain \"text/plain\"",
			zap.Any(HeaderKey, r.Header),
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Header '%s' is not contain '%s'", ContentTypeKey, TextPlainKey)))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to read request body",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bodyStr := string(body)

	userID, err := getContextUserID(r.Context())
	if err != nil {
		logger.Warn("No user ID",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err = ah.AppUsecase.CreateOrder(r.Context(), userID, bodyStr)
	if err != nil {
		logger.Warn("Failed to create new order",
			zap.Error(err),
		)
		switch err {
		case app.ErrOrderUploaded:
			w.WriteHeader(http.StatusOK)
		case app.ErrOrderUploadedByAnotherUser:
			w.WriteHeader(http.StatusConflict)
		case app.ErrInvalidOrderNumber:
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (ah *AppHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	logger := loggerInternal.GetContextLogger(r.Context())

	logger.Info("Getting orders")

	userID, err := getContextUserID(r.Context())
	if err != nil {
		logger.Warn("No user ID",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orders, err := ah.AppUsecase.GetOrders(r.Context(), userID)
	if err != nil {
		logger.Warn("Failed to get orders",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if len(orders) == 0 {
		logger.Warn("No orders")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set(ContentTypeKey, ApplicationJSONKey)

	enc := json.NewEncoder(w)
	err = enc.Encode(orders)
	if err != nil {
		logger.Error("Failed to encode orders",
			zap.Error(err),
		)
	}
}

func (ah *AppHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	logger := loggerInternal.GetContextLogger(r.Context())

	logger.Info("Getting balance")

	userID, err := getContextUserID(r.Context())
	if err != nil {
		logger.Warn("No user ID",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	balance, err := ah.AppUsecase.GetBalance(r.Context(), userID)
	if err != nil {
		logger.Warn("Failed to get balance",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set(ContentTypeKey, ApplicationJSONKey)

	enc := json.NewEncoder(w)
	err = enc.Encode(balance)
	if err != nil {
		logger.Error("Failed to encode balance",
			zap.Error(err),
		)
	}
}

func (ah *AppHandler) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	logger := loggerInternal.GetContextLogger(r.Context())

	logger.Info("Getting balance")

	userID, err := getContextUserID(r.Context())
	if err != nil {
		logger.Warn("No user ID",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	contentType := r.Header.Get(ContentTypeKey)
	if !strings.Contains(contentType, ApplicationJSONKey) {
		logger.Warn("Request header \"Content-Type\" does not contain \"application/json\"",
			zap.Any(HeaderKey, r.Header),
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Header '%s' is not contain '%s'", ContentTypeKey, ApplicationJSONKey)))
		return
	}

	type Request struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}

	var req Request
	dec := json.NewDecoder(r.Body)
	err = dec.Decode(&req)
	if err != nil {
		logger.Warn("Failed to decode request body",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Request body could not be deserialized"))
		return
	}

	_, err = ah.AppUsecase.CreateWithdrawal(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		logger.Warn("Failed to create withdraw",
			zap.Error(err),
		)
		switch err {
		case app.ErrInsufficientFunds:
			w.WriteHeader(http.StatusPaymentRequired)
		case app.ErrInvalidOrderNumber:
			w.WriteHeader(http.StatusUnprocessableEntity)
		case app.ErrOrderUploaded:
			w.WriteHeader(http.StatusOK)
		case app.ErrOrderUploadedByAnotherUser:
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ah *AppHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	logger := loggerInternal.GetContextLogger(r.Context())

	logger.Info("Getting withdrawals")

	userID, err := getContextUserID(r.Context())
	if err != nil {
		logger.Warn("No user ID",
			zap.Any(RequestBodyKey, r.Body),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	withdrawals, err := ah.AppUsecase.GetWithdrawals(r.Context(), userID)
	if err != nil {
		logger.Warn("Failed to get withdrawals",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if len(withdrawals) == 0 {
		logger.Warn("No withdrawals")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set(ContentTypeKey, ApplicationJSONKey)

	enc := json.NewEncoder(w)
	err = enc.Encode(withdrawals)
	if err != nil {
		logger.Error("Failed to encode withdrawals",
			zap.Error(err),
		)
	}

	logger.Debug("Response body", zap.Any("response_body", withdrawals))
}
