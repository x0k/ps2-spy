package ps2spy_data_provider

import (
	"context"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func (p *DataProvider) Alerts(ctx context.Context) (meta.Loaded[ps2.Alerts], error) {
	alerts := make(ps2.Alerts, 0)
	for _, platform := range ps2_platforms.Platforms {
		tracker := p.worldTracker.WorldTracker(platform)
		if tracker == nil {
			p.log.Warn(ctx, "no alerts tracker for platform", slog.String("platform", string(platform)))
			continue
		}
		alerts = append(alerts, tracker.Alerts()...)
	}
	return meta.LoadedNow(p.appName, alerts), nil
}
