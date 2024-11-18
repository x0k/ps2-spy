package discord_handlers

import (
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func New(
	message discord.LocalizedMessages,
	characterLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.CharacterId, ps2.Character],
) map[discord.EventType]discord.HandlerFactory {
	return map[discord.EventType]discord.HandlerFactory{
		discord.PlayerLoginType: NewLoginHandlerFactory(
			message,
			characterLoaders,
		),
	}
}
