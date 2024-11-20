package discord_commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/stringsx"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type ChannelSettingsSaver interface {
	Save(ctx context.Context, channelId discord.ChannelId, platform ps2_platforms.Platform, settings discord.SubscriptionSettings) error
}

func NewSubscription(
	messages discord.LocalizedMessages,
	settingsLoader loader.Keyed[discord.SettingsQuery, discord.SubscriptionSettings],
	characterNamesLoader loader.Queried[discord.PlatformQuery[[]ps2.CharacterId], []string],
	characterIdsLoader loader.Queried[discord.PlatformQuery[[]string], []ps2.CharacterId],
	outfitTagsLoader loader.Queried[discord.PlatformQuery[[]ps2.OutfitId], []string],
	outfitIdsLoader loader.Queried[discord.PlatformQuery[[]string], []ps2.OutfitId],
	channelSettingsSaver ChannelSettingsSaver,
) *discord.Command {
	submitHandlers := make(map[string]discord.InteractionHandler, len(ps2_platforms.Platforms))
	for _, platform := range ps2_platforms.Platforms {
		submitHandlers[discord.SUBSCRIPTION_MODAL_CUSTOM_IDS[platform]] = discord.DeferredEphemeralResponse(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.LocalizedEdit {
			data := i.ModalSubmitData()
			var err error
			var outfitsIds []ps2.OutfitId
			outfitTagsFromInput := stringsx.SplitAndTrim(
				data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
				",",
			)
			if outfitTagsFromInput[0] != "" {
				outfitsIds, err = outfitIdsLoader(ctx, discord.PlatformQuery[[]string]{
					Platform: platform,
					Value:    outfitTagsFromInput,
				})
				if err != nil {
					return messages.OutfitIdsLoadError(outfitTagsFromInput, platform, err)
				}
			}
			var charIds []ps2.CharacterId
			charNamesFromInput := stringsx.SplitAndTrim(
				data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
				",",
			)
			if charNamesFromInput[0] != "" {
				charIds, err = characterIdsLoader(ctx, discord.PlatformQuery[[]string]{
					Platform: platform,
					Value:    charNamesFromInput,
				})
				if err != nil {
					return messages.CharacterIdsLoadError(charNamesFromInput, platform, err)
				}
			}
			channelId := discord.ChannelId(i.ChannelID)
			err = channelSettingsSaver.Save(
				ctx,
				channelId,
				platform,
				discord.SubscriptionSettings{
					Outfits:    outfitsIds,
					Characters: charIds,
				},
			)
			if err != nil {
				return messages.SubscriptionSettingsSaveError(channelId, platform, err)
			}
			outfitTags, err := outfitTagsLoader(ctx, discord.PlatformQuery[[]ps2.OutfitId]{
				Platform: platform,
				Value:    outfitsIds,
			})
			if err != nil {
				return messages.SubscriptionSettingsOutfitTagsLoadError(outfitsIds, platform, err)
			}
			charNames, err := characterNamesLoader(ctx, discord.PlatformQuery[[]ps2.CharacterId]{
				Platform: platform,
				Value:    charIds,
			})
			if err != nil {
				return messages.SubscriptionSettingsCharacterNamesLoadError(charIds, platform, err)
			}
			return messages.SubscriptionSettingsUpdate(discord.TrackableEntities[[]string, []string]{
				Outfits:    outfitTags,
				Characters: charNames,
			})
		})
	}
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name: "subscription",
			NameLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "настроика",
			},
			Description: "Manage subscription settings for this channel",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "Управление подписками для этого канала",
			},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        string(ps2_platforms.PC),
					Description: "Subscription settings for the PC platform",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Настройки подписки для ПК",
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        string(ps2_platforms.PS4_EU),
					Description: "Subscription settings for the PS4 EU platform",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Настройки подписки для PS4 EU",
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        string(ps2_platforms.PS4_US),
					Description: "Subscription settings for the PS4 US platform",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Настройки подписки для PS4 US",
					},
				},
			},
		},
		Handler: discord.ShowModal(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.LocalizedResponse {
			platform := ps2_platforms.Platform(i.ApplicationCommandData().Options[0].Name)
			channelId := discord.ChannelId(i.ChannelID)
			settings, err := settingsLoader(ctx, discord.SettingsQuery{
				ChannelId: channelId,
				Platform:  platform,
			})
			if err != nil {
				return messages.SubscriptionSettingsLoadError(channelId, platform, err)
			}
			tags, err := outfitTagsLoader(ctx, discord.PlatformQuery[[]ps2.OutfitId]{
				Platform: platform,
				Value:    settings.Outfits,
			})
			if err != nil {
				return messages.OutfitTagsLoadError(settings.Outfits, platform, err)
			}
			names, err := characterNamesLoader(ctx, discord.PlatformQuery[[]ps2.CharacterId]{
				Platform: platform,
				Value:    settings.Characters,
			})
			if err != nil {
				return messages.CharacterNamesLoadError(settings.Characters, platform, err)
			}
			return messages.SubscriptionSettingsModal(
				discord.SUBSCRIPTION_MODAL_CUSTOM_IDS[platform],
				tags,
				names,
			)
		}),
		SubmitHandlers: submitHandlers,
	}
}
