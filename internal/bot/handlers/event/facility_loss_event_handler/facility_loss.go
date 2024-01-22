package facility_loss_event_handler

import (
	"context"

	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/facilities_manager"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(
	outfitLoader loaders.KeyedLoader[string, ps2.Outfit],
	facilityLoader loaders.KeyedLoader[string, ps2.Facility],
) handlers.Ps2EventHandler[facilities_manager.FacilityLoss] {
	return handlers.SimpleMessage[facilities_manager.FacilityLoss](func(
		ctx context.Context,
		cfg *handlers.Ps2EventHandlerConfig,
		event facilities_manager.FacilityLoss,
	) (string, error) {
		worldId, err := ps2.ToWorldId(event.WorldID)
		if err != nil {
			return "", err
		}
		facility, err := facilityLoader.Load(ctx, event.FacilityID)
		if err != nil {
			return "", err
		}
		outfitTag, err := outfitLoader.Load(ctx, event.OldOutfitId)
		if err != nil {
			return "", err
		}
		return render.RenderFacilityLoss(worldId, outfitTag, facility), nil
	})
}
