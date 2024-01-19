package outfit_trackers_count_loader

import (
	"context"

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

func (s *StorageLoader) Load(ctx context.Context, outfitTag string) (int, error) {
	return s.storage.CountOutfitTrackingChannels(ctx, s.platform, outfitTag)
}
