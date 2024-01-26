package outfits_cache

import (
	"context"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
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
	const op = "cache.outfits_cache.StorageCache.Get"
	log := infra.Logger(ctx).With(
		slog.String("platform", string(s.platform)),
		slog.Any("outfit_ids", outfitIds),
	)
	res := make(map[ps2.OutfitId]ps2.Outfit)
	if len(outfitIds) == 0 {
		return res, true
	}
	outfits, err := s.storage.Outfits(ctx, s.platform, outfitIds)
	if err != nil {
		log.Error("failed to get outfits", sl.Err(err))
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
