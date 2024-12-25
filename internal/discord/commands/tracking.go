package discord_commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/stringsx"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type ChannelTrackingSettingsUpdater = func(
	ctx context.Context,
	channelId discord.ChannelId,
	platform ps2_platforms.Platform,
	outfitTags []string,
	charNames []string,
) error

func NewTracking(
	messages *discord_messages.Messages,
	settingsLoader loader.Keyed[discord.SettingsQuery, discord.RichTrackingSettings],
	channelTrackingSettingsUpdater ChannelTrackingSettingsUpdater,
) *discord.Command {
	submitHandlers := make(map[string]discord.InteractionHandler, len(ps2_platforms.Platforms))
	for _, platform := range ps2_platforms.Platforms {
		submitHandlers[discord.TRACKING_MODAL_CUSTOM_IDS[platform]] = discord.DeferredEphemeralResponse(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.ResponseEdit {
			data := i.ModalSubmitData()
			outfitTagsFromInput := stringsx.SplitAndTrim(
				data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
				",",
			)
			charNamesFromInput := stringsx.SplitAndTrim(
				data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
				",",
			)

			channelId := discord.ChannelId(i.ChannelID)
			err := channelTrackingSettingsUpdater(
				ctx,
				channelId,
				platform,
				outfitTagsFromInput,
				charNamesFromInput,
			)
			if err != nil {
				return messages.TrackingSettingsSaveError(channelId, platform, err)
			}
			return messages.TrackingSettingsUpdate()
		})
	}
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name: "tracking",
			NameLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "отслеживание",
			},
			Description: "Manage tracking settings for this channel",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "Управление отслеживанием в этом канале",
			},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        string(ps2_platforms.PC),
					Description: "Tracking settings for the PC platform",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Настройки отслеживания для ПК",
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        string(ps2_platforms.PS4_EU),
					Description: "Tracking settings for the PS4 EU platform",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Настройки отслеживания для PS4 EU",
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        string(ps2_platforms.PS4_US),
					Description: "Tracking settings for the PS4 US platform",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Настройки отслеживания для PS4 US",
					},
				},
			},
		},
		Handler: discord.ShowModal(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.Response {
			// TODO: Show current tracking settings for unprivileged users
			if !discord.IsChannelsManagerOrDM(i) {
				return discord_messages.MissingPermissionError[discordgo.InteractionResponseData]()
			}
			platform := ps2_platforms.Platform(i.ApplicationCommandData().Options[0].Name)
			channelId := discord.ChannelId(i.ChannelID)
			settings, err := settingsLoader(ctx, discord.SettingsQuery{
				ChannelId: channelId,
				Platform:  platform,
			})
			if err != nil {
				return messages.TrackingSettingsLoadError(channelId, platform, err)
			}
			tags := make([]string, 0, len(settings.Outfits))
			for _, outfit := range settings.Outfits {
				tags = append(tags, outfit.Tag)
			}
			names := make([]string, 0, len(settings.Characters))
			for _, character := range settings.Characters {
				names = append(names, character.Name)
			}
			return messages.TrackingSettingsModal(
				discord.TRACKING_MODAL_CUSTOM_IDS[platform],
				tags,
				names,
			)
		}),
		SubmitHandlers: submitHandlers,
	}
}
