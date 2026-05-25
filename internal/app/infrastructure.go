package app

import (
	"net/http"

	sql_facility_cache "github.com/x0k/ps2-spy/internal/cache/facility/sql"
	census_data_provider "github.com/x0k/ps2-spy/internal/data_providers/census"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/metrics"
	ps2_census_characters_repo "github.com/x0k/ps2-spy/internal/ps2/census_characters_repo"
	ps2_census_outfits_repo "github.com/x0k/ps2-spy/internal/ps2/census_outfits_repo"
	ps2_storage_facility_repo "github.com/x0k/ps2-spy/internal/ps2/storage_facility_repo"
	ps2_storage_outfits_repo "github.com/x0k/ps2-spy/internal/ps2/storage_outfits_repo"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
)

type infrastructureDeps struct {
	censusClient         *census2.Client
	censusDataProvider   *census_data_provider.DataProvider
	censusCharactersRepo *ps2_census_characters_repo.Repository
	censusOutfitsRepo    *ps2_census_outfits_repo.Repository
	storageOutfitsRepo   *ps2_storage_outfits_repo.Repository
	facilityCache        *sql_facility_cache.Cache
	platformServices     *PlatformServices
}

func newInfrastructure(
	log *logger.Logger,
	cfg *Config,
	store *sql_storage.Storage,
	httpClient *http.Client,
	mt *metrics.Metrics,
) (*infrastructureDeps, error) {
	censusClient := census2.NewClient("https://census.daybreakgames.com", cfg.Census.ServiceId, httpClient)

	storageFacilityRepo := ps2_storage_facility_repo.New(store)

	facilityCache := sql_facility_cache.New(
		log.With(sl.Component("facility_cache")),
		storageFacilityRepo,
	)

	platformServices := newPlatformServices()

	censusCharactersRepo := ps2_census_characters_repo.New(
		log.With(sl.Component("census_characters_repo")),
		censusClient,
	)
	censusOutfitsRepo := ps2_census_outfits_repo.New(
		log.With(sl.Component("census_outfits_repo")),
		censusClient,
	)

	storageOutfitsRepo := ps2_storage_outfits_repo.New(store)

	censusDataProvider, err := census_data_provider.New(
		log.With(sl.Component("census_data_provider")),
		censusClient,
	)
	if err != nil {
		return nil, err
	}

	return &infrastructureDeps{
		censusClient:         censusClient,
		censusDataProvider:   censusDataProvider,
		censusCharactersRepo: censusCharactersRepo,
		censusOutfitsRepo:    censusOutfitsRepo,
		storageOutfitsRepo:   storageOutfitsRepo,
		facilityCache:        facilityCache,
		platformServices:     platformServices,
	}, nil
}
