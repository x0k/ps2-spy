package ru_messages

import (
	"strings"

	"github.com/x0k/ps2-spy/internal/discord"
)

func RenderSubscriptionsSettingsUpdate(newSettings discord.TrackableEntities[[]string, []string]) string {
	builder := strings.Builder{}
	builder.WriteString("Настройки обновлены.\n\n**Аутфиты:**\n")
	outfits := newSettings.Outfits
	if len(outfits) == 0 {
		builder.WriteString("Нет аутфитов")
	} else {
		builder.WriteString(outfits[0])
		for i := 1; i < len(outfits); i++ {
			builder.WriteString(", ")
			builder.WriteString(outfits[i])
		}
	}
	builder.WriteString("\n\n**Персонажи:**\n")
	characters := newSettings.Characters
	if len(characters) == 0 {
		builder.WriteString("Нет персонажей")
	} else {
		builder.WriteString(characters[0])
		for i := 1; i < len(characters); i++ {
			builder.WriteString(", ")
			builder.WriteString(characters[i])
		}
	}
	return builder.String()
}
