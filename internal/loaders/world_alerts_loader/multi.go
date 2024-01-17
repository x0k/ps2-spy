package world_alerts_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/loaders/alerts_loader"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type MultiLoader struct {
	loader *alerts_loader.MultiLoader
}

func NewMulti(
	loader *alerts_loader.MultiLoader,
) *MultiLoader {
	return &MultiLoader{loader}
}

func (l *MultiLoader) Start(ctx context.Context) {
	l.loader.Start(ctx)
}

func (l *MultiLoader) Stop() {
	l.loader.Stop()
}

func (l *MultiLoader) Loaders() []string {
	return l.loader.Loaders()
}

func (l *MultiLoader) Load(ctx context.Context, query loaders.MultiLoaderQuery[ps2.WorldId]) (loaders.Loaded[ps2.Alerts], error) {
	loaded, err := l.loader.Load(ctx, query.Loader)
	if err != nil {
		return loaders.Loaded[ps2.Alerts]{}, err
	}
	worldAlerts := make(ps2.Alerts, 0, len(loaded.Value))
	for _, a := range loaded.Value {
		if a.WorldId == query.Key {
			worldAlerts = append(worldAlerts, a)
		}
	}
	loaded.Value = worldAlerts
	return loaded, nil
}
