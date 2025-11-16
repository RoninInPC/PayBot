package model

import "time"

type Payment struct {
	Id           int
	UserTgID     int64
	Amount       int64
	Timestamp    time.Time
	Status       string
	ReceiptPhoto []byte
}
