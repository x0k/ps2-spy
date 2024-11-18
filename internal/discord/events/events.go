package discord_events

import (
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/discord"
)

type PlayerLogin characters_tracker.PlayerLogin

func (e PlayerLogin) Type() discord.EventType {
	return discord.PlayerLoginType
}
