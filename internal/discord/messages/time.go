package discord_messages

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
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

func (m *Messages) timezoneOptions(l string, selected *time.Location) []discordgo.SelectMenuOption {
	timezoneSelectOptions := make([]discordgo.SelectMenuOption, 0, len(m.timezones))
	defaultTz := selected.String()
	for _, tz := range m.timezones {
		timezoneSelectOptions = append(timezoneSelectOptions, discordgo.SelectMenuOption{
			Label:   fmt.Sprintf("%s: %s", l, tz),
			Value:   tz,
			Default: defaultTz == tz,
		})
	}
	return timezoneSelectOptions
}
