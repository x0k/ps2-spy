package discord_events

import (
	"context"

	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"golang.org/x/text/message"
)

func NewLogoutHandlerFactory(
	messages *discord_messages.Messages,
	characterLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.CharacterId, ps2.Character],
) HandlerFactory {
	return func(platform ps2_platforms.Platform) Handler {
		characterLoader := characterLoaders[platform]
		_ = characterLoader
		return SimpleMessage(func(ctx context.Context, e PlayerLogout) func(*message.Printer) (string, *discord.Error) {
			char, err := characterLoader(ctx, e.CharacterId)
			if err != nil {
				return messages.CharacterLoadError(e.CharacterId, err)
			}
			return messages.CharacterLogout(char)
		})
	}
}
