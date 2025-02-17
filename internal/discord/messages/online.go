package discord_messages

import (
	"strings"

	"github.com/x0k/ps2-spy/internal/ps2"
	factions "github.com/x0k/ps2-spy/internal/ps2/factions"
	"golang.org/x/text/message"
)

func renderOnline(
	p *message.Printer,
	outfitCharacters map[ps2.OutfitId][]ps2.Character,
	characters []ps2.Character,
	outfits map[ps2.OutfitId]ps2.Outfit,
) string {
	if len(outfitCharacters) == 0 && len(characters) == 0 {
		return p.Sprintf("No characters online")
	}
	builder := strings.Builder{}
	builder.WriteString(p.Sprintf("Characters online:"))
	for outfitId, characters := range outfitCharacters {
		if len(characters) == 0 {
			continue
		}
		builder.WriteString("\n**")
		if outfit, ok := outfits[outfitId]; ok {
			builder.WriteString(outfit.Name)
			builder.WriteString(" [")
			builder.WriteString(outfit.Tag)
		} else {
			builder.WriteString(string(characters[0].OutfitId))
			builder.WriteString(" [")
			builder.WriteString(characters[0].OutfitTag)
		}
		builder.WriteString(p.Sprintf("] outfit ("))
		// TODO: Find non NSO faction
		builder.WriteString(factions.FactionNameById(characters[0].FactionId))
		builder.WriteString(", ")
		builder.WriteString(ps2.WorldNameById(characters[0].WorldId))
		builder.WriteString("):**")
		for _, char := range characters {
			builder.WriteString("\n- ")
			builder.WriteString(char.Name)
		}
	}
	if len(characters) > 0 {
		builder.WriteString(p.Sprintf("\n**Other characters:**"))
		for _, char := range characters {
			builder.WriteString("\n- ")
			if char.OutfitTag != "" {
				builder.WriteByte('[')
				builder.WriteString(char.OutfitTag)
				builder.WriteString("] ")
			}
			builder.WriteString(char.Name)
			builder.WriteString(" (")
			builder.WriteString(factions.FactionNameById(char.FactionId))
			builder.WriteString(", ")
			builder.WriteString(ps2.WorldNameById(char.WorldId))
			builder.WriteByte(')')
		}
	}
	return builder.String()
}
