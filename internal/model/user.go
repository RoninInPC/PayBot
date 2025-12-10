package model

import "time"

type User struct {
	Id          int64
	TgID        int64
	Username    *string
	FirstTime   time.Time
	TotalSub    int
	ContainsSub bool
	PromocodeID *int64
}
