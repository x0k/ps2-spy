package facility_cache

import (
	"context"

	"github.com/x0k/ps2-spy/internal/ps2"
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
	facility, err := s.storage.Facility(ctx, facilityId)
	return facility, err == nil
}

func (s *StorageCache) Add(ctx context.Context, facilityId ps2.FacilityId, facility ps2.Facility) error {
	return s.storage.SaveFacility(ctx, facility)
}
