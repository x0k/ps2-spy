package app

import (
	"net/http"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	fisu_data_provider "github.com/x0k/ps2-spy/internal/data_providers/fisu"
	honu_data_provider "github.com/x0k/ps2-spy/internal/data_providers/honu"
	ps2alerts_data_provider "github.com/x0k/ps2-spy/internal/data_providers/ps2alerts"
	ps2live_data_provider "github.com/x0k/ps2-spy/internal/data_providers/ps2live"
	ps2spy_data_provider "github.com/x0k/ps2-spy/internal/data_providers/ps2spy"
	saerro_data_provider "github.com/x0k/ps2-spy/internal/data_providers/saerro"
	sanctuary_data_provider "github.com/x0k/ps2-spy/internal/data_providers/sanctuary"
	voidwell_data_provider "github.com/x0k/ps2-spy/internal/data_providers/voidwell"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/fisu"
	"github.com/x0k/ps2-spy/internal/lib/honu"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/ps2alerts"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/population"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/lib/voidwell"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type providerDeps struct {
	ps2spy    *ps2spy_data_provider.DataProvider
	honu      *honu_data_provider.DataProvider
	ps2alerts *ps2alerts_data_provider.DataProvider
	voidwell  *voidwell_data_provider.DataProvider
	fisu      *fisu_data_provider.DataProvider
	ps2live   *ps2live_data_provider.DataProvider
	sanctuary *sanctuary_data_provider.DataProvider
	saerro    *saerro_data_provider.DataProvider
}

type loaderDeps struct {
	population      map[string]loader.Simple[meta.Loaded[ps2.WorldsPopulation]]
	worldPopulation map[string]loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]]
	alerts          map[string]loader.Simple[meta.Loaded[ps2.Alerts]]
}

func newProviders(
	log *logger.Logger,
	cfg *Config,
	httpClient *http.Client,
	censusClient *census2.Client,
	charactersTracker *characters_tracker.Tracker,
	platformServices *PlatformServices,
) (*providerDeps, *loaderDeps) {
	ps2spy := ps2spy_data_provider.New(
		log,
		cfg.AppName,
		charactersTracker,
		platformServices,
	)

	honu := honu_data_provider.New(
		honu.NewClient("https://wt.honu.pw", httpClient),
	)
	ps2alertsProvider := ps2alerts_data_provider.New(
		ps2alerts.NewClient("https://api.ps2alerts.com", httpClient),
	)
	voidwell := voidwell_data_provider.New(
		voidwell.NewClient("https://api.voidwell.com", httpClient),
	)
	fisu := fisu_data_provider.New(
		fisu.NewClient("https://ps2.fisu.pw", httpClient),
	)
	ps2live := ps2live_data_provider.New(
		population.NewClient("https://agg.ps2.live", httpClient),
	)
	sanctuary := sanctuary_data_provider.New(
		census2.NewClient("https://census.lithafalcon.cc", cfg.Census.ServiceId, httpClient),
	)
	saerro := saerro_data_provider.New(
		saerro.NewClient("https://saerro.ps2.live", httpClient),
	)

	providers := &providerDeps{
		ps2spy:    ps2spy,
		honu:      honu,
		ps2alerts: ps2alertsProvider,
		voidwell:  voidwell,
		fisu:      fisu,
		ps2live:   ps2live,
		sanctuary: sanctuary,
		saerro:    saerro,
	}

	loaders := &loaderDeps{
		population: map[string]loader.Simple[meta.Loaded[ps2.WorldsPopulation]]{
			"spy":       ps2spy.Population,
			"honu":      honu.Population,
			"ps2live":   ps2live.Population,
			"saerro":    saerro.Population,
			"fisu":      fisu.Population,
			"sanctuary": sanctuary.Population,
			"voidwell":  voidwell.Population,
		},
		worldPopulation: map[string]loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]]{
			"spy":      ps2spy.WorldPopulation,
			"honu":     honu.WorldPopulation,
			"saerro":   saerro.WorldPopulation,
			"voidwell": voidwell.WorldPopulation,
		},
		alerts: map[string]loader.Simple[meta.Loaded[ps2.Alerts]]{
			"spy":      ps2spy.Alerts,
			"honu":     honu.Alerts,
			"census":   nil, // filled after creation
			"voidwell": voidwell.Alerts,
		},
	}

	return providers, loaders
}
