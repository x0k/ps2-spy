package world_territory_control_loader

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type WorldsTrackerLoader struct {
	log            *logger.Logger
	botName        string
	worldsTrackers map[platforms.Platform]*worlds_tracker.WorldsTracker
}

func NewWorldsTrackerLoader(
	log *logger.Logger,
	botName string,
	worldsTrackers map[platforms.Platform]*worlds_tracker.WorldsTracker,
) *WorldsTrackerLoader {
	return &WorldsTrackerLoader{
		log: log.With(
			slog.String(
				"component",
				"loaders.world_territory_control_loader.WorldsTrackerLoader",
			),
		),
		botName:        botName,
		worldsTrackers: worldsTrackers,
	}
}

func (l *WorldsTrackerLoader) Load(
	ctx context.Context,
	worldId ps2.WorldId,
) (loaders.Loaded[ps2.WorldTerritoryControl], error) {
	const op = "loaders.world_territory_control_loader.WorldsTrackerLoader.Load"
	platform, ok := ps2.WorldPlatforms[worldId]
	if !ok {
		return loaders.Loaded[ps2.WorldTerritoryControl]{}, fmt.Errorf("%s: unknown world %q", op, worldId)
	}
	tracker, ok := l.worldsTrackers[platform]
	if !ok {
		return loaders.Loaded[ps2.WorldTerritoryControl]{}, fmt.Errorf("%s: no worlds tracker for platform %s", op, platform)
	}
	territoryControl, err := tracker.WorldTerritoryControl(ctx, worldId)
	if err != nil {
		return loaders.Loaded[ps2.WorldTerritoryControl]{}, fmt.Errorf("%s: getting territory control: %w", op, err)
	}
	return loaders.LoadedNow(l.botName, territoryControl), nil
}
