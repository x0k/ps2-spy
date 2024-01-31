package facility_loss_event_handler

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func New(
	outfitLoader loaders.KeyedLoader[ps2.OutfitId, ps2.Outfit],
	facilityLoader loaders.KeyedLoader[ps2.FacilityId, ps2.Facility],
) handlers.Ps2EventHandler[worlds_tracker.FacilityLoss] {
	return handlers.SimpleMessage[worlds_tracker.FacilityLoss](func(
		ctx context.Context,
		cfg *handlers.Ps2EventHandlerConfig,
		event worlds_tracker.FacilityLoss,
	) (string, *handlers.Error) {
		const op = "bot.handlers.event.facility_loss_event_handler"
		worldId := ps2.WorldId(event.WorldID)
		facility, err := facilityLoader.Load(ctx, ps2.FacilityId(event.FacilityID))
		if err != nil {
			return "", &handlers.Error{
				Msg: "Failed to get facility data",
				Err: fmt.Errorf("%s getting facility: %w", op, err),
			}
		}
		outfitTag, err := outfitLoader.Load(ctx, event.OldOutfitId)
		if err != nil {
			return "", &handlers.Error{
				Msg: "Failed to get outfit data",
				Err: fmt.Errorf("%s getting outfit: %w", op, err),
			}
		}
		return render.RenderFacilityLoss(worldId, outfitTag, facility), nil
	})
}
