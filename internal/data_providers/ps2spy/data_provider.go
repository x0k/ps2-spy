package ps2spy_data_provider

import (
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type WorldTrackerProvider interface {
	WorldTracker(platform ps2_platforms.Platform) *worlds_tracker.WorldsTracker
}

type DataProvider struct {
	appName           string
	log               *logger.Logger
	charactersTracker *characters_tracker.Tracker
	worldTracker      WorldTrackerProvider
}

func New(
	log *logger.Logger,
	appName string,
	charactersTracker *characters_tracker.Tracker,
	worldTracker WorldTrackerProvider,
) *DataProvider {
	return &DataProvider{
		log:               log,
		appName:           appName,
		charactersTracker: charactersTracker,
		worldTracker:      worldTracker,
	}
}
