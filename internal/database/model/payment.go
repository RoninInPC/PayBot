package model

import "time"

type Payment struct {
	UserID       int64
	Amount       int64
	Timestamp    time.Time
	Status       string
	ReceiptPhoto []byte
}
