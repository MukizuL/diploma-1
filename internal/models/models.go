package models

import (
	"database/sql/driver"
	"fmt"
	"github.com/MukizuL/diploma-1/internal/errs"
	"time"
)

type Status int

const (
	StatusNew Status = iota
	StatusProcessing
	StatusInvalid
	StatusProcessed
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

func (ss *Status) String() string {
	return statusName[*ss]
}

func NewStatus(in string) (Status, error) {
	if v, ok := stringStatus[in]; !ok {
		return StatusNew, errs.ErrNoStatus
	} else {
		return v, nil
	}
}

func (ss *Status) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		status, err := NewStatus(string(v))
		if err != nil {
			return err
		}
		*ss = status
		return nil
	case string:
		status, err := NewStatus(v)
		if err != nil {
			return err
		}
		*ss = status
		return nil
	case nil:
		// Handle NULL case if needed
		return nil
	default:
		return fmt.Errorf("unsupported type for Status: %T", value)
	}
}

func (ss Status) Value() (driver.Value, error) {
	return ss.String(), nil
}

type User struct {
	ID           string
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}

type Order struct {
	ID        int64
	UserID    string
	Status    Status
	Accrual   int64
	CreatedAt time.Time
}

type Withdrawal struct {
	ID        string
	UserID    string
	OrderID   int64
	Sum       int64
	CreatedAt time.Time
}
