package discord_commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"golang.org/x/text/language"
)

type ChannelLoader = loader.Keyed[discord.ChannelId, discord.Channel]
type ChannelLanguageSaver = func(ctx context.Context, channelId discord.ChannelId, language language.Tag) error

func NewChannelSettings(
	messages *discord_messages.Messages,
	channelLoader ChannelLoader,
	channelLanguageSaver ChannelLanguageSaver,
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
		Handler: discord.DeferredEphemeralResponse(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.ResponseEdit {
			channelId := discord.ChannelId(i.Interaction.ChannelID)
			channel, err := channelLoader(ctx, channelId)
			if err != nil {
				return messages.ChannelLoadError(channelId, err)
			}
			return messages.ChannelSettingsForm(
				discord.CHANNEL_LANGUAGE_COMPONENT_CUSTOM_ID,
				discord.CHANNEL_CHARACTER_NOTIFICATIONS_COMPONENT_CUSTOM_ID,
				discord.CHANNEL_OUTFIT_NOTIFICATIONS_COMPONENT_CUSTOM_ID,
				discord.CHANNEL_TITLE_UPDATES_COMPONENT_CUSTOM_ID,
				channel,
			)
		}),
		ComponentHandlers: map[string]discord.InteractionHandler{
			discord.CHANNEL_LANGUAGE_COMPONENT_CUSTOM_ID: discord.DeferredEphemeralFollowUp(func(
				ctx context.Context,
				s *discordgo.Session,
				i *discordgo.InteractionCreate,
			) discord.FollowUp {

			}),
		},
	}
}
