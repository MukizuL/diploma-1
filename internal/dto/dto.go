package dto

import "time"

type AuthFormIn struct {
	Login    string `json:"login" binding:"required,min=3,max=255"`
	Password string `json:"password" binding:"required,max=72"`
}

type OrderOut struct {
	OrderID   string    `json:"number"`
	Status    string    `json:"status"`
	Accrual   float64   `json:"accrual"`
	CreatedAt time.Time `json:"uploaded_at"`
}

type BalanceOut struct {
	Balance   float64 `json:"balance"`
	Withdrawn float64 `json:"withdrawn"`
}

type OrderIn struct {
	OrderID string  `json:"order" binding:"required,max=18"`
	Sum     float64 `json:"sum" binding:"required"`
}

type WithdrawalOut struct {
	OrderID   string    `json:"order"`
	Sum       float64   `json:"sum"`
	CreatedAt time.Time `json:"processed_at"`
}

type AccrualResp struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
