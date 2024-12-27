package discord_messages

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
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
	diff tracking.SettingsDiffView,
) string {
	b := strings.Builder{}
	b.WriteString(p.Sprintf(
		"Settings have been updated by <@%s>\n\n",
		updater,
	))
	if len(diff.Characters.ToAdd) > 0 {
		b.WriteString(p.Sprintf("**Added characters:**\n"))
		b.WriteString(diff.Characters.ToAdd[0])
		for i := 1; i < len(diff.Characters.ToAdd); i++ {
			b.WriteString(", ")
			b.WriteString(diff.Characters.ToAdd[i])
		}
	}
	if len(diff.Characters.ToDel) > 0 {
		b.WriteString(p.Sprintf("\n\n**Removed characters:**\n"))
		b.WriteString(diff.Characters.ToDel[0])
		for i := 1; i < len(diff.Characters.ToDel); i++ {
			b.WriteString(", ")
			b.WriteString(diff.Characters.ToDel[i])
		}
	}
	if len(diff.Outfits.ToAdd) > 0 {
		b.WriteString(p.Sprintf("\n\n**Added outfits:**\n"))
		b.WriteString(diff.Outfits.ToAdd[0])
		for i := 1; i < len(diff.Outfits.ToAdd); i++ {
			b.WriteString(", ")
			b.WriteString(diff.Outfits.ToAdd[i])
		}
	}
	if len(diff.Outfits.ToDel) > 0 {
		b.WriteString(p.Sprintf("\n\n**Removed outfits:**\n"))
		b.WriteString(diff.Outfits.ToDel[0])
		for i := 1; i < len(diff.Outfits.ToDel); i++ {
			b.WriteString(", ")
			b.WriteString(diff.Outfits.ToDel[i])
		}
	}
	return b.String()
}

func renderTrackingMissingEntities(
	p *message.Printer,
	missingOutfitTags []string,
	missingCharacterNames []string,
) string {
	b := strings.Builder{}
	b.WriteString(p.Sprintf(
		"We couldn't find the following:",
	))
	if len(missingOutfitTags) > 0 {
		b.WriteString(p.Sprintf("\n- Outfits: %s", missingOutfitTags[0]))
		for i := 1; i < len(missingOutfitTags); i++ {
			b.WriteString(", ")
			b.WriteString(missingOutfitTags[i])
		}
	}
	if len(missingCharacterNames) > 0 {
		b.WriteString(p.Sprintf("\n- Characters: %s", missingCharacterNames[0]))
		for i := 1; i < len(missingCharacterNames); i++ {
			b.WriteString(", ")
			b.WriteString(missingCharacterNames[i])
		}
	}
	return b.String()
}

func newTrackingEditButton(
	p *message.Printer,
	platform ps2_platforms.Platform,
	outfits []string,
	characters []string,
) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label: p.Sprintf("Edit"),
					CustomID: discord.NewTrackingSettingsEditButtonCustomId(
						platform, outfits, characters,
					),
				},
			},
		},
	}
}
