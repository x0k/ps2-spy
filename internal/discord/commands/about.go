package discord_commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
)

func NewAbout(
	messages *discord_messages.Messages,
) *discord.Command {
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name:        "about",
			Description: "About this bot",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "Общие сведения о боте",
			},
		},
		Handler: discord.DeferredEphemeralResponse(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.ResponseEdit {
			return messages.About()
		}),
	}
}
