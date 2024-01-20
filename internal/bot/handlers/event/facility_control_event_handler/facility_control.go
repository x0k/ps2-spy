package facility_control_event_handler

import (
	"context"

	"github.com/x0k/ps2-spy/internal/bot/handlers"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
)

func New() handlers.Ps2EventHandler[ps2events.FacilityControl] {
	return handlers.SimpleMessage[ps2events.FacilityControl](func(
		ctx context.Context,
		cfg *handlers.Ps2EventHandlerConfig,
		event ps2events.FacilityControl,
	) (string, error) {
		// Defended base
		if event.NewFactionID == event.OldFactionID {
			return "", nil
		}
		return "", nil
	})
}
