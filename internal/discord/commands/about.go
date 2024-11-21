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
			Name: "about",
			NameLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "сведения",
			},
			Description: "About this bot",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "Общие сведения о боте",
			},
		},
		Handler: discord.DeferredEphemeralResponse(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.Edit {
			return messages.About()
		}),
	}
}
