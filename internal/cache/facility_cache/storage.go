package facility_cache

import (
	"context"
	"errors"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type StorageCache struct {
	storage *sqlite.Storage
}

func NewStorage(storage *sqlite.Storage) *StorageCache {
	return &StorageCache{
		storage: storage,
	}
}

func (s *StorageCache) Get(ctx context.Context, facilityId ps2.FacilityId) (ps2.Facility, bool) {
	const op = "cache.facility_cache.StorageCache.Get"
	log := infra.Logger(ctx).With(
		slog.String("facility_id", string(facilityId)),
	)
	facility, err := s.storage.Facility(ctx, facilityId)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		log.Error("failed to get facility", sl.Err(err))
	}
	return facility, err == nil
}

func (s *StorageCache) Add(ctx context.Context, facilityId ps2.FacilityId, facility ps2.Facility) error {
	return s.storage.SaveFacility(ctx, facility)
}
