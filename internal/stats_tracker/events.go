package stats_tracker

import (
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
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
	ChannelId discord.ChannelId
	StartedAt time.Time
	StoppedAt time.Time
	Platforms map[ps2_platforms.Platform]PlatformStats
}

func (e ChannelTrackerStopped) Type() EventType {
	return ChannelTrackerStoppedType
}
