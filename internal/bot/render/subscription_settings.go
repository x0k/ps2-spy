package render

import (
	"strings"

	"github.com/x0k/ps2-spy/internal/meta"
)

func RenderSubscriptionsSettingsUpdate(newSettings meta.TrackableEntities[[]string, []string]) string {
	builder := strings.Builder{}
	builder.WriteString("Settings are updated.\n\n**Outfits:**\n")
	outfits := newSettings.Outfits
	if len(outfits) == 0 {
		builder.WriteString("No outfits")
	} else {
		builder.WriteString(outfits[0])
		for i := 1; i < len(outfits); i++ {
			builder.WriteString(", ")
			builder.WriteString(outfits[i])
		}
	}
	builder.WriteString("\n\n**Characters:**\n")
	characters := newSettings.Characters
	if len(characters) == 0 {
		builder.WriteString("No characters")
	} else {
		builder.WriteString(characters[0])
		for i := 1; i < len(characters); i++ {
			builder.WriteString(", ")
			builder.WriteString(characters[i])
		}
	}
	return builder.String()
}
