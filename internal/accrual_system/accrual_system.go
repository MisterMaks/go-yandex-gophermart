package accrual_system

import (
	"context"
	"errors"
	loggerInternal "github.com/MisterMaks/go-yandex-gophermart/internal/logger"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var (
	ErrOrderNotRegistered  = errors.New("order not registered")
	ErrTooManyRequests     = errors.New("too many requests")
	ErrInternalServerError = errors.New("internal server error")
)

type AccrualSystemClient struct {
	client *resty.Client
}

func NewAccrualSystemClient(accrualSystemAddress string, timeout time.Duration) *AccrualSystemClient {
	client := resty.New()
	client.SetBaseURL(accrualSystemAddress)
	client.SetTimeout(timeout)
	return &AccrualSystemClient{
		client: client,
	}
}

type OrderInfo struct {
	Number  string   `json:"number"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

func (asc AccrualSystemClient) GetOrderInfo(ctx context.Context, orderNumber string) (OrderInfo, error) {
	logger := loggerInternal.GetContextLogger(ctx)

	var orderInfo OrderInfo

	resp, err := asc.client.R().SetResult(&orderInfo).SetPathParams(map[string]string{
		"number": orderNumber,
	}).Get("/api/orders/{number}")
	if err != nil {
		return OrderInfo{}, err
	}

	statusCode := resp.StatusCode()

	if statusCode != http.StatusOK {
		logger.Warn("Order info not received", zap.String("order_number", orderNumber), zap.String("status", resp.Status()))
		switch statusCode {
		case 204:
			return OrderInfo{}, ErrOrderNotRegistered
		case 429:
			return OrderInfo{}, ErrTooManyRequests
		case 500:
			return OrderInfo{}, ErrInternalServerError
		}
	}

	return orderInfo, nil
}
