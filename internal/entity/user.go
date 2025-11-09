package entity

import "time"

type User struct {
	ID             int
	ContainsSub    bool
	TotalSub       int
	PromocodeID    int
	UserTelegramId int64
	FirstTime      time.Time
	UserName       string
}
