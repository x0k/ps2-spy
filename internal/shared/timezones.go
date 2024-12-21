package shared

import "time"

var Timezones = []string{
	"UTC", "America/New_York", "America/Chicago", "America/Los_Angeles",
	"Europe/London", "Europe/Paris", "Europe/Berlin", "Europe/Moscow",
	"Asia/Tokyo", "Asia/Shanghai", "Asia/Kolkata", "Asia/Dubai",
	"Australia/Sydney", "Pacific/Auckland", "America/Sao_Paulo",
	"Africa/Johannesburg", "Asia/Singapore",
}

func ShiftDate(weekday time.Weekday, t1 time.Duration, offset time.Duration) (time.Weekday, time.Duration) {
	t2 := t1 + offset
	if t2 < 0 {
		if weekday == time.Sunday {
			weekday = time.Saturday
		} else {
			weekday--
		}
		t2 += 24 * time.Hour
	} else if t2 >= 24*time.Hour {
		if weekday == time.Saturday {
			weekday = time.Sunday
		} else {
			weekday++
		}
		t2 -= 24 * time.Hour
	}
	return weekday, t2
}
