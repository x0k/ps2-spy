package discord_events

import (
	"context"

	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func NewFacilityLossHandlerFactory(
	messages *discord_messages.Messages,
	outfitLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.OutfitId, ps2.Outfit],
	facilityLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.FacilityId, ps2.Facility],
) HandlerFactory {
	outfitLoader := outfitLoaders[ps2_platforms.PC]
	facilityLoader := facilityLoaders[ps2_platforms.PC]
	return func(platform ps2_platforms.Platform) Handler {
		return SimpleMessage(func(ctx context.Context, event FacilityControl) discord.Message {
			facilityId := ps2.FacilityId(event.FacilityID)
			facility, err := facilityLoader(ctx, facilityId)
			if err != nil {
				return messages.FacilityLoadError(facilityId, err)
			}
			outfitId := ps2.OutfitId(event.OutfitID)
			outfitTag, err := outfitLoader(ctx, outfitId)
			if err != nil {
				return messages.OutfitLoadError(outfitId, platform, err)
			}
			worldId := ps2.WorldId(event.WorldID)
			return messages.FacilityLoss(worldId, outfitTag, facility)
		})
	}
}
