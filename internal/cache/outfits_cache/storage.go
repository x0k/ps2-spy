package outfits_cache

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

func (s *StorageCache) Get(ctx context.Context, outfitIds []ps2.OutfitId) (map[ps2.OutfitId]ps2.Outfit, bool) {
	res := make(map[ps2.OutfitId]ps2.Outfit)
	outfits, err := s.storage.Outfits(ctx, s.platform, outfitIds)
	if err != nil {
		return res, false
	}
	for _, outfit := range outfits {
		res[outfit.Id] = outfit
	}
	return res, len(outfits) == len(outfitIds)
}

func (s *StorageCache) Add(ctx context.Context, outfits map[ps2.OutfitId]ps2.Outfit) error {
	return s.storage.Begin(ctx, 0, func(tx *sqlite.Storage) error {
		for _, outfit := range outfits {
			if err := tx.SaveOutfit(ctx, outfit); err != nil {
				return err
			}
		}
		return nil
	})
}
