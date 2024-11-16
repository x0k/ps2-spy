package characters_tracker_population_loader

import (
	"context"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func New(
	log *logger.Logger,
	botName string,
	charactersTrackers map[ps2_platforms.Platform]*characters_tracker.CharactersTracker,
) func(context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
	return func(ctx context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
		total := 0
		worlds := make([]ps2.WorldPopulation, 0)
		for _, platform := range ps2_platforms.Platforms {
			tracker, ok := charactersTrackers[platform]
			if !ok {
				log.Warn(ctx, "no population tracker for platform", slog.String("platform", string(platform)))
				continue
			}
			population := tracker.WorldsPopulation()
			total += population.Total
			worlds = append(worlds, population.Worlds...)
		}
		return meta.LoadedNow(botName, ps2.WorldsPopulation{
			Total:  total,
			Worlds: worlds,
		}), nil
	}
}