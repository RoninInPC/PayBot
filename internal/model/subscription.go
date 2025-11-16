package model

import "time"

type Subscription struct {
	Id        int
	UserTgID  int64
	TariffID  int64
	StartDate time.Time
	EndDate   time.Time
	Status    string
}
