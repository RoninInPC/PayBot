package entity

import "time"

type Subscription struct {
	Id        int64
	UserId    int
	TariffID  int
	StartDate time.Time
	EndDate   time.Time
	Status    string
}
