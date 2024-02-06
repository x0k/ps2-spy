package population_loader

import (
	"context"
	"maps"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/loaders/multi_loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type MultiLoader struct {
	value          *loaders.CachedQueryLoader[string, loaders.Loaded[ps2.WorldsPopulation]]
	fallbackLoader *loaders.FallbackLoader[loaders.Loaded[ps2.WorldsPopulation]]
	loaders        []string
}

func NewMulti(
	log *logger.Logger,
	loadersMap map[string]loaders.Loader[loaders.Loaded[ps2.WorldsPopulation]],
	priority []string,
) *MultiLoader {
	loadersWithDefault := maps.Clone(loadersMap)
	fallbackLoader := loaders.NewFallbackLoader(
		log.Logger,
		loadersMap,
		priority,
	)
	loadersWithDefault[multi_loaders.DefaultLoader] = fallbackLoader
	multiLoader := loaders.NewMultiLoader(loadersWithDefault)
	value := loaders.NewCachedQueriedLoader(
		log.Logger,
		multiLoader,
		containers.NewExpiableLRU[string, loaders.Loaded[ps2.WorldsPopulation]](len(priority)+1, time.Minute),
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

func (l *MultiLoader) Load(ctx context.Context, provider string) (loaders.Loaded[ps2.WorldsPopulation], error) {
	return l.value.Load(ctx, multi_loaders.LoaderName(provider))
}
