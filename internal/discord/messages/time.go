package discord_messages

import (
	"fmt"
	"time"
)

func renderTime(t time.Time) string {
	return fmt.Sprintf("<t:%d:t>", t.Unix())
}

func renderRelativeTime(t time.Time) string {
	return fmt.Sprintf("<t:%d:R>", t.Unix())
}
