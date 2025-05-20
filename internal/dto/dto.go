package dto

import "time"

type AuthForm struct {
	Login    string `json:"login" binding:"required,min=3,max=255"`
	Password string `json:"password" binding:"required,max=72"`
}

type Order struct {
	OrderID   string    `json:"number"`
	Status    string    `json:"status"`
	Accrual   int64     `json:"accrual"`
	CreatedAt time.Time `json:"uploaded_at"`
}
