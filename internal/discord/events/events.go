package discord_events

import (
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	PlayerLoginType  EventType = "player_login"
	PlayerLogoutType EventType = "player_logout"
)

type PlayerLogin characters_tracker.PlayerLogin

func (e PlayerLogin) Type() EventType {
	return PlayerLoginType
}

type PlayerLogout characters_tracker.PlayerLogout

func (e PlayerLogout) Type() EventType {
	return PlayerLoginType
}
