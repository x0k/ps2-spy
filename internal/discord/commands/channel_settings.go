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
type ChannelCharacterNotificationsSaver = func(ctx context.Context, channelId discord.ChannelId, enabled bool) error
type ChannelOutfitNotificationsSaver = func(ctx context.Context, channelId discord.ChannelId, enabled bool) error
type ChannelTitleUpdatesSaver = func(ctx context.Context, channelId discord.ChannelId, enabled bool) error

func makeFieldHandler[V any](
	messages *discord_messages.Messages,
	valueExtractor func(*discordgo.InteractionCreate) (V, error),
	saver func(ctx context.Context, channelId discord.ChannelId, value V) error,
	channelLoader ChannelLoader,
) discord.InteractionHandler {
	return discord.MessageUpdate(func(
		ctx context.Context,
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
	) discord.Response {
		value, err := valueExtractor(i)
		if err != nil {
			return messages.FieldValueExtractError(err)
		}
		channelId := discord.ChannelId(i.Interaction.ChannelID)
		if err := saver(ctx, channelId, value); err != nil {
			return discord_messages.ChannelLanguageSaveError[discordgo.InteractionResponseData](
				channelId,
				err,
			)
		}
		channel, err := channelLoader(ctx, channelId)
		if err != nil {
			return discord_messages.ChannelLoadError[discordgo.InteractionResponseData](
				channelId,
				err,
			)
		}
		return messages.ChannelSettingsFormUpdate(channel)
	})
}

func extractBool(ic *discordgo.InteractionCreate) (bool, error) {
	return ic.MessageComponentData().Values[0] == "on", nil
}

func NewChannelSettings(
	messages *discord_messages.Messages,
	channelLoader ChannelLoader,
	channelLanguageSaver ChannelLanguageSaver,
	channelCharacterNotificationsSaver ChannelCharacterNotificationsSaver,
	channelOutfitNotificationsSaver ChannelOutfitNotificationsSaver,
	channelTitleUpdatesSaver ChannelTitleUpdatesSaver,
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
				return discord_messages.ChannelLoadError[discordgo.WebhookEdit](
					channelId,
					err,
				)
			}
			return messages.ChannelSettingsForm(channel)
		}),
		ComponentHandlers: map[string]discord.InteractionHandler{
			discord.CHANNEL_LANGUAGE_COMPONENT_CUSTOM_ID: makeFieldHandler(
				messages,
				func(ic *discordgo.InteractionCreate) (language.Tag, error) {
					return language.Parse(string(ic.MessageComponentData().Values[0]))
				},
				channelLanguageSaver,
				channelLoader,
			),
			discord.CHANNEL_CHARACTER_NOTIFICATIONS_COMPONENT_CUSTOM_ID: makeFieldHandler(
				messages,
				extractBool,
				channelCharacterNotificationsSaver,
				channelLoader,
			),
			discord.CHANNEL_OUTFIT_NOTIFICATIONS_COMPONENT_CUSTOM_ID: makeFieldHandler(
				messages,
				extractBool,
				channelOutfitNotificationsSaver,
				channelLoader,
			),
			discord.CHANNEL_TITLE_UPDATES_COMPONENT_CUSTOM_ID: makeFieldHandler(
				messages,
				extractBool,
				channelTitleUpdatesSaver,
				channelLoader,
			),
		},
	}
}
