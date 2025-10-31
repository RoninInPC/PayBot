package entity

import "time"

type User struct {
	ID          int
	ContainsSub bool
	TotalSub    int
	PromocodeID int
	UserName    string
	FirstTime   time.Time
}
