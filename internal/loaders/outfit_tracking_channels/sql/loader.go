package sql_outfit_tracking_channels_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
)

func New(
	storage *sql_storage.Storage,
	platform ps2_platforms.Platform,
) func(context.Context, ps2.OutfitId) ([]discord.ChannelId, error) {
	return func(ctx context.Context, oi ps2.OutfitId) ([]discord.ChannelId, error) {
		return storage.TrackingChannelIdsForOutfit(ctx, platform, oi)
	}
}
