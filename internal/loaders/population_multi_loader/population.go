package population_multi_loader

import (
	"context"
	"maps"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/loaders/multi_loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type PopulationMultiLoader struct {
	value          *containers.QueriedLoadableValue[string, string, loaders.Loaded[ps2.WorldsPopulation]]
	fallbackLoader *loaders.FallbackLoader[loaders.Loaded[ps2.WorldsPopulation]]
	loaders        []string
}

func New(
	loadersMap map[string]loaders.Loader[loaders.Loaded[ps2.WorldsPopulation]],
	priority []string,
) *PopulationMultiLoader {
	loadersWithDefault := maps.Clone(loadersMap)
	fallbackLoader := loaders.NewFallbackLoader(
		"Population",
		loadersMap,
		priority,
	)
	loadersWithDefault[multi_loaders.DefaultLoader] = fallbackLoader
	multiLoader := loaders.NewMultiLoader(loadersWithDefault)
	value := containers.NewKeyedLoadableValue(multiLoader, len(priority)+1, time.Minute)
	return &PopulationMultiLoader{
		value:          value,
		fallbackLoader: fallbackLoader,
		loaders:        priority,
	}
}

func (l *PopulationMultiLoader) Start(ctx context.Context) {
	l.fallbackLoader.Start(ctx)
}

func (l *PopulationMultiLoader) Stop() {
	l.fallbackLoader.Stop()
}

func (l *PopulationMultiLoader) Loaders() []string {
	return l.loaders
}

func (l *PopulationMultiLoader) Load(ctx context.Context, provider string) (loaders.Loaded[ps2.WorldsPopulation], error) {
	return l.value.Load(ctx, multi_loaders.LoaderName(provider))
}
