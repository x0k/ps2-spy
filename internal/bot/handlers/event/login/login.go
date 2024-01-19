package login

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(charLoader loaders.KeyedLoader[string, ps2.Character]) handlers.Ps2EventHandler[ps2events.PlayerLogin] {
	return handlers.SimpleMessage[ps2events.PlayerLogin](func(ctx context.Context, cfg *handlers.Ps2EventHandlerConfig, event ps2events.PlayerLogin) (string, error) {
		const op = "handlers.login"
		character, err := charLoader.Load(ctx, event.CharacterID)
		if err != nil {
			return "", fmt.Errorf("%s error getting character: %w", op, err)
		}
		return render.RenderCharacterLogin(character), nil
	})
}
