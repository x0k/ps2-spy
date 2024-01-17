package alerts_multi_loader

import (
	"context"
	"maps"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/loaders/multi_loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type AlertsMultiLoader struct {
	alerts         *containers.QueriedLoadableValue[string, string, loaders.Loaded[ps2.Alerts]]
	fallbackLoader *loaders.FallbackLoader[loaders.Loaded[ps2.Alerts]]
	loaders        []string
}

func New(
	loadersMap map[string]loaders.Loader[loaders.Loaded[ps2.Alerts]],
	priority []string,
) *AlertsMultiLoader {
	loadersWithDefault := maps.Clone(loadersMap)
	fallbackLoader := loaders.NewFallbackLoader[loaders.Loaded[ps2.Alerts]](
		"Alerts",
		loadersMap,
		priority,
	)
	loadersWithDefault[multi_loaders.DefaultLoader] = fallbackLoader
	multiLoader := loaders.NewMultiLoader(loadersWithDefault)
	alerts := containers.NewKeyedLoadableValue(multiLoader, len(priority)+1, time.Minute)
	return &AlertsMultiLoader{
		alerts:         alerts,
		fallbackLoader: fallbackLoader,
		loaders:        priority,
	}
}

func (l *AlertsMultiLoader) Start(ctx context.Context) {
	l.fallbackLoader.Start(ctx)
}

func (l *AlertsMultiLoader) Stop() {
	l.fallbackLoader.Stop()
}

func (l *AlertsMultiLoader) Loaders() []string {
	return l.loaders
}

func (l *AlertsMultiLoader) Load(ctx context.Context, provider string) (loaders.Loaded[ps2.Alerts], error) {
	return l.alerts.Load(ctx, multi_loaders.LoaderName(provider))
}