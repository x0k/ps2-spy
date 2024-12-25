package tracking_settings_repo

import (
	"context"

	"github.com/x0k/ps2-spy/internal/discord"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking"
)

type Repository struct {
	storage storage.Storage
}

func New(storage storage.Storage) *Repository {
	return &Repository{
		storage: storage,
	}
}

func (r *Repository) Transaction(ctx context.Context, f func(r *Repository) error) error {
	return r.storage.Transaction(ctx, func(s storage.Storage) error {
		return f(&Repository{
			storage: s,
		})
	})
}

func (r *Repository) Settings(
	ctx context.Context,
	channelId discord.ChannelId,
	platform ps2_platforms.Platform,
) (tracking.Settings, error) {

}
