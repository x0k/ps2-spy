package alerts

import (
	"context"

	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type WorldAlertsLoader struct {
	loader *AlertsMultiLoader
}

func NewWorldAlertsLoader(
	loader *AlertsMultiLoader,
) *WorldAlertsLoader {
	return &WorldAlertsLoader{loader}
}

func (l *WorldAlertsLoader) Start() {
	l.loader.Start()
}

func (l *WorldAlertsLoader) Stop() {
	l.loader.Stop()
}

func (l *WorldAlertsLoader) Loaders() []string {
	return l.loader.Loaders()
}

func (l *WorldAlertsLoader) Load(ctx context.Context, query loaders.MultiLoaderQuery[ps2.WorldId]) (loaders.Loaded[ps2.Alerts], error) {
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
