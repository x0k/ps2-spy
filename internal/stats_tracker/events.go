package stats_tracker

import (
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	ChannelTrackerStartedType EventType = "start_stats_tracker"
	ChannelTrackerStoppedType EventType = "stop_stats_tracker"
)

type ChannelTrackerStarted struct {
	ChannelId discord.ChannelId
	StartedAt time.Time
}

func (e ChannelTrackerStarted) Type() EventType {
	return ChannelTrackerStartedType
}

type ChannelTrackerStopped struct {
	ChannelId  discord.ChannelId
	StartedAt  time.Time
	Characters map[ps2.CharacterId]*CharacterStats
}

func (e ChannelTrackerStopped) Type() EventType {
	return ChannelTrackerStoppedType
}
