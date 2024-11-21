package discord_commands

import (
	"context"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/x0k/ps2-spy/internal/lib/cache/memory"
	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type alertsLoader struct {
	fallbacks *containers.Fallbacks[loader.Simple[meta.Loaded[ps2.Alerts]]]
	load      loader.Keyed[string, meta.Loaded[ps2.Alerts]]
}

func newAlertsLoader(
	log *logger.Logger,
	loaders map[string]loader.Simple[meta.Loaded[ps2.Alerts]],
	loadersPriority []string,
) *alertsLoader {
	fallbacks := containers.NewFallbacks(
		log.Logger.With(sl.Component("alerts_loader_fallbacks")),
		loaders,
		loadersPriority,
		time.Hour,
	)
	fallbackLoader := loader.NewFallback(fallbacks)
	cached := loader.WithQueriedCache(
		log.Logger.With(sl.Component("alerts_loader_cache")),
		func(ctx context.Context, provider string) (meta.Loaded[ps2.Alerts], error) {
			if loader, ok := loaders[provider]; ok {
				return loader(ctx)
			}
			return fallbackLoader(ctx)
		},
		memory.NewKeyedExpirableCache(
			expirable.NewLRU[string, meta.Loaded[ps2.Alerts]](
				len(loaders)+1,
				nil,
				time.Minute,
			),
		),
	)
	return &alertsLoader{
		fallbacks: fallbacks,
		load:      loader.Keyed[string, meta.Loaded[ps2.Alerts]](cached),
	}
}

func (l *alertsLoader) Start(ctx context.Context) {
	l.fallbacks.Start(ctx)
}
