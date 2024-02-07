package online_command_handler

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func New(
	onlineTrackableEntitiesLoader loaders.KeyedLoader[meta.SettingsQuery, meta.TrackableEntities[
		map[ps2.OutfitId][]ps2.Character,
		[]ps2.Character,
	]],
	outfitsLoader loaders.QueriedLoader[meta.PlatformQuery[[]ps2.OutfitId], map[ps2.OutfitId]ps2.Outfit],
) handlers.InteractionHandler {
	return handlers.DeferredEphemeralResponse(func(
		ctx context.Context,
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
	) (*discordgo.WebhookEdit, *handlers.Error) {
		platform := platforms.Platform(i.ApplicationCommandData().Options[0].Name)
		onlineMembers, err := onlineTrackableEntitiesLoader.Load(ctx, meta.SettingsQuery{
			ChannelId: meta.ChannelId(i.ChannelID),
			Platform:  platform,
		})
		if err != nil {
			return nil, &handlers.Error{
				Msg: "Failed to get online members",
				Err: err,
			}
		}
		outfitIds := make([]ps2.OutfitId, 0, len(onlineMembers.Outfits))
		for id := range onlineMembers.Outfits {
			outfitIds = append(outfitIds, id)
		}
		outfits, err := outfitsLoader.Load(ctx, meta.PlatformQuery[[]ps2.OutfitId]{
			Platform: platform,
			Value:    outfitIds,
		})
		if err != nil {
			return nil, &handlers.Error{
				Msg: "Failed to get outfit data",
				Err: err,
			}
		}
		content := render.RenderOnline(onlineMembers.Outfits, onlineMembers.Characters, outfits)
		return &discordgo.WebhookEdit{
			Content: &content,
		}, nil
	})
}
