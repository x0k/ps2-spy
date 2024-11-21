package discord_events

import (
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func NewHandlers(
	messages *discord_messages.Messages,
	characterLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.CharacterId, ps2.Character],
) map[EventType]HandlerFactory {
	return map[EventType]HandlerFactory{
		PlayerLoginType: NewLoginHandlerFactory(
			messages,
			characterLoaders,
		),
		PlayerLogoutType: NewLogoutHandlerFactory(
			messages,
			characterLoaders,
		),
	}
}
