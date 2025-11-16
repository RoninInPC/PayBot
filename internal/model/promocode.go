package model

import "time"

type Promocode struct {
	Id        int64
	Code      string
	Discount  int64
	ExpiresAt time.Time
	UsedCount int
}
