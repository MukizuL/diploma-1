package dto

import "time"

type AuthForm struct {
	Login    string `json:"login" binding:"required,min=3,max=255"`
	Password string `json:"password" binding:"required,max=72"`
}

type Order struct {
	OrderID   string    `json:"number"`
	Status    string    `json:"status"`
	Accrual   float64   `json:"accrual"`
	CreatedAt time.Time `json:"uploaded_at"`
}

type Balance struct {
	Balance   float64 `json:"balance"`
	Withdrawn float64 `json:"withdrawn"`
}
