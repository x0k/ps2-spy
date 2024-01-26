package outfit_cache

import (
	"context"
	"errors"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
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
	const op = "cache.outfit_cache.StorageCache.Get"
	log := infra.Logger(ctx).With(
		slog.String("platform", string(s.platform)),
		slog.String("outfit_id", string(id)),
	)
	outfit, err := s.storage.Outfit(ctx, s.platform, id)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		log.Error("failed to get outfit", sl.Err(err))
	}
	return outfit, err == nil
}

func (s *StorageCache) Add(ctx context.Context, outfitId ps2.OutfitId, outfit ps2.Outfit) error {
	return s.storage.SaveOutfit(ctx, outfit)
}
