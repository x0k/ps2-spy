package outfit_sync_at_loader

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type StorageLoader struct {
	storage  *sqlite.Storage
	platform string
}

func NewStorage(storage *sqlite.Storage, platform string) *StorageLoader {
	return &StorageLoader{
		storage:  storage,
		platform: platform,
	}
}

func (s *StorageLoader) Load(ctx context.Context, key string) (time.Time, error) {
	const op = "loaders.outfit_sync_at_loader.StorageLoader.Load"
	syncAt, err := s.storage.OutfitSynchronizedAt(ctx, s.platform, key)
	if errors.Is(err, storage.ErrNotFound) {
		return time.Time{}, fmt.Errorf("%s: %w", op, loaders.ErrNotFound)
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("%s: %w", op, err)
	}
	return syncAt, nil
}
