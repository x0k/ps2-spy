package sql_character_tracking_channels_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/ps2"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
)

func New(
	storage *sql_storage.Storage,
) func(context.Context, ps2.Character) ([]discord.ChannelId, error) {
	return func(ctx context.Context, c ps2.Character) ([]discord.ChannelId, error) {
		return storage.TrackingChannelIdsForCharacter(ctx, c.Platform, c.Id, c.OutfitId)
	}
}
