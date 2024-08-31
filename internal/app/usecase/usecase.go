package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/MisterMaks/go-yandex-gophermart/internal/accrual_system"
	"github.com/MisterMaks/go-yandex-gophermart/internal/app"
	"github.com/golang-jwt/jwt/v4"
	"strconv"
	"time"
)

type appRepoInterface interface {
	CreateUser(login, passwordHash string) (*app.User, error)
	AuthUser(login, passwordHash string) (*app.User, error)
	CreateOrder(userID uint, number string) (*app.Order, error)
	UpdateOrder(order *app.Order) error
	GetOrders(userID uint) ([]*app.Order, error)
	GetNewOrders() ([]*app.Order, error)
	GetBalance(userID uint) (*app.Balance, error)
	CreateWithdraw(userID uint, orderNumber string, sum float64) (*app.Withdrawal, error)
	GetWithdrawals(userID uint) ([]*app.Withdrawal, error)
}

type accrualSystemClientInterface interface {
	GetOrderInfo(number string) (accrual_system.OrderInfo, error)
}

type AppUsecase struct {
	appRepo appRepoInterface

	accrualSystemClient accrualSystemClientInterface

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
	appRepo appRepoInterface,
	accrualSystemClient accrualSystemClientInterface,
	passwordKey string,
	tokenKey string,
	tokenExp time.Duration,
	processOrderChanSize uint,
	processOrderWaitingTime time.Duration,
	updateExistedOrdersWaitingTime time.Duration,
) *AppUsecase {
	processOrderCtx, processOrderCtxCancel := context.WithCancel(context.Background())

	appUsecase := &AppUsecase{
		appRepo: appRepo,

		accrualSystemClient: accrualSystemClient,

		passwordKey: passwordKey,

		tokenKey: tokenKey,
		tokenExp: tokenExp,

		processOrdersChan:            make(chan *app.Order, processOrderChanSize),
		processOrdersTicker:          time.NewTicker(processOrderWaitingTime),
		updateExistedNewOrdersTicker: time.NewTicker(updateExistedOrdersWaitingTime),
		processOrdersCtx:             processOrderCtx,
		processOrdersCtxCancel:       processOrderCtxCancel,
	}

	go appUsecase.processOrder()

	return appUsecase
}

func (au AppUsecase) updateExistedNewOrders() {
Loop:
	for {
		select {
		case <-au.updateExistedNewOrdersTicker.C:
			orders, err := au.appRepo.GetNewOrders()
			if err != nil {
				continue Loop
			}
			for _, order := range orders {
				select {
				case au.processOrdersChan <- order:
				case <-au.processOrdersCtx.Done():
					return
				}
			}
		}
	}
}

func (au AppUsecase) processOrder() {
	// TODO: Добавить логи с контекстом
	orders := make([]*app.Order, 0, 2*len(au.processOrdersChan))
	for {
		select {
		case order := <-au.processOrdersChan:
			orders = append(orders, order)
		case <-au.processOrdersTicker.C:
			if len(orders) == 0 {
				continue
			}

		Loop:
			for _, order := range orders {
				orderInfo, err := au.accrualSystemClient.GetOrderInfo(order.Number)

				if err != nil {
					switch err {
					case accrual_system.ErrTooManyRequests:
						break Loop
					case accrual_system.ErrInternalServerError:
						break Loop
					case accrual_system.ErrOrderNotRegistered:
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
					continue Loop
				}

				err = au.appRepo.UpdateOrder(order)
				if err != nil {
					continue
				}
			}
		}
	}

}

func (au AppUsecase) hashPassword(password string) string {
	// подписываем алгоритмом HMAC, используя SHA-256
	h := hmac.New(sha256.New, []byte(au.passwordKey))
	h.Write([]byte(password))
	passwordHash := h.Sum(nil)

	passwordHashStr := fmt.Sprintf("%x", passwordHash)

	return passwordHashStr
}

func (au AppUsecase) Register(login, password string) (*app.User, error) {
	passwordHash := au.hashPassword(password)
	user, err := au.appRepo.CreateUser(login, passwordHash)
	return user, err
}

func (au AppUsecase) Login(login, password string) (*app.User, error) {
	passwordHash := au.hashPassword(password)
	user, err := au.appRepo.AuthUser(login, passwordHash)
	return user, err
}

type Claims struct {
	jwt.RegisteredClaims
	UserID uint
}

func (au AppUsecase) BuildJWTString(userID uint) (string, error) {
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

func luhnAlgorithm(number string) (bool, error) {
	size := len(number)

	coef := 0
	if size%2 == 0 {
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

func (au AppUsecase) CreateOrder(userID uint, number string) (*app.Order, error) {
	ok, err := luhnAlgorithm(number)
	if !ok || err != nil {
		return nil, app.ErrInvalidOrderNumber
	}

	order, err := au.appRepo.CreateOrder(userID, number)
	if err != nil {
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

func (au AppUsecase) GetOrders(userID uint) ([]*app.Order, error) {
	return au.appRepo.GetOrders(userID)
}

func (au AppUsecase) GetBalance(userID uint) (*app.Balance, error) {
	return au.appRepo.GetBalance(userID)
}

func (au AppUsecase) CreateWithdraw(userID uint, orderNumber string, sum float64) (*app.Withdrawal, error) {
	return au.appRepo.CreateWithdraw(userID, orderNumber, sum)
}

func (au AppUsecase) GetWithdrawals(userID uint) ([]*app.Withdrawal, error) {
	return au.appRepo.GetWithdrawals(userID)
}
