package timex

import "time"

func LocationToOffset(loc *time.Location) time.Duration {
	_, offsetInSeconds := time.Now().In(loc).Zone()
	return time.Duration(offsetInSeconds) * time.Second
}

func NormalizeDate(weekday time.Weekday, d time.Duration) (time.Weekday, time.Duration) {
	// TODO: Use modulo arithmetic
	for d < 0 {
		if weekday == time.Sunday {
			weekday = time.Saturday
		} else {
			weekday--
		}
		d += 24 * time.Hour
	}
	for d >= 24*time.Hour {
		if weekday == time.Saturday {
			weekday = time.Sunday
		} else {
			weekday++
		}
		d -= 24 * time.Hour
	}
	return weekday, d
}
