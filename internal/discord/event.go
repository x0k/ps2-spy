package discord

import (
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	PlayerLoginType EventType = "player_login"
)
