package outfit_members_ids_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type StoreLoader struct {
	storage *sqlite.Storage
}

func NewStorage(storage *sqlite.Storage) *StoreLoader {
	return &StoreLoader{
		storage: storage,
	}
}

func (l *StoreLoader) Load(ctx context.Context, outfitTag string) ([]string, error) {
	return l.storage.OutfitMembers(ctx, outfitTag)
}
