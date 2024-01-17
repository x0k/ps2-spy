package alerts_loader

import (
	"context"
	"maps"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/loaders/multi_loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type MultiLoader struct {
	alerts         *containers.QueriedLoadableValue[string, string, loaders.Loaded[ps2.Alerts]]
	fallbackLoader *loaders.FallbackLoader[loaders.Loaded[ps2.Alerts]]
	loaders        []string
}

func NewMulti(
	loadersMap map[string]loaders.Loader[loaders.Loaded[ps2.Alerts]],
	priority []string,
) *MultiLoader {
	loadersWithDefault := maps.Clone(loadersMap)
	fallbackLoader := loaders.NewFallbackLoader[loaders.Loaded[ps2.Alerts]](
		"Alerts",
		loadersMap,
		priority,
	)
	loadersWithDefault[multi_loaders.DefaultLoader] = fallbackLoader
	multiLoader := loaders.NewMultiLoader(loadersWithDefault)
	alerts := containers.NewKeyedLoadableValue(multiLoader, len(priority)+1, time.Minute)
	return &MultiLoader{
		alerts:         alerts,
		fallbackLoader: fallbackLoader,
		loaders:        priority,
	}
}

func (l *MultiLoader) Start(ctx context.Context) {
	l.fallbackLoader.Start(ctx)
}

func (l *MultiLoader) Stop() {
	l.fallbackLoader.Stop()
}

func (l *MultiLoader) Loaders() []string {
	return l.loaders
}

func (l *MultiLoader) Load(ctx context.Context, provider string) (loaders.Loaded[ps2.Alerts], error) {
	return l.alerts.Load(ctx, multi_loaders.LoaderName(provider))
}
