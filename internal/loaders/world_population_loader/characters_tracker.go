package world_population_loader

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type CharactersTrackerLoader struct {
	botName            string
	charactersTrackers map[platforms.Platform]*characters_tracker.CharactersTracker
}

func NewCharactersTrackerLoader(
	botName string,
	charactersTrackers map[platforms.Platform]*characters_tracker.CharactersTracker,
) *CharactersTrackerLoader {
	return &CharactersTrackerLoader{
		botName:            botName,
		charactersTrackers: charactersTrackers,
	}
}

func (l *CharactersTrackerLoader) Load(ctx context.Context, worldId ps2.WorldId) (loaders.Loaded[ps2.DetailedWorldPopulation], error) {
	const op = "loaders.world_population_loader.CharactersTrackerLoader.Load"
	platform, ok := ps2.WorldPlatforms[worldId]
	if !ok {
		return loaders.Loaded[ps2.DetailedWorldPopulation]{}, fmt.Errorf("%s: unknown world %q", op, worldId)
	}
	tracker, ok := l.charactersTrackers[platform]
	if !ok {
		return loaders.Loaded[ps2.DetailedWorldPopulation]{}, fmt.Errorf("%s: no population tracker for platform %s", op, platform)
	}
	population, err := tracker.DetailedWorldPopulation(worldId)
	if err != nil {
		return loaders.Loaded[ps2.DetailedWorldPopulation]{}, fmt.Errorf("%s: getting population: %w", op, err)
	}
	return loaders.LoadedNow(l.botName, population), nil
}
