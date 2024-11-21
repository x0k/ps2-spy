package discord_messages

import (
	"strings"

	"github.com/x0k/ps2-spy/internal/discord"
	"golang.org/x/text/message"
)

func renderSubscriptionsSettingsUpdate(p *message.Printer, newSettings discord.TrackableEntities[[]string, []string]) string {
	builder := strings.Builder{}
	builder.WriteString(p.Sprintf("Settings are updated.\n\n**Outfits:**\n"))
	outfits := newSettings.Outfits
	if len(outfits) == 0 {
		builder.WriteString(p.Sprintf("No outfits"))
	} else {
		builder.WriteString(outfits[0])
		for i := 1; i < len(outfits); i++ {
			builder.WriteString(", ")
			builder.WriteString(outfits[i])
		}
	}
	builder.WriteString(p.Sprintf("\n\n**Characters:**\n"))
	characters := newSettings.Characters
	if len(characters) == 0 {
		builder.WriteString(p.Sprintf("No characters"))
	} else {
		builder.WriteString(characters[0])
		for i := 1; i < len(characters); i++ {
			builder.WriteString(", ")
			builder.WriteString(characters[i])
		}
	}
	return builder.String()
}
