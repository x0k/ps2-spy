package shared

import "time"

var Timezones = []string{
	"UTC", "America/New_York", "America/Chicago", "America/Los_Angeles",
	"Europe/London", "Europe/Paris", "Europe/Berlin", "Europe/Moscow",
	"Asia/Tokyo", "Asia/Shanghai", "Asia/Kolkata", "Asia/Dubai",
	"Australia/Sydney", "Pacific/Auckland", "America/Sao_Paulo",
	"Africa/Johannesburg", "Asia/Singapore",
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
