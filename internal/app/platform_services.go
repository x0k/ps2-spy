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
	ps2_storage_outfits_repo "github.com/x0k/ps2-spy/internal/ps2/storage_outfits_repo"
	"github.com/x0k/ps2-spy/internal/shared"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type PlatformServices struct {
	worldTrackerSubsManagers map[ps2_platforms.Platform]pubsub.SubscriptionsManager[worlds_tracker.EventType]
	worldTrackers            map[ps2_platforms.Platform]*worlds_tracker.WorldsTracker
	charactersLoaders        map[ps2_platforms.Platform]loader.Multi[ps2.CharacterId, ps2.Character]
	characterLoaders         map[ps2_platforms.Platform]loader.Keyed[ps2.CharacterId, ps2.Character]
	outfitsLoaders           map[ps2_platforms.Platform]loader.Multi[ps2.OutfitId, ps2.Outfit]
	outfitLoaders            map[ps2_platforms.Platform]loader.Keyed[ps2.OutfitId, ps2.Outfit]
	facilityLoaders          map[ps2_platforms.Platform]loader.Keyed[ps2.FacilityId, ps2.Facility]
}

func newPlatformServices() *PlatformServices {
	return &PlatformServices{
		worldTrackerSubsManagers: make(map[ps2_platforms.Platform]pubsub.SubscriptionsManager[worlds_tracker.EventType], len(ps2_platforms.Platforms)),
		worldTrackers:            make(map[ps2_platforms.Platform]*worlds_tracker.WorldsTracker, len(ps2_platforms.Platforms)),
		charactersLoaders:        make(map[ps2_platforms.Platform]loader.Multi[ps2.CharacterId, ps2.Character], len(ps2_platforms.Platforms)),
		characterLoaders:         make(map[ps2_platforms.Platform]loader.Keyed[ps2.CharacterId, ps2.Character], len(ps2_platforms.Platforms)),
		outfitsLoaders:           make(map[ps2_platforms.Platform]loader.Multi[ps2.OutfitId, ps2.Outfit], len(ps2_platforms.Platforms)),
		outfitLoaders:            make(map[ps2_platforms.Platform]loader.Keyed[ps2.OutfitId, ps2.Outfit], len(ps2_platforms.Platforms)),
		facilityLoaders:          make(map[ps2_platforms.Platform]loader.Keyed[ps2.FacilityId, ps2.Facility], len(ps2_platforms.Platforms)),
	}
}

func (p *PlatformServices) WorldTracker(platform ps2_platforms.Platform) *worlds_tracker.WorldsTracker {
	return p.worldTrackers[platform]
}

func (p *PlatformServices) CharacterLoader(platform ps2_platforms.Platform) loader.Keyed[ps2.CharacterId, ps2.Character] {
	return p.characterLoaders[platform]
}

func (p *PlatformServices) CharactersLoader(platform ps2_platforms.Platform) loader.Multi[ps2.CharacterId, ps2.Character] {
	return p.charactersLoaders[platform]
}

func (p *PlatformServices) OutfitLoader(platform ps2_platforms.Platform) loader.Keyed[ps2.OutfitId, ps2.Outfit] {
	return p.outfitLoaders[platform]
}

func (p *PlatformServices) OutfitsLoader(platform ps2_platforms.Platform) loader.Multi[ps2.OutfitId, ps2.Outfit] {
	return p.outfitsLoaders[platform]
}

func (p *PlatformServices) FacilityLoader(platform ps2_platforms.Platform) loader.Keyed[ps2.FacilityId, ps2.Facility] {
	return p.facilityLoaders[platform]
}

func (p *PlatformServices) WorldTrackerSubsManager(platform ps2_platforms.Platform) pubsub.SubscriptionsManager[worlds_tracker.EventType] {
	return p.worldTrackerSubsManagers[platform]
}

func (p *PlatformServices) Init(
	log *logger.Logger,
	cfg *Config,
	m *module.Root,
	mt *metrics.Metrics,
	censusDataProvider *census_data_provider.DataProvider,
	outfitsRepo *ps2_storage_outfits_repo.Repository,
	facilityCache *sql_facility_cache.Cache,
	charactersTracker *characters_tracker.Tracker,
	statsTracker *stats_tracker.StatsTracker,
) error {
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
			return err
		}
		m.Append(eventsModule)

		charactersLoader := metrics.InstrumentMultiKeyedLoaderWithSubjectsCounter(
			metrics.PlatformLoaderSubjectsCounterMetric(mt, metrics.CharactersPlatformLoaderName, platform),
			func(ctx context.Context, k []ps2.CharacterId) (map[ps2.CharacterId]ps2.Character, error) {
				return censusDataProvider.Characters(ctx, platform, k)
			},
		)

		charactersCache := expirable.NewLRU[ps2.CharacterId, ps2.Character](50000, nil, 24*time.Hour)
		cachedCharactersLoader := loader.WithMultiCache(
			pl.Logger.With(sl.Component("characters_loader_cache")),
			charactersLoader,
			memory.NewMultiExpirableCache(charactersCache),
		)
		p.charactersLoaders[platform] = cachedCharactersLoader

		batchedCharactersLoader := loader.WithBatching(
			cachedCharactersLoader,
			10*time.Second,
			shared.ErrNotFound,
		)
		m.AppendVR(
			fmt.Sprintf("%s.batched_characters_loader", platform),
			batchedCharactersLoader.Start,
		)

		cachedBatchedCharactersLoader := loader.WithQueriedCache(
			pl.Logger.With(sl.Component("cached_batched_characters_loader")),
			metrics.InstrumentQueriedLoaderWithCounterMetric(
				metrics.PlatformLoadsCounterMetric(mt, metrics.CharacterPlatformLoaderName, platform),
				batchedCharactersLoader.Load,
			),
			memory.NewKeyedExpirableCache(charactersCache),
		)
		p.characterLoaders[platform] = cachedBatchedCharactersLoader

		worldsTrackerPubSub := pubsub.New[worlds_tracker.EventType]()

		worldsTackerPublisher := metrics.InstrumentPlatformPublisher(
			mt,
			metrics.WorldsTrackerPlatformPublisher,
			platform,
			worldsTrackerPubSub,
		)
		p.worldTrackerSubsManagers[platform] = worldsTrackerPubSub

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
		p.worldTrackers[platform] = worldsTracker

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
			metrics.InstrumentMultiKeyedLoaderWithSubjectsCounter(
				metrics.PlatformLoaderSubjectsCounterMetric(mt, metrics.OutfitsPlatformLoaderName, platform),
				func(ctx context.Context, k []ps2.OutfitId) (map[ps2.OutfitId]ps2.Outfit, error) {
					return censusDataProvider.Outfits(ctx, platform, k)
				},
			),
			sql_outfits_cache.New(
				log.With(sl.Component("outfits_cache")),
				outfitsRepo,
				platform,
			),
		)
		p.outfitsLoaders[platform] = outfitsLoader

		outfitCache := expirable.NewLRU[ps2.OutfitId, ps2.Outfit](5000, nil, 5*time.Minute)
		p.outfitLoaders[platform] = loader.WithQueriedCache(
			log.Logger.With(sl.Component("outfit_loader_cache")),
			metrics.InstrumentQueriedLoaderWithCounterMetric(
				metrics.PlatformLoadsCounterMetric(mt, metrics.OutfitPlatformLoaderName, platform),
				func(ctx context.Context, oi ps2.OutfitId) (ps2.Outfit, error) {
					outfits, err := outfitsLoader(ctx, []ps2.OutfitId{oi})
					if err != nil {
						return ps2.Outfit{}, err
					}
					return outfits[oi], nil
				},
			),
			memory.NewKeyedExpirableCache(outfitCache),
		)

		p.facilityLoaders[platform] = loader.WithKeyedCache(
			log.Logger.With(sl.Component("facilities_loader_cache")),
			metrics.InstrumentQueriedLoaderWithCounterMetric(
				metrics.PlatformLoadsCounterMetric(mt, metrics.FacilitiesPlatformLoaderName, platform),
				func(ctx context.Context, id ps2.FacilityId) (ps2.Facility, error) {
					return censusDataProvider.Facility(ctx, ns, id)
				},
			),
			facilityCache,
		)
	}

	return nil
}
