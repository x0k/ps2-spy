package tracking_storage_tracking_repo

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/discord"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
)

type Repository struct {
	storage storage.Storage
}

func New(storage storage.Storage) *Repository {
	return &Repository{
		storage: storage,
	}
}

func (r *Repository) PlatformsByChannelId(
	ctx context.Context, channelId discord.ChannelId,
) ([]ps2_platforms.Platform, error) {
	data, err := r.storage.Queries().ListChannelTrackablePlatforms(
		ctx, string(channelId),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list channel %q trackable platforms: %w", channelId, err)
	}
	platforms := make([]ps2_platforms.Platform, 0, len(data))
	for _, p := range data {
		platforms = append(platforms, ps2_platforms.Platform(p))
	}
	return platforms, nil
}
