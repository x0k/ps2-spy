package sql_trackable_character_ids_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
)

func New(storage *sql_storage.Storage, platform ps2_platforms.Platform) func(context.Context) ([]ps2.CharacterId, error) {
	return func(ctx context.Context) ([]ps2.CharacterId, error) {
		return storage.AllTrackableCharacterIdsWithDuplicationsForPlatform(ctx, platform)
	}
}
