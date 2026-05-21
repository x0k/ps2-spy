package timex

import "time"

func LocationToOffset(loc *time.Location) time.Duration {
	_, offsetInSeconds := time.Now().In(loc).Zone()
	return time.Duration(offsetInSeconds) * time.Second
}

func NormalizeDate(weekday time.Weekday, d time.Duration) (time.Weekday, time.Duration) {
	const (
		day  = 24 * time.Hour
		week = 7 * day
	)
	normalized := (time.Duration(weekday)*day + d) % week
	if normalized < 0 {
		normalized += week
	}
	return time.Weekday(normalized / day), normalized % day
}
