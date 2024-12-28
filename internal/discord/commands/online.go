package discord_commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/tracking"
)

type OutfitsLoader = func(
	context.Context, ps2_platforms.Platform, []ps2.OutfitId,
) (map[ps2.OutfitId]ps2.Outfit, error)

type TrackingSettingsDataLoader = func(
	context.Context, discord.ChannelId, ps2_platforms.Platform,
) (tracking.SettingsData, error)

func NewOnline(
	messages *discord_messages.Messages,
	trackingSettingsDataLoader TrackingSettingsDataLoader,
	outfitsLoader OutfitsLoader,
) *discord.Command {
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name:        "online",
			Description: "Returns online trackable outfits members and characters",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "Возвращает участников аутфитов и других персонажей онлайн",
			},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        string(ps2_platforms.PC),
					Description: "For PC platform",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Для ПК",
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        string(ps2_platforms.PS4_EU),
					Description: "For PS4 EU platform",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Для PS4 EU",
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        string(ps2_platforms.PS4_US),
					Description: "For PS4 US platform",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Для PS4 US",
					},
				},
			},
		},
		Handler: discord.DeferredEphemeralResponse(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.ResponseEdit {
			platform := ps2_platforms.Platform(i.ApplicationCommandData().Options[0].Name)
			channelId := discord.ChannelId(i.ChannelID)
			onlineMembers, err := trackingSettingsDataLoader(ctx, channelId, platform)
			if err != nil {
				return messages.OnlineMembersLoadError(channelId, platform, err)
			}
			outfitIds := make([]ps2.OutfitId, 0, len(onlineMembers.Outfits))
			for id := range onlineMembers.Outfits {
				outfitIds = append(outfitIds, id)
			}
			outfits, err := outfitsLoader(ctx, platform, outfitIds)
			if err != nil {
				return messages.OutfitsLoadError(outfitIds, platform, err)
			}
			return messages.MembersOnline(onlineMembers.Outfits, onlineMembers.Characters, outfits)
		}),
	}
}
