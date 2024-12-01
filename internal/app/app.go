package app

import (
	"encoding/json"
	"errors"
	"time"
)

type User struct {
	ID           uint
	Login        string
	PasswordHash string
}

type Balance struct {
	ID        uint    `json:"-"`
	UserID    uint    `json:"-"`
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Order struct {
	ID         uint      `json:"-"`
	UserID     uint      `json:"-"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func (o Order) MarshalJSON() ([]byte, error) {
	type OrderAlias Order

	aliasValue := struct {
		OrderAlias
		UploadedAt string `json:"uploaded_at"`
	}{
		OrderAlias: OrderAlias(o),
		UploadedAt: o.UploadedAt.Format(time.RFC3339),
	}

	return json.Marshal(aliasValue)
}

type Withdrawal struct {
	ID          uint      `json:"-"`
	UserID      uint      `json:"-"`
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (w Withdrawal) MarshalJSON() ([]byte, error) {
	type WithdrawalAlias Withdrawal

	aliasValue := struct {
		WithdrawalAlias
		ProcessedAt string `json:"processed_at"`
	}{
		WithdrawalAlias: WithdrawalAlias(w),
		ProcessedAt:     w.ProcessedAt.Format(time.RFC3339),
	}

	return json.Marshal(aliasValue)
}

var (
	ErrLoginTaken                 = errors.New("login already taken")
	ErrInvalidLoginPassword       = errors.New("invalid login/password")
	ErrInvalidLoginPasswordFormat = errors.New("invalid login/password format")

	ErrOrderUploaded              = errors.New("order number has already been uploaded by this user")
	ErrOrderUploadedByAnotherUser = errors.New("order number has already been uploaded by another user")

	ErrInsufficientFunds  = errors.New("there are insufficient funds in the account")
	ErrInvalidOrderNumber = errors.New("invalid order number")
)
