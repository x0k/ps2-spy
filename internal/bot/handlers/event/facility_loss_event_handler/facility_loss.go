package facility_loss_event_handler

import (
	"context"

	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/facilities_manager"
)

func New() handlers.Ps2EventHandler[facilities_manager.FacilityLoss] {
	return handlers.SimpleMessage[facilities_manager.FacilityLoss](func(
		ctx context.Context,
		cfg *handlers.Ps2EventHandlerConfig,
		event facilities_manager.FacilityLoss,
	) (string, error) {
		// Defended base
		if event.NewFactionID == event.OldFactionID {
			return "", nil
		}
		return "", nil
	})
}
