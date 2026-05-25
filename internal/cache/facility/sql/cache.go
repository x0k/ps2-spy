package sql_facility_cache

import (
	"context"
	"errors"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_storage_facility_repo "github.com/x0k/ps2-spy/internal/ps2/storage_facility_repo"
	"github.com/x0k/ps2-spy/internal/shared"
)

type Cache struct {
	log  *logger.Logger
	repo *ps2_storage_facility_repo.Repository
}

func New(log *logger.Logger, repo *ps2_storage_facility_repo.Repository) *Cache {
	return &Cache{
		log:  log,
		repo: repo,
	}
}

func (s *Cache) Get(ctx context.Context, facilityId ps2.FacilityId) (ps2.Facility, bool) {
	facility, err := s.repo.Facility(ctx, facilityId)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		s.log.Error(ctx, "failed to get facility", slog.String("facility_id", string(facilityId)), sl.Err(err))
	}
	return facility, err == nil
}

func (s *Cache) Add(ctx context.Context, facilityId ps2.FacilityId, facility ps2.Facility) error {
	return s.repo.SaveFacility(ctx, facility)
}
