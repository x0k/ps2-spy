package discord_commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
)

func NewAbout(
	messages discord.LocalizedMessages,
) *discord.Command {
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name:        "about",
			Description: "About this bot",
		},
		Handler: discord.DeferredEphemeralResponse(func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, *discord.LocalizedError) {
			content := messages.About()(discord.LocaleFromInteraction(i))
			return &discordgo.WebhookEdit{
				Content: &content,
			}, nil
		}),
	}
}
