package discord_events

import (
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/discord"
)

type base[E any] struct {
	channels []discord.ChannelId
	event    E
}

type PlayerLogin base[characters_tracker.PlayerLogin]

func (e PlayerLogin) Type() discord.EventType {
	return discord.PlayerLoginType
}
