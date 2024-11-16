package sql_outfit_sync_at_loader

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/shared"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
)

func New(
	storage *sql_storage.Storage,
	platform ps2_platforms.Platform,
) func(context.Context, ps2.OutfitId) (time.Time, error) {
	return func(ctx context.Context, c ps2.OutfitId) (time.Time, error) {
		const op = "sql_outfit_sync_at_loader.Load"
		syncAt, err := storage.OutfitSynchronizedAt(ctx, platform, c)
		if errors.Is(err, shared.ErrNotFound) {
			return time.Time{}, fmt.Errorf("%s: %w", op, loader.ErrNotFound)
		}
		if err != nil {
			return time.Time{}, fmt.Errorf("%s: %w", op, err)
		}
		return syncAt, nil
	}
}
