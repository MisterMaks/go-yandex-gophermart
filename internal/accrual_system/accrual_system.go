package accrual_system

import (
	"errors"
	"github.com/go-resty/resty/v2"
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

func (asc AccrualSystemClient) GetOrderInfo(number string) (OrderInfo, error) {
	var orderInfo OrderInfo

	resp, err := asc.client.R().SetResult(&orderInfo).SetPathParams(map[string]string{
		"number": number,
	}).Get("/api/orders/{number}")
	if err != nil {
		return OrderInfo{}, err
	}

	statusCode := resp.StatusCode()

	// TODO: Добавить лог с текстом ошибки

	if statusCode != 200 {
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
