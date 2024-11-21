package sql_facility_cache

import (
	"context"
	"errors"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/shared"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
)

type Cache struct {
	log     *logger.Logger
	storage *sql_storage.Storage
}

func New(log *logger.Logger, storage *sql_storage.Storage) *Cache {
	return &Cache{
		log:     log,
		storage: storage,
	}
}

func (s *Cache) Get(ctx context.Context, facilityId ps2.FacilityId) (ps2.Facility, bool) {
	facility, err := s.storage.Facility(ctx, facilityId)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		s.log.Error(ctx, "failed to get facility", slog.String("facility_id", string(facilityId)), sl.Err(err))
	}
	return facility, err == nil
}

func (s *Cache) Add(ctx context.Context, facilityId ps2.FacilityId, facility ps2.Facility) error {
	return s.storage.SaveFacility(ctx, facility)
}
