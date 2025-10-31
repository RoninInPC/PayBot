package entity

import "time"

type PromoCode struct {
	ID        int
	Code      string
	Discount  float64
	ExpiresAt time.Time
	UsedCount int
}
