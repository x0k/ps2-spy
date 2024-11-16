package characters_tracker_world_population_loader

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func New(
	botName string,
	charactersTrackers map[ps2_platforms.Platform]*characters_tracker.CharactersTracker,
) func(context.Context, ps2.WorldId) (meta.Loaded[ps2.DetailedWorldPopulation], error) {
	return func(ctx context.Context, worldId ps2.WorldId) (meta.Loaded[ps2.DetailedWorldPopulation], error) {
		const op = "loaders.world_population_loader.CharactersTrackerLoader.Load"
		platform, ok := ps2.WorldPlatforms[worldId]
		if !ok {
			return meta.Loaded[ps2.DetailedWorldPopulation]{}, fmt.Errorf("%s: unknown world %q", op, worldId)
		}
		tracker, ok := charactersTrackers[platform]
		if !ok {
			return meta.Loaded[ps2.DetailedWorldPopulation]{}, fmt.Errorf("%s: no population tracker for platform %s", op, platform)
		}
		population, err := tracker.DetailedWorldPopulation(worldId)
		if err != nil {
			return meta.Loaded[ps2.DetailedWorldPopulation]{}, fmt.Errorf("%s: getting population: %w", op, err)
		}
		return meta.LoadedNow(botName, population), nil
	}
}
