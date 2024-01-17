package world_population_loader

import (
	"context"
	"maps"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/loaders/multi_loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type MultiLoader struct {
	value          *containers.QueriedLoadableValue[loaders.MultiLoaderQuery[ps2.WorldId], loaders.MultiLoaderQuery[ps2.WorldId], loaders.Loaded[ps2.DetailedWorldPopulation]]
	fallbackLoader *loaders.KeyedFallbackLoader[ps2.WorldId, loaders.Loaded[ps2.DetailedWorldPopulation]]
	loaders        []string
}

func NewMulti(
	loadersMap map[string]loaders.KeyedLoader[ps2.WorldId, loaders.Loaded[ps2.DetailedWorldPopulation]],
	priority []string,
) *MultiLoader {
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
	return &MultiLoader{
		value:          value,
		fallbackLoader: fallbackLoader,
		loaders:        priority,
	}
}

func (l *MultiLoader) Start(ctx context.Context, wg *sync.WaitGroup) {
	l.fallbackLoader.Start(ctx, wg)
}

func (l *MultiLoader) Loaders() []string {
	return l.loaders
}

func (l *MultiLoader) Load(ctx context.Context, query loaders.MultiLoaderQuery[ps2.WorldId]) (loaders.Loaded[ps2.DetailedWorldPopulation], error) {
	query.Loader = multi_loaders.LoaderName(query.Loader)
	return l.value.Load(ctx, query)
}
