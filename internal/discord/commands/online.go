package discord_commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func NewOnline(
	messages discord.LocalizedMessages,
	onlineTrackableEntitiesLoader loader.Keyed[discord.SettingsQuery, discord.TrackableEntities[
		map[ps2.OutfitId][]ps2.Character,
		[]ps2.Character,
	]],
	outfitsLoader loader.Queried[discord.PlatformQuery[[]ps2.OutfitId], map[ps2.OutfitId]ps2.Outfit],
) *discord.Command {
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name: "online",
			NameLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "онлаин",
			},
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
		) discord.LocalizedEdit {
			platform := ps2_platforms.Platform(i.ApplicationCommandData().Options[0].Name)
			channelId := discord.ChannelId(i.ChannelID)
			onlineMembers, err := onlineTrackableEntitiesLoader(ctx, discord.SettingsQuery{
				ChannelId: channelId,
				Platform:  platform,
			})
			if err != nil {
				return messages.OnlineMembersLoadError(channelId, platform, err)
			}
			outfitIds := make([]ps2.OutfitId, 0, len(onlineMembers.Outfits))
			for id := range onlineMembers.Outfits {
				outfitIds = append(outfitIds, id)
			}
			outfits, err := outfitsLoader(ctx, discord.PlatformQuery[[]ps2.OutfitId]{
				Platform: platform,
				Value:    outfitIds,
			})
			if err != nil {
				return messages.OutfitsLoadError(outfitIds, platform, err)
			}
			return messages.MembersOnline(onlineMembers.Outfits, onlineMembers.Characters, outfits)
		}),
	}
}
