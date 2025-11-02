package entity

import "time"

type User struct {
	ID          int
	ContainsSub bool
	TotalSub    int
	PromocodeID int
	UserName    string
	UserTgID    int64
	FirstTime   time.Time
}
