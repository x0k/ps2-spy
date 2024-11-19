package en_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
)

func (m *messages) About() (*discordgo.WebhookEdit, *discord.Error) {
	content := `# PlanetSide 2 Spy

Simple discord bot for PlanetSide 2 outfits

## Links

- [GitHub](https://github.com/x0k/ps2-spy)
		
`
	return &discordgo.WebhookEdit{
		Content: &content,
	}, nil
}
