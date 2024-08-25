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
	ID        uint
	UserID    uint
	Current   float64
	Withdrawn float64
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
	ID          uint
	UserID      uint
	OrderNumber string
	Sum         float64
	ProcessedAt *time.Time
}

var (
	ErrLoginTaken                 = errors.New("login already taken")
	ErrInvalidLoginPassword       = errors.New("invalid login/password")
	ErrInvalidLoginPasswordFormat = errors.New("invalid login/password format")

	ErrOrderUploaded              = errors.New("order number has already been uploaded by this user")
	ErrOrderUploadedByAnotherUser = errors.New("order number has already been uploaded by this user")
	ErrInvalidOrderNumberFormat   = errors.New("invalid order number format")
)
