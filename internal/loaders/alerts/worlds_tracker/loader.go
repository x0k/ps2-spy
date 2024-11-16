package worlds_tracker_alerts_loader

import (
	"context"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func New(
	log *logger.Logger,
	appName string,
	worldsTrackers map[ps2_platforms.Platform]*worlds_tracker.WorldsTracker,
) func(context.Context) (meta.Loaded[ps2.Alerts], error) {
	return func(ctx context.Context) (meta.Loaded[ps2.Alerts], error) {
		alerts := make(ps2.Alerts, 0)
		for _, platform := range ps2_platforms.Platforms {
			tracker, ok := worldsTrackers[platform]
			if !ok {
				log.Warn(ctx, "no alerts tracker for platform", slog.String("platform", string(platform)))
				continue
			}
			alerts = append(alerts, tracker.Alerts()...)
		}
		return meta.LoadedNow(appName, alerts), nil
	}
}
