package models

import (
	"github.com/MukizuL/diploma-1/internal/errs"
	"time"
)

type Status int

const (
	StatusNew Status = iota
	StatusProcessing
	StatusProcessed
	StatusInvalid
)

var statusName = map[Status]string{
	StatusNew:        "NEW",
	StatusProcessing: "PROCESSING",
	StatusProcessed:  "PROCESSED",
	StatusInvalid:    "INVALID",
}

var stringStatus = map[string]Status{
	"NEW":        StatusNew,
	"PROCESSING": StatusProcessing,
	"PROCESSED":  StatusProcessed,
	"INVALID":    StatusInvalid,
}

func (ss Status) String() string {
	return statusName[ss]
}

func NewStatus(in string) (Status, error) {
	if v, ok := stringStatus[in]; !ok {
		return StatusNew, errs.ErrNoStatus
	} else {
		return v, nil
	}
}

type User struct {
	ID           string
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}

type Order struct {
	ID        string
	UserID    string
	OrderID   int64
	Status    Status
	Accrual   float64
	CreatedAt time.Time
}

type Withdrawal struct {
	ID        string
	UserID    string
	OrderID   int64
	Sum       float64
	CreatedAt time.Time
}
