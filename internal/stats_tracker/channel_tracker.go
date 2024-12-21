package stats_tracker

import (
	"time"

	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type channelTracker struct {
	trackers  map[ps2_platforms.Platform]*platformTracker
	startedAt time.Time
}
