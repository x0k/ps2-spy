package trackable_character_ids_loader

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

func (s *StoreLoader) Load(ctx context.Context) ([]string, error) {
	return s.storage.AllTrackableCharactersForPlatform(ctx, s.platform)
}
