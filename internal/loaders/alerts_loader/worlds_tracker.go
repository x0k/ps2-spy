package alerts_loader

import (
	"context"
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
			slog.String("component", "loaders.alerts_loader.WorldsTrackerLoader"),
		),
		botName:        botName,
		worldsTrackers: worldsTrackers,
	}
}

func (l *WorldsTrackerLoader) Load(ctx context.Context) (loaders.Loaded[ps2.Alerts], error) {
	const op = "loaders.alerts_loader.WorldsTrackerLoader.Load"
	alerts := make(ps2.Alerts, 0)
	for _, platform := range platforms.Platforms {
		tracker, ok := l.worldsTrackers[platform]
		if !ok {
			l.log.Warn(ctx, "no alerts tracker for platform", slog.String("platform", string(platform)))
			continue
		}
		alerts = append(alerts, tracker.Alerts()...)
	}
	return loaders.LoadedNow(l.botName, alerts), nil
}
