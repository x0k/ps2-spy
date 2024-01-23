package outfit_cache

import (
	"context"

	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type StorageCache struct {
	storage  *sqlite.Storage
	platform platforms.Platform
}

func NewStorage(storage *sqlite.Storage, platform platforms.Platform) *StorageCache {
	return &StorageCache{
		storage:  storage,
		platform: platform,
	}
}

func (s *StorageCache) Get(ctx context.Context, id ps2.OutfitId) (ps2.Outfit, bool) {
	outfit, err := s.storage.Outfit(ctx, s.platform, id)
	return outfit, err == nil
}

func (s *StorageCache) Add(ctx context.Context, outfitId ps2.OutfitId, outfit ps2.Outfit) error {
	return s.storage.SaveOutfit(ctx, outfit)
}
