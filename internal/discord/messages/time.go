package discord_messages

import (
	"fmt"
	"time"

	"golang.org/x/text/message"
)

func renderTime(t time.Time) string {
	return fmt.Sprintf("<t:%d:t>", t.Unix())
}

func renderRelativeTime(t time.Time) string {
	return fmt.Sprintf("<t:%d:R>", t.Unix())
}

func renderDuration(p *message.Printer, d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	minutes := p.Sprintf("%dm", m)
	if h == 0 {
		return minutes
	}
	return p.Sprintf("%dh ", h) + " " + minutes
}
