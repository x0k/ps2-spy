package discord_commands

import (
	"context"
	"maps"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/stringsx"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"golang.org/x/text/language"
)

type CharacterIdsByNameLoader = func(context.Context, ps2_platforms.Platform, []string) (map[string]ps2.CharacterId, error)
type OutfitIdsByTagLoader = func(context.Context, ps2_platforms.Platform, []string) (map[string]ps2.OutfitId, error)
type ChannelTrackingSettingsSaver = func(
	ctx context.Context,
	channelId discord.ChannelId,
	platform ps2_platforms.Platform,
	settings discord.TrackingSettings,
	lang language.Tag,
) error

func NewTracking(
	messages *discord_messages.Messages,
	settingsLoader loader.Keyed[discord.SettingsQuery, discord.RichTrackingSettings],
	outfitsByTagLoader OutfitIdsByTagLoader,
	charactersByNameLoader CharacterIdsByNameLoader,
	channelTrackingSettingsSaver ChannelTrackingSettingsSaver,
) *discord.Command {
	submitHandlers := make(map[string]discord.InteractionHandler, len(ps2_platforms.Platforms))
	for _, platform := range ps2_platforms.Platforms {
		submitHandlers[discord.TRACKING_MODAL_CUSTOM_IDS[platform]] = discord.DeferredEphemeralResponse(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.ResponseEdit {
			data := i.ModalSubmitData()
			var outfitIds map[string]ps2.OutfitId
			outfitTagsFromInput := stringsx.SplitAndTrim(
				data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
				",",
			)
			if outfitTagsFromInput[0] != "" {
				outfitIds, _ = outfitsByTagLoader(ctx, platform, outfitTagsFromInput)
			}
			var characterIds map[string]ps2.CharacterId
			charNamesFromInput := stringsx.SplitAndTrim(
				data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
				",",
			)
			if charNamesFromInput[0] != "" {
				characterIds, _ = charactersByNameLoader(ctx, platform, charNamesFromInput)
			}

			var missingOutfits []string
			if len(outfitTagsFromInput) > len(outfitIds) {
				missingOutfits = make([]string, 0, len(outfitTagsFromInput)-len(outfitIds))
				for _, tag := range outfitTagsFromInput {
					if _, ok := outfitIds[tag]; !ok {
						missingOutfits = append(missingOutfits, tag)
					}
				}
			}
			var missingCharacters []string
			if len(charNamesFromInput) > len(characterIds) {
				missingCharacters = make([]string, 0, len(charNamesFromInput)-len(characterIds))
				for _, name := range charNamesFromInput {
					if _, ok := characterIds[name]; !ok {
						missingCharacters = append(missingCharacters, name)
					}
				}
			}
			if len(missingOutfits) > 0 || len(missingCharacters) > 0 {
				return messages.TrackingSettingsFailure(
					outfitTagsFromInput,
					outfitIds,
					missingOutfits,
					charNamesFromInput,
					characterIds,
					missingCharacters,
				)
			}

			channelId := discord.ChannelId(i.ChannelID)
			langTag := discord.DEFAULT_LANG_TAG
			if i.GuildLocale != nil {
				langTag = discord.LangTagFromInteraction(i)
			}
			err := channelTrackingSettingsSaver(
				ctx,
				channelId,
				platform,
				discord.TrackingSettings{
					Outfits:    slices.Collect(maps.Values(outfitIds)),
					Characters: slices.Collect(maps.Values(characterIds)),
				},
				langTag,
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
