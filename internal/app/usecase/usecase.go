package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"github.com/MisterMaks/go-yandex-gophermart/internal/accrual"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app"
	loggerInternal "github.com/MisterMaks/go-yandex-gophermart/internal/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"regexp"
	"strconv"
	"time"
)

var (
	ErrEmptyPasswordKey = errors.New("empty password key")
	ErrEmptyTokenKey    = errors.New("empty token key")
)

type AppRepoInterface interface {
	CreateUser(ctx context.Context, login, passwordHash string) (*app.User, error)
	AuthUser(ctx context.Context, login, passwordHash string) (*app.User, error)
	CreateOrder(ctx context.Context, userID uint, number string) (*app.Order, error)
	UpdateOrder(ctx context.Context, order *app.Order) error
	GetOrders(ctx context.Context, userID uint) ([]*app.Order, error)
	GetNewOrders(ctx context.Context) ([]*app.Order, error)
	GetBalance(ctx context.Context, userID uint) (*app.Balance, error)
	CreateWithdrawal(ctx context.Context, userID uint, orderNumber string, sum float64) (*app.Withdrawal, error)
	GetWithdrawals(ctx context.Context, userID uint) ([]*app.Withdrawal, error)
}

type AccrualSystemClientInterface interface {
	GetOrderInfo(ctx context.Context, number string) (accrual.OrderInfo, error)
}

type AppUsecase struct {
	AppRepo AppRepoInterface

	AccrualSystemClient AccrualSystemClientInterface

	passwordKey string

	tokenKey string
	tokenExp time.Duration

	processOrdersChan            chan *app.Order
	processOrdersTicker          *time.Ticker
	updateExistedNewOrdersTicker *time.Ticker
	processOrdersCtx             context.Context
	processOrdersCtxCancel       context.CancelFunc
}

func NewAppUsecase(
	appRepo AppRepoInterface,
	accrualSystemClient AccrualSystemClientInterface,
	passwordKey string,
	tokenKey string,
	tokenExp time.Duration,
	processOrderChanSize uint,
	processOrderWaitingTime time.Duration,
	updateExistedNewOrdersWaitingTime time.Duration,
) (*AppUsecase, error) {
	if passwordKey == "" {
		return nil, ErrEmptyPasswordKey
	}
	if tokenKey == "" {
		return nil, ErrEmptyTokenKey
	}

	processOrderCtx, processOrderCtxCancel := context.WithCancel(context.Background())

	appUsecase := &AppUsecase{
		AppRepo: appRepo,

		AccrualSystemClient: accrualSystemClient,

		passwordKey: passwordKey,

		tokenKey: tokenKey,
		tokenExp: tokenExp,

		processOrdersChan:            make(chan *app.Order, processOrderChanSize),
		processOrdersTicker:          time.NewTicker(processOrderWaitingTime),
		updateExistedNewOrdersTicker: time.NewTicker(updateExistedNewOrdersWaitingTime),
		processOrdersCtx:             processOrderCtx,
		processOrdersCtxCancel:       processOrderCtxCancel,
	}

	go appUsecase.processOrder()
	go appUsecase.updateExistedNewOrders()

	return appUsecase, nil
}

func (au *AppUsecase) Close() {
	au.processOrdersCtxCancel()
	close(au.processOrdersChan)
}

func (au *AppUsecase) updateExistedNewOrders() {
	logger := loggerInternal.Log
Loop:
	for {
		select {
		case <-au.processOrdersCtx.Done():
			return
		case <-au.updateExistedNewOrdersTicker.C:
			iterationID := uuid.New().String()
			ctxLogger := logger.With(
				zap.String("update_existed_new_orders_iteration_id", iterationID),
			)
			ctx := context.WithValue(context.Background(), loggerInternal.LoggerKey, ctxLogger)

			orders, err := au.AppRepo.GetNewOrders(ctx)
			if err != nil {
				continue Loop
			}
			for _, order := range orders {
				select {
				case <-au.processOrdersCtx.Done():
					return
				case au.processOrdersChan <- order:
				}
			}
		}
	}
}

func (au *AppUsecase) processOrder() {
	logger := loggerInternal.Log

	orders := make([]*app.Order, 0, 2*len(au.processOrdersChan))
	for {
		select {
		case <-au.processOrdersCtx.Done():
			return
		case order := <-au.processOrdersChan:
			orders = append(orders, order)
		case <-au.processOrdersTicker.C:
			if len(orders) == 0 {
				continue
			}

			iterationID := uuid.New().String()
			ctxLogger := logger.With(
				zap.String("process_order_iteration_id", iterationID),
			)
			ctx := context.WithValue(context.Background(), loggerInternal.LoggerKey, ctxLogger)

		Loop:
			for _, order := range orders {
				orderInfo, err := au.AccrualSystemClient.GetOrderInfo(ctx, order.Number)

				if err != nil {
					logger.Warn("Failed to get order info", zap.Any("order", order), zap.Error(err))
					switch err {
					case accrual.ErrTooManyRequests:
						break Loop
					case accrual.ErrInternalServerError:
						break Loop
					case accrual.ErrOrderNotRegistered:
						continue Loop
					default:
						continue Loop
					}
				}

				switch orderInfo.Status {
				case "REGISTERED":
					continue Loop
				case "INVALID":
					order.Status = "INVALID"
				case "PROCESSING":
					order.Status = "PROCESSING"
				case "PROCESSED":
					order.Status = "PROCESSED"
					order.Accrual = orderInfo.Accrual
				default:
					logger.Error("Unknown order status", zap.Any("order", order), zap.Any("order_info", orderInfo))
					continue Loop
				}

				err = au.AppRepo.UpdateOrder(ctx, order)
				if err != nil {
					continue
				}
			}
			orders = orders[:0]
		}
	}

}

func (au *AppUsecase) hashPassword(password string) string {
	// подписываем алгоритмом HMAC, используя SHA-256
	h := hmac.New(sha256.New, []byte(au.passwordKey))
	h.Write([]byte(password))
	passwordHash := h.Sum(nil)

	passwordHashStr := fmt.Sprintf("%x", passwordHash)

	return passwordHashStr
}

func (au *AppUsecase) checkLogin(login string) (bool, error) {
	okInvalidSymbols, err := regexp.MatchString(`[^\w\.\-]+`, login)
	return !okInvalidSymbols, err
}

func (au *AppUsecase) checkPassword(password string) (bool, error) {
	okInvalidSymbols, err := regexp.MatchString(`[^\w\.\-]+`, password)
	return !okInvalidSymbols, err
}

func (au *AppUsecase) Register(ctx context.Context, login, password string) (*app.User, error) {
	ok, err := au.checkLogin(login)
	if !ok || err != nil {
		return nil, app.ErrInvalidLoginPasswordFormat
	}

	ok, err = au.checkPassword(password)
	if !ok || err != nil {
		return nil, app.ErrInvalidLoginPasswordFormat
	}

	passwordHash := au.hashPassword(password)
	user, err := au.AppRepo.CreateUser(ctx, login, passwordHash)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == "23505" && pgErr.Message == "duplicate key value violates unique constraint \"user_login_key\"":
			return nil, app.ErrLoginTaken
		case pgErr.Code == "23514" && pgErr.Message == "new row for relation \"user\" violates check constraint \"user_login_check\"":
			return nil, app.ErrInvalidLoginPasswordFormat
		default:
			return nil, err
		}
	}
	return user, err
}

func (au *AppUsecase) Login(ctx context.Context, login, password string) (*app.User, error) {
	ok, err := au.checkLogin(login)
	if !ok || err != nil {
		return nil, app.ErrInvalidLoginPasswordFormat
	}

	ok, err = au.checkPassword(password)
	if !ok || err != nil {
		return nil, app.ErrInvalidLoginPasswordFormat
	}

	passwordHash := au.hashPassword(password)
	user, err := au.AppRepo.AuthUser(ctx, login, passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, app.ErrInvalidLoginPassword
	}
	return user, err
}

type Claims struct {
	jwt.RegisteredClaims
	UserID uint
}

func (au *AppUsecase) BuildJWTString(ctx context.Context, userID uint) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(au.tokenExp)),
		},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(au.tokenKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func (au *AppUsecase) GetUserID(token string) (uint, error) {
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(au.tokenKey), nil
	})
	if err != nil {
		return 0, err
	}

	// возвращаем ID пользователя в читаемом виде
	return claims.UserID, nil
}

func luhnAlgorithm(number string) (bool, error) {
	size := len(number)

	coef := 0
	if size%2 != 0 {
		coef = 1
	}

	sum := 0
	for i, digitRune := range number {
		digit, err := strconv.Atoi(string(digitRune))
		if err != nil {
			return false, err
		}

		if (i+coef)%2 == 0 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	if sum%10 != 0 {
		return false, nil
	}

	return true, nil
}

func (au *AppUsecase) CreateOrder(ctx context.Context, userID uint, number string) (*app.Order, error) {
	ok, err := luhnAlgorithm(number)
	if !ok || err != nil {
		return nil, app.ErrInvalidOrderNumber
	}

	order, err := au.AppRepo.CreateOrder(ctx, userID, number)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, app.ErrOrderUploaded
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == "23505" && pgErr.Message == "duplicate key value violates unique constraint \"order_number_key\"":
				return nil, app.ErrOrderUploadedByAnotherUser
			default:
				return nil, err
			}
		}

		return nil, err
	}

	go func() {
		select {
		case au.processOrdersChan <- order:
		case <-au.processOrdersCtx.Done():
			return
		}
	}()

	return order, nil
}

func (au *AppUsecase) GetOrders(ctx context.Context, userID uint) ([]*app.Order, error) {
	return au.AppRepo.GetOrders(ctx, userID)
}

func (au *AppUsecase) GetBalance(ctx context.Context, userID uint) (*app.Balance, error) {
	return au.AppRepo.GetBalance(ctx, userID)
}

func (au *AppUsecase) CreateWithdrawal(ctx context.Context, userID uint, orderNumber string, sum float64) (*app.Withdrawal, error) {
	ok, err := luhnAlgorithm(orderNumber)
	if !ok || err != nil {
		return nil, app.ErrInvalidOrderNumber
	}

	withdrawal, err := au.AppRepo.CreateWithdrawal(ctx, userID, orderNumber, sum)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, app.ErrOrderUploaded
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == "23514" && pgErr.Message == "new row for relation \"balance\" violates check constraint \"balance_current_check\"":
			return nil, app.ErrInsufficientFunds
		case pgErr.Code == "23505" && pgErr.Message == "duplicate key value violates unique constraint \"order_number_key\"":
			return nil, app.ErrOrderUploadedByAnotherUser
		default:
			return nil, err
		}
	}

	return withdrawal, err
}

func (au *AppUsecase) GetWithdrawals(ctx context.Context, userID uint) ([]*app.Withdrawal, error) {
	return au.AppRepo.GetWithdrawals(ctx, userID)
}
