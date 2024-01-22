package outfit_member_ids_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type StoreLoader struct {
	storage  *sqlite.Storage
	platform string
}

func NewStorage(storage *sqlite.Storage, platform string) *StoreLoader {
	return &StoreLoader{
		storage:  storage,
		platform: platform,
	}
}

func (l *StoreLoader) Load(ctx context.Context, outfitId string) ([]string, error) {
	return l.storage.OutfitMembers(ctx, l.platform, outfitId)
}
