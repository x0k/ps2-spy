package population_loader

import (
	"context"
	"log/slog"
	"maps"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/loaders/multi_loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type MultiLoader struct {
	value          *containers.QueriedLoadableValue[string, string, loaders.Loaded[ps2.WorldsPopulation]]
	fallbackLoader *loaders.FallbackLoader[loaders.Loaded[ps2.WorldsPopulation]]
	loaders        []string
}

func NewMulti(
	log *slog.Logger,
	loadersMap map[string]loaders.Loader[loaders.Loaded[ps2.WorldsPopulation]],
	priority []string,
) *MultiLoader {
	loadersWithDefault := maps.Clone(loadersMap)
	fallbackLoader := loaders.NewFallbackLoader(
		log.With(slog.String("component", "MultiLoader.WorldsPopulation")),
		loadersMap,
		priority,
	)
	loadersWithDefault[multi_loaders.DefaultLoader] = fallbackLoader
	multiLoader := loaders.NewMultiLoader(loadersWithDefault)
	value := containers.NewKeyedLoadableValue(multiLoader, len(priority)+1, time.Minute)
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

func (l *MultiLoader) Load(ctx context.Context, provider string) (loaders.Loaded[ps2.WorldsPopulation], error) {
	return l.value.Load(ctx, multi_loaders.LoaderName(provider))
}
