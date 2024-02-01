package logout_event_handler

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(
	charLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character],
) handlers.Ps2EventHandler[characters_tracker.PlayerLogout] {
	return handlers.SimpleMessage[characters_tracker.PlayerLogout](func(
		ctx context.Context,
		cfg *handlers.Ps2EventHandlerConfig,
		event characters_tracker.PlayerLogout,
	) (string, *handlers.Error) {
		const op = "bot.handlers.events.logout_event_handler"
		character, err := charLoader.Load(ctx, event.CharacterId)
		if err != nil {
			return "", &handlers.Error{
				Msg: "Failed to get character",
				Err: fmt.Errorf("%s getting character: %w", op, err),
			}
		}
		return render.RenderCharacterLogout(character), nil
	})
}
