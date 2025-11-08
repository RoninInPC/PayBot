package model

type User struct {
	Id          int64
	Username    *string
	FirstName   string
	TotalSub    int
	ContainsSub int
	PromocodeID int64
}
