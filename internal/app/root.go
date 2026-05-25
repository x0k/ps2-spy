package app

import (
	"net/http"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/metrics"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking"

	// migration tools
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewRoot(cfg *Config, log *logger.Logger) (*module.Module, error) {
	m := module.New(log.Logger, "root", module.WithSignalHandling())

	if cfg.Profiler.Enabled {
		srv := newProfilerService(cfg.Profiler.Address, log.Logger)
		m.Go(module.NewRun(srv.Name(), srv.Run))
	}

	var mt *metrics.Metrics
	if cfg.Metrics.Enabled {
		mt = metrics.New("ps2spy")
		srv := newMetricsService(mt, cfg.Metrics.Address, log.Logger)
		m.Go(module.NewRun(srv.Name(), srv.Run))
	}

	httpClient := &http.Client{
		Timeout: cfg.HttpClient.Timeout,
		Transport: metrics.InstrumentTransport(
			mt,
			metrics.DefaultTransportName,
			http.DefaultTransport,
		),
	}

	storePubSub := pubsub.New[storage.EventType]()
	store := newStorage(log, cfg, m, storePubSub)

	infra, err := newInfrastructure(log, cfg, store, httpClient, mt)
	if err != nil {
		return nil, err
	}

	ps2PubSub := pubsub.New[ps2.EventType]()
	statsTrackerPubSub := pubsub.New[stats_tracker.EventType]()
	charactersTrackerPubSub := pubsub.New[characters_tracker.EventType]()

	outfitSync := newOutfitSync(log, cfg, m, ps2PubSub, infra)

	trackers := newTrackers(
		log, cfg, m, store, infra,
		statsTrackerPubSub, charactersTrackerPubSub, mt,
	)

	if err := infra.platformServices.Init(
		log, cfg, m, mt,
		infra.censusDataProvider, infra.storageOutfitsRepo,
		infra.facilityCache, trackers.charactersTracker,
		trackers.statsTracker,
	); err != nil {
		return nil, err
	}

	newStorageEventsSubscription(log, m, ps2PubSub, trackers)

	trackingPubSub := pubsub.New[tracking.EventType]()

	settingsService := newSettingsService(log, cfg, storePubSub, infra, trackers, trackingPubSub)

	newTrackingSettingsSubscription(log, m, trackingPubSub, trackers, outfitSync)

	loaders := newProviders(
		log, cfg, httpClient, infra,
		trackers.charactersTracker,
	)

	if err := newDiscordModule(
		log, cfg, m, store, storePubSub,
		ps2PubSub, trackingPubSub, charactersTrackerPubSub,
		statsTrackerPubSub, trackers, loaders,
		infra, settingsService,
	); err != nil {
		return nil, err
	}

	return m, nil
}
