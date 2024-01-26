package population_loader

import (
	"context"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/population_tracker"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type PopulationTrackerLoader struct {
	botName            string
	populationTrackers map[platforms.Platform]*population_tracker.PopulationTracker
}

func NewPopulationTrackerLoader(
	botName string,
	populationTrackers map[platforms.Platform]*population_tracker.PopulationTracker,
) *PopulationTrackerLoader {
	return &PopulationTrackerLoader{
		botName:            botName,
		populationTrackers: populationTrackers,
	}
}

func (l *PopulationTrackerLoader) Load(ctx context.Context) (loaders.Loaded[ps2.WorldsPopulation], error) {
	const op = "loaders.population_loader.PopulationTrackerLoader.Load"
	log := infra.OpLogger(ctx, op)
	total := 0
	worlds := make([]ps2.WorldPopulation, 0)
	for _, platform := range platforms.Platforms {
		tracker, ok := l.populationTrackers[platform]
		if !ok {
			log.Warn("no population tracker for platform", slog.String("platform", string(platform)))
			continue
		}
		population := tracker.WorldsPopulation()
		total += population.Total
		worlds = append(worlds, population.Worlds...)
	}
	return loaders.LoadedNow(l.botName, ps2.WorldsPopulation{
		Total:  total,
		Worlds: worlds,
	}), nil
}
