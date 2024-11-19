package messages_shared

import (
	"fmt"
	"time"
)

func RenderTime(t time.Time) string {
	return fmt.Sprintf("<t:%d:t>", t.Unix())
}

func RenderRelativeTime(t time.Time) string {
	return fmt.Sprintf("<t:%d:R>", t.Unix())
}
