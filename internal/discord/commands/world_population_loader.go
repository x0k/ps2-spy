package discord_commands

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/cache/memory"
	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type worldPopulationLoader struct {
	name      string
	fallbacks *containers.Fallbacks[loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]]]
	load      loader.Queried[query[ps2.WorldId], meta.Loaded[ps2.DetailedWorldPopulation]]
}

func newWorldPopulationLoader(
	name string,
	log *logger.Logger,
	loaders map[string]loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]],
	loadersPriority []string,
) *worldPopulationLoader {
	fallbacks := containers.NewFallbacks(
		log.Logger.With(sl.Component("world_population_loader_fallbacks")),
		loaders,
		loadersPriority,
		time.Hour,
	)
	withDefault := maps.Clone(loaders)
	withDefault[defaultProvider] = loader.NewKeyedFallback(fallbacks)
	cached := loader.WithQueriedCache(
		log.With(sl.Component("world_population_loader_cache")),
		func(ctx context.Context, query query[ps2.WorldId]) (meta.Loaded[ps2.DetailedWorldPopulation], error) {
			if loader, ok := withDefault[providerName(query.Provider)]; ok {
				return loader(ctx, query.Key)
			}
			return meta.Loaded[ps2.DetailedWorldPopulation]{}, fmt.Errorf("unknown provider: %s", query.Provider)
		},
		memory.NewKeyedExpirableCache[query[ps2.WorldId], meta.Loaded[ps2.DetailedWorldPopulation]](
			(len(loaders)+1)*len(ps2.ZoneNames),
			time.Minute,
		),
	)
	return &worldPopulationLoader{
		name:      name,
		fallbacks: fallbacks,
		load:      loader.Queried[query[ps2.WorldId], meta.Loaded[ps2.DetailedWorldPopulation]](cached),
	}
}

func (l *worldPopulationLoader) Name() string {
	return l.name
}

func (l *worldPopulationLoader) Start(ctx context.Context) {
	l.fallbacks.Start(ctx)
}
