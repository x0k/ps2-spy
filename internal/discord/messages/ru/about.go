package ru_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
)

func (m *messages) About() (*discordgo.WebhookEdit, *discord.Error) {
	content := `# PlanetSide 2 Spy

Простой дискорд бот для PlanetSide 2 аутфитов

## Ссылки

- [GitHub](https://github.com/x0k/ps2-spy)
		
`
	return &discordgo.WebhookEdit{
		Content: &content,
	}, nil
}
