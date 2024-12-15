package discord_commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
)

type ChannelLoader = loader.Keyed[discord.ChannelId, discord.Channel]

func NewChannelSettings(
	messages *discord_messages.Messages,
	channelLoader ChannelLoader,
) *discord.Command {
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name: "channel-settings",
			NameLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "установки-канала",
			},
			Description: "Change current channel settings",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "Изменить текущие установки канала",
			},
		},
		Handler: discord.DeferredEphemeralEdit(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.Edit {
			channelId := discord.ChannelId(i.Interaction.ChannelID)
			channel, err := channelLoader(ctx, channelId)
			if err != nil {
				return messages.ChannelLoadError(channelId, err)
			}
			return messages.ChannelSettingsForm(
				"channel-language",
				"channel-character-notifications",
				"channel-outfit-notifications",
				"channel-title-updates",
				channel,
			)
		}),
	}
}
