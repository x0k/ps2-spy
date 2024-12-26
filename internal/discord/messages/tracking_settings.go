package discord_messages

import (
	"strings"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/tracking"
	"golang.org/x/text/message"
)

func renderTrackingSettings(
	p *message.Printer,
	settings tracking.SettingsView,
) string {
	builder := strings.Builder{}
	builder.WriteString(p.Sprintf("**Tracked outfits:**\n"))
	outfits := settings.Outfits
	if len(outfits) == 0 {
		builder.WriteString(p.Sprintf("No outfits"))
	} else {
		builder.WriteString(outfits[0])
		for i := 1; i < len(outfits); i++ {
			builder.WriteString(", ")
			builder.WriteString(outfits[i])
		}
	}
	builder.WriteString(p.Sprintf("\n\n**Tracked characters:**\n"))
	characters := settings.Characters
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

func renderTrackingSettingsUpdate(
	p *message.Printer,
	updater discord.UserId,
	_ tracking.SettingsDiff,
) string {
	b := strings.Builder{}
	b.WriteString(p.Sprintf(
		"Settings updated by <@%s>",
		updater,
	))
	return b.String()
}

func renderTrackingMissingEntities(
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
