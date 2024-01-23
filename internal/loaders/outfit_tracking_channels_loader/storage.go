package outfit_tracking_channels_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type StorageLoader struct {
	storage  *sqlite.Storage
	platform platforms.Platform
}

func NewStorage(storage *sqlite.Storage, platform platforms.Platform) *StorageLoader {
	return &StorageLoader{
		storage:  storage,
		platform: platform,
	}
}

func (s *StorageLoader) Load(ctx context.Context, outfitId ps2.OutfitId) ([]meta.ChannelId, error) {
	return s.storage.TrackingChannelsIdsForOutfit(ctx, s.platform, outfitId)
}
