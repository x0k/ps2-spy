package world_population_loader

import (
	"context"
	"fmt"

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

func (l *PopulationTrackerLoader) Load(ctx context.Context, worldId ps2.WorldId) (loaders.Loaded[ps2.DetailedWorldPopulation], error) {
	const op = "loaders.world_population_loader.PopulationTrackerLoader.Load"
	platform, ok := ps2.WorldPlatforms[worldId]
	if !ok {
		return loaders.Loaded[ps2.DetailedWorldPopulation]{}, fmt.Errorf("%s: unknown world %q", op, worldId)
	}
	tracker, ok := l.populationTrackers[platform]
	if !ok {
		return loaders.Loaded[ps2.DetailedWorldPopulation]{}, fmt.Errorf("%s: no population tracker for platform %s", op, platform)
	}
	population, err := tracker.DetailedWorldPopulation(worldId)
	if err != nil {
		return loaders.Loaded[ps2.DetailedWorldPopulation]{}, fmt.Errorf("%s: getting population: %w", op, err)
	}
	return loaders.LoadedNow(l.botName, population), nil
}
