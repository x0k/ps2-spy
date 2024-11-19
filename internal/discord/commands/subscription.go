package discord_commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func NewSubscription(
	messages discord.LocalizedMessages,
	settingsLoader loader.Keyed[discord.SettingsQuery, discord.SubscriptionSettings],
	namesLoader loader.Queried[discord.PlatformQuery[[]ps2.CharacterId], []string],
	outfitTagsLoader loader.Queried[discord.PlatformQuery[[]ps2.OutfitId], []string],
) *discord.Command {
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name: "subscription",
			NameLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "настройка",
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
			names, err := namesLoader(ctx, discord.PlatformQuery[[]ps2.CharacterId]{
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
	}
}
