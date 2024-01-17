package world_population_multi_loader

import (
	"context"
	"maps"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/loaders/multi_loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type WorldPopulationMultiLoader struct {
	value          *containers.QueriedLoadableValue[loaders.MultiLoaderQuery[ps2.WorldId], loaders.MultiLoaderQuery[ps2.WorldId], loaders.Loaded[ps2.DetailedWorldPopulation]]
	fallbackLoader *loaders.KeyedFallbackLoader[ps2.WorldId, loaders.Loaded[ps2.DetailedWorldPopulation]]
	loaders        []string
}

func New(
	loadersMap map[string]loaders.KeyedLoader[ps2.WorldId, loaders.Loaded[ps2.DetailedWorldPopulation]],
	priority []string,
) *WorldPopulationMultiLoader {
	loadersWithDefault := maps.Clone(loadersMap)
	fallbackLoader := loaders.NewKeyedFallbackLoader(
		"World",
		loadersMap,
		priority,
	)
	loadersWithDefault[multi_loaders.DefaultLoader] = fallbackLoader
	multiLoader := loaders.NewKeyedMultiLoader(loadersWithDefault)
	value := containers.NewKeyedLoadableValue[loaders.MultiLoaderQuery[ps2.WorldId], loaders.Loaded[ps2.DetailedWorldPopulation]](
		multiLoader,
		(len(priority)+1)*len(ps2.ZoneNames),
		time.Minute,
	)
	return &WorldPopulationMultiLoader{
		value:          value,
		fallbackLoader: fallbackLoader,
		loaders:        priority,
	}
}

func (l *WorldPopulationMultiLoader) Start(ctx context.Context) {
	l.fallbackLoader.Start(ctx)
}

func (l *WorldPopulationMultiLoader) Stop() {
	l.fallbackLoader.Stop()
}

func (l *WorldPopulationMultiLoader) Loaders() []string {
	return l.loaders
}

func (l *WorldPopulationMultiLoader) Load(ctx context.Context, query loaders.MultiLoaderQuery[ps2.WorldId]) (loaders.Loaded[ps2.DetailedWorldPopulation], error) {
	query.Loader = multi_loaders.LoaderName(query.Loader)
	return l.value.Load(ctx, query)
}