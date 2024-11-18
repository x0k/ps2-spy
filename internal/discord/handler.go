package discord

import (
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type EventType int

type Event = pubsub.Event[EventType]

const (
	PlayerLoginType EventType = iota
)

type Handler interface {
	ForPlatform(platform ps2_platforms.Platform) pubsub.Handler[EventType]
}
