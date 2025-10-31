package entity

import "time"

type Subscription struct {
	UserId    int
	TariffID  int
	StartDate time.Time
	EndDate   time.Time
	Status    string
}
