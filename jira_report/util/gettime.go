package util

import (
	"time"
)

func GetTime() (string, string) {
	t := time.Now()
	newTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	ld1Time := newTime.AddDate(0, 0, -1)
	yesterday := ld1Time.Format("2006-01-02")
	nowday := newTime.Format("2006-01-02")
	return nowday, yesterday
}
