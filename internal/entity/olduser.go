package entity

import "time"

type OldUser struct {
	UserID       string    `json:"user_id" csv:"USER_ID"`
	Registration time.Time `json:"registration" csv:"registration"`
	Username     string    `json:"username" csv:"username"`
	Name         string    `json:"name" csv:"name"`
	RefLink      string    `json:"ref_link" csv:"ref_link"`
	Phone        string    `json:"phone" csv:"phone"`
	Email        string    `json:"email" csv:"email"`
	Comment      string    `json:"comment" csv:"comment"`
	Active       bool      `json:"active" csv:"active"`
	Plan         string    `json:"plan" csv:"plan"`
	EndDate      time.Time `json:"end_date" csv:"end_date"`
	UseTrial     bool      `json:"use_trial" csv:"use_trial"`
	PayCount     int       `json:"pay_count" csv:"pay_count"`
	PayLast      time.Time `json:"pay_last" csv:"pay_last"`
	Pay1st       time.Time `json:"pay_1st" csv:"pay_1st"`
}
