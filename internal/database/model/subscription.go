package model

import "time"

type Subscription struct {
	UserID    int64
	TariffID  int64
	StartDate time.Time
	EndDate   time.Time
	Status    string
}
