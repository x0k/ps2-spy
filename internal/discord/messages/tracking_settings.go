package discord_messages

import (
	"strings"

	"github.com/x0k/ps2-spy/internal/discord"
	"golang.org/x/text/message"
)

func renderTrackingSettings(p *message.Printer, newSettings discord.TrackableEntities[[]string, []string]) string {
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

func renderTrackingSettingsFailureText(
	p *message.Printer,
	missingOutfitTags []string,
	missingCharacterNames []string,
) string {
	b := strings.Builder{}
	b.WriteString(p.Sprintf(
		"Something went wrong during the identification of the entities. We couldn't find the following:",
	))
	if len(missingOutfitTags) > 0 {
		b.WriteString(p.Sprintf("\noutfits: %s", missingOutfitTags[0]))
		for i := 1; i < len(missingOutfitTags); i++ {
			b.WriteString(", ")
			b.WriteString(missingOutfitTags[i])
		}
	}
	if len(missingCharacterNames) > 0 {
		b.WriteString(p.Sprintf("\ncharacters: %s", missingCharacterNames[0]))
		for i := 1; i < len(missingCharacterNames); i++ {
			b.WriteString(", ")
			b.WriteString(missingCharacterNames[i])
		}
	}
	b.WriteString(p.Sprintf("\n\nSo you can:"))
	return b.String()
}
