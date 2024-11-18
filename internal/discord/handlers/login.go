package discord_handlers

import (
	"context"

	"github.com/x0k/ps2-spy/internal/discord"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func NewLoginHandlerFactory(
	characterLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.CharacterId, ps2.Character],
) discord.HandlerFactory {
	return func(platform ps2_platforms.Platform) discord.Handler {
		characterLoader := characterLoaders[platform]
		return discord.SimpleMessage(func(ctx context.Context, e discord_events.PlayerLogin) (discord.MessageRenderer, *discord.Error) {
			char, err := characterLoader(ctx, e.CharacterId)
			if err != nil {
				return "", &discord.Error{
					Msg: "Failed to load character",
					Err: err,
				}
			}
			return render.RenderCharacterLogin(char), nil
		})
	}
}
