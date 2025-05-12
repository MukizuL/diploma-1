package models

import "time"

type User struct {
	Id        string
	Login     string
	CreatedAt time.Time
}
