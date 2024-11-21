package sql_outfits_cache

import (
	"context"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
)

type Cache struct {
	log      *logger.Logger
	storage  *sql_storage.Storage
	platform ps2_platforms.Platform
}

func New(log *logger.Logger, storage *sql_storage.Storage, platform ps2_platforms.Platform) *Cache {
	return &Cache{
		log:      log,
		storage:  storage,
		platform: platform,
	}
}

func (s *Cache) Get(ctx context.Context, outfitIds []ps2.OutfitId) (map[ps2.OutfitId]ps2.Outfit, bool) {
	res := make(map[ps2.OutfitId]ps2.Outfit)
	if len(outfitIds) == 0 {
		return res, true
	}
	outfits, err := s.storage.Outfits(ctx, s.platform, outfitIds)
	if err != nil {
		s.log.Error(ctx, "failed to get outfits", slog.Any("outfit_ids", outfitIds), sl.Err(err))
		return res, false
	}
	for _, outfit := range outfits {
		res[outfit.Id] = outfit
	}
	return res, len(outfits) == len(outfitIds)
}

func (s *Cache) Add(ctx context.Context, outfits map[ps2.OutfitId]ps2.Outfit) error {
	return s.storage.Begin(ctx, 0, func(tx *sql_storage.Storage) error {
		for _, outfit := range outfits {
			if err := tx.SaveOutfit(ctx, outfit); err != nil {
				return err
			}
		}
		return nil
	})
}
