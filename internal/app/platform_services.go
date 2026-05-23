package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	sql_facility_cache "github.com/x0k/ps2-spy/internal/cache/facility/sql"
	sql_outfits_cache "github.com/x0k/ps2-spy/internal/cache/outfits/sql"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	census_data_provider "github.com/x0k/ps2-spy/internal/data_providers/census"
	"github.com/x0k/ps2-spy/internal/lib/cache/memory"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/metrics"
	events_module "github.com/x0k/ps2-spy/internal/modules/events"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/shared"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	storage_sql "github.com/x0k/ps2-spy/internal/storage/sql"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type PlatformServices struct {
	WorldTrackerSubsManagers map[ps2_platforms.Platform]pubsub.SubscriptionsManager[worlds_tracker.EventType]
	WorldTrackers            map[ps2_platforms.Platform]*worlds_tracker.WorldsTracker
	CharactersLoaders        map[ps2_platforms.Platform]loader.Multi[ps2.CharacterId, ps2.Character]
	CharacterLoaders         map[ps2_platforms.Platform]loader.Keyed[ps2.CharacterId, ps2.Character]
	OutfitsLoaders           map[ps2_platforms.Platform]loader.Multi[ps2.OutfitId, ps2.Outfit]
	OutfitLoaders            map[ps2_platforms.Platform]loader.Keyed[ps2.OutfitId, ps2.Outfit]
	FacilityLoaders          map[ps2_platforms.Platform]loader.Keyed[ps2.FacilityId, ps2.Facility]
}

func newPlatformServices(
	log *logger.Logger,
	cfg *Config,
	m *module.Root,
	mt *metrics.Metrics,
	censusDataProvider *census_data_provider.DataProvider,
	store *storage_sql.Storage,
	facilityCache *sql_facility_cache.Cache,
	charactersTracker *characters_tracker.Tracker,
	statsTracker *stats_tracker.StatsTracker,
	worldTrackerSubsManagers map[ps2_platforms.Platform]pubsub.SubscriptionsManager[worlds_tracker.EventType],
	worldTrackers map[ps2_platforms.Platform]*worlds_tracker.WorldsTracker,
	charactersLoaders map[ps2_platforms.Platform]loader.Multi[ps2.CharacterId, ps2.Character],
	characterLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.CharacterId, ps2.Character],
	outfitsLoaders map[ps2_platforms.Platform]loader.Multi[ps2.OutfitId, ps2.Outfit],
	outfitLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.OutfitId, ps2.Outfit],
	facilityLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.FacilityId, ps2.Facility],
) (*PlatformServices, error) {
	for _, platform := range ps2_platforms.Platforms {
		pl := log.With(slog.String("platform", string(platform)))
		ns := ps2_platforms.PlatformNamespace(platform)

		eventsPubSub := pubsub.New[events.EventType]()

		eventsModule, err := events_module.New(
			pl.With(sl.Module("events")),
			platform,
			cfg.Census.StreamingEndpoint,
			cfg.Census.ServiceId,
			eventsPubSub,
			mt,
		)
		if err != nil {
			return nil, err
		}
		m.Append(eventsModule)

		charactersLoader := metrics.InstrumentMultiKeyedLoaderWithSubjectsCounter(
			metrics.PlatformLoaderSubjectsCounterMetric(mt, metrics.CharactersPlatformLoaderName, platform),
			func(ctx context.Context, k []ps2.CharacterId) (map[ps2.CharacterId]ps2.Character, error) {
				return censusDataProvider.Characters(ctx, platform, k)
			},
		)

		charactersCache := expirable.NewLRU[ps2.CharacterId, ps2.Character](0, nil, 24*time.Hour)
		cachedCharactersLoader := loader.WithMultiCache(
			pl.Logger.With(sl.Component("characters_loader_cache")),
			charactersLoader,
			memory.NewMultiExpirableCache(charactersCache),
		)
		charactersLoaders[platform] = cachedCharactersLoader

		batchedCharactersLoader := loader.WithBatching(
			cachedCharactersLoader,
			10*time.Second,
			shared.ErrNotFound,
		)
		m.AppendVR(
			fmt.Sprintf("%s.batched_characters_loader", platform),
			batchedCharactersLoader.Start,
		)

		cachedBatchedCharactersLoader := loader.Keyed[ps2.CharacterId, ps2.Character](
			loader.WithQueriedCache(
				pl.Logger.With(sl.Component("cached_batched_characters_loader")),
				metrics.InstrumentQueriedLoaderWithCounterMetric(
					metrics.PlatformLoadsCounterMetric(mt, metrics.CharacterPlatformLoaderName, platform),
					batchedCharactersLoader.Load,
				),
				memory.NewKeyedExpirableCache(charactersCache),
			),
		)
		characterLoaders[platform] = cachedBatchedCharactersLoader

		worldsTrackerPubSub := pubsub.New[worlds_tracker.EventType]()

		worldsTackerPublisher := metrics.InstrumentPlatformPublisher(
			mt,
			metrics.WorldsTrackerPlatformPublisher,
			platform,
			worldsTrackerPubSub,
		)
		worldTrackerSubsManagers[platform] = worldsTrackerPubSub

		worldsTracker := worlds_tracker.New(
			pl.With(sl.Component("worlds_tracker")),
			platform,
			5*time.Minute,
			worldsTackerPublisher,
			func(ctx context.Context, wi ps2.WorldId) (ps2.WorldMap, error) {
				return censusDataProvider.WorldMap(ctx, ns, wi)
			},
		)
		m.AppendR(fmt.Sprintf("%s.worlds_tracker", platform), worldsTracker.Start)
		worldTrackers[platform] = worldsTracker

		m.Append(newEventsSubscriptionService(
			pl.With(sl.Component("events_subscription_service")),
			platform,
			m,
			eventsPubSub,
			charactersTracker,
			worldsTracker,
			statsTracker,
		))

		outfitsLoader := loader.WithMultiCache(
			log.Logger.With(sl.Component("outfits_loader_cache")),
			func(ctx context.Context, k []ps2.OutfitId) (map[ps2.OutfitId]ps2.Outfit, error) {
				return censusDataProvider.Outfits(ctx, platform, k)
			},
			sql_outfits_cache.New(
				log.With(sl.Component("outfits_cache")),
				store,
				platform,
			),
		)
		outfitsLoaders[platform] = outfitsLoader
		outfitLoaders[platform] = func(ctx context.Context, oi ps2.OutfitId) (ps2.Outfit, error) {
			outfit, err := outfitsLoader(ctx, []ps2.OutfitId{oi})
			if err != nil {
				return ps2.Outfit{}, err
			}
			return outfit[oi], nil
		}

		facilityLoaders[platform] = loader.WithKeyedCache(
			log.Logger.With(sl.Component("facilities_loader_cache")),
			func(ctx context.Context, id ps2.FacilityId) (ps2.Facility, error) {
				return censusDataProvider.Facility(ctx, ns, id)
			},
			facilityCache,
		)
	}

	return &PlatformServices{
		WorldTrackerSubsManagers: worldTrackerSubsManagers,
		WorldTrackers:            worldTrackers,
		CharactersLoaders:        charactersLoaders,
		CharacterLoaders:         characterLoaders,
		OutfitsLoaders:           outfitsLoaders,
		OutfitLoaders:            outfitLoaders,
		FacilityLoaders:          facilityLoaders,
	}, nil
}
