package characters_tracker

import (
	"time"

	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	PlayerLoginType     EventType = "player_login"
	PlayerFakeLoginType EventType = "player_fake_login"
	PlayerLogoutType    EventType = "player_logout"
)

type PlayerLogin struct {
	Time      time.Time
	Character ps2.Character
}

func (e PlayerLogin) Type() EventType {
	return PlayerLoginType
}

type PlayerFakeLogin struct {
	Time      time.Time
	Character ps2.Character
}

func (e PlayerFakeLogin) Type() EventType {
	return PlayerFakeLoginType
}

type PlayerLogout struct {
	Time        time.Time
	CharacterId ps2.CharacterId
	WorldId     ps2.WorldId
}

func (e PlayerLogout) Type() EventType {
	return PlayerLogoutType
}
