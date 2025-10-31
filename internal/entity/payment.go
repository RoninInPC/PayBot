package entity

import "time"

type Payment struct {
	ID           int
	UserID       int
	Amount       int
	TimeStamp    time.Time
	Status       string
	ReceiptPhoto string
}
