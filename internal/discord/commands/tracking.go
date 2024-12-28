package discord_commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/stringsx"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/tracking"
)

type TrackingSettingsLoader = func(context.Context, discord.ChannelId, ps2_platforms.Platform) (tracking.SettingsView, error)
type TrackingSettingsUpdater = func(context.Context, discord.ChannelId, ps2_platforms.Platform, tracking.SettingsView, discord.UserId) error

func NewTracking(
	messages *discord_messages.Messages,
	trackingSettingsLoader TrackingSettingsLoader,
	trackingSettingsUpdater TrackingSettingsUpdater,
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
			err := trackingSettingsUpdater(
				ctx,
				channelId,
				platform,
				tracking.SettingsView{
					Characters: charNamesFromInput,
					Outfits:    outfitTagsFromInput,
				},
				discord.MemberOrUserId(i),
			)
			if err != nil {
				return messages.TrackingSettingsUpdateFailure(platform, err)
			}
			return messages.TrackingSettingsUpdate()
		})
	}
	showEditModal := discord.ShowModal(func(
		ctx context.Context,
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
	) discord.Response {
		platform := ps2_platforms.Platform(i.ApplicationCommandData().Options[0].Name)
		channelId := discord.ChannelId(i.ChannelID)
		settings, err := trackingSettingsLoader(ctx, channelId, platform)
		if err != nil {
			return discord_messages.TrackingSettingsLoadError[*discordgo.InteractionResponseData](
				channelId, platform, err,
			)
		}
		return messages.TrackingSettingsModal(
			discord.TRACKING_MODAL_CUSTOM_IDS[platform],
			settings.Outfits,
			settings.Characters,
		)
	})
	showSettingsMessage := discord.DeferredEphemeralResponse(func(
		ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
	) discord.ResponseEdit {
		platform := ps2_platforms.Platform(i.ApplicationCommandData().Options[0].Name)
		channelId := discord.ChannelId(i.ChannelID)
		settings, err := trackingSettingsLoader(ctx, channelId, platform)
		if err != nil {
			return discord_messages.TrackingSettingsLoadError[*discordgo.WebhookEdit](
				channelId, platform, err,
			)
		}
		return messages.TrackingSettings(settings)
	})
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name:        "tracking",
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
		Handler: func(
			ctx context.Context, log *logger.Logger, s *discordgo.Session, i *discordgo.InteractionCreate,
		) error {
			if discord.IsChannelsManagerOrDM(i) {
				return showEditModal(ctx, log, s, i)
			}
			return showSettingsMessage(ctx, log, s, i)
		},
		SubmitHandlers: submitHandlers,
		ComponentHandlers: map[string]discord.InteractionHandler{
			discord.TRACKING_EDIT_BUTTON_CUSTOM_ID: discord.ShowModal(func(
				ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
			) discord.Response {
				platform, outfits, characters := discord.CustomIdToPlatformAndOutfitsAndCharacters(
					i.MessageComponentData().CustomID,
				)
				return messages.TrackingSettingsModal(
					discord.TRACKING_MODAL_CUSTOM_IDS[platform],
					outfits,
					characters,
				)
			}),
		},
	}
}
