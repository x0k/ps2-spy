package outfit_member_ids_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type StoreLoader struct {
	storage  *sqlite.Storage
	platform platforms.Platform
}

func NewStorage(storage *sqlite.Storage, platform platforms.Platform) *StoreLoader {
	return &StoreLoader{
		storage:  storage,
		platform: platform,
	}
}

func (l *StoreLoader) Load(ctx context.Context, outfitId ps2.OutfitId) ([]ps2.CharacterId, error) {
	return l.storage.OutfitMembers(ctx, l.platform, outfitId)
}
