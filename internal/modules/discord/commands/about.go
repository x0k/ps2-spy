package commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	discord_module "github.com/x0k/ps2-spy/internal/modules/discord"
)

func NewAbout() *discord_module.Command {
	return &discord_module.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name:        "about",
			Description: "About this bot",
		},
		Handler: discord_module.DeferredEphemeralResponse(func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, *discord_module.Error) {
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
