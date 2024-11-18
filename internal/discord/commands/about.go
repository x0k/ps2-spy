package discord_commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
)

func NewAbout() *discord.Command {
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name:        "about",
			Description: "About this bot",
		},
		Handler: discord.DeferredEphemeralResponse(func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, *discord.Error) {
			content := `# PlanetSide 2 Spy

Simple discord bot for PlanetSide 2 outfits

## Links

- [GitHub](https://github.com/x0k/ps2-spy)
		
`
			return &discordgo.WebhookEdit{
				Content: &content,
			}, nil
		}),
	}
}
