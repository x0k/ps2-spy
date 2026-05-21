package ps2spy_data_provider

import (
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type DataProvider struct {
	appName           string
	log               *logger.Logger
	charactersTracker *characters_tracker.Tracker
	worldTrackers     map[ps2_platforms.Platform]*worlds_tracker.WorldsTracker
}

func New(
	log *logger.Logger,
	appName string,
	charactersTracker *characters_tracker.Tracker,
	worldTrackers map[ps2_platforms.Platform]*worlds_tracker.WorldsTracker,
) *DataProvider {
	return &DataProvider{
		log:               log,
		appName:           appName,
		charactersTracker: charactersTracker,
		worldTrackers:     worldTrackers,
	}
}
