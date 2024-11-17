package discord_module

import (
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

type EventType int

type Event = pubsub.Event[EventType]

const (
	PlayerLoginType EventType = iota
)

type PlayerLogin characters_tracker.PlayerLogin

func (e PlayerLogin) Type() EventType {
	return PlayerLoginType
}
