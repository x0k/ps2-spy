package facility_cache

import (
	"context"
	"errors"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type StorageCache struct {
	log     *logger.Logger
	storage *sqlite.Storage
}

func NewStorage(log *logger.Logger, storage *sqlite.Storage) *StorageCache {
	return &StorageCache{
		log:     log.With(slog.String("component", "cache.facility_cache.StorageCache")),
		storage: storage,
	}
}

func (s *StorageCache) Get(ctx context.Context, facilityId ps2.FacilityId) (ps2.Facility, bool) {
	facility, err := s.storage.Facility(ctx, facilityId)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		s.log.Error(ctx, "failed to get facility", slog.String("facility_id", string(facilityId)), sl.Err(err))
	}
	return facility, err == nil
}

func (s *StorageCache) Add(ctx context.Context, facilityId ps2.FacilityId, facility ps2.Facility) error {
	return s.storage.SaveFacility(ctx, facility)
}
