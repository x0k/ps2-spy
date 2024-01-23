package character_tracking_channels_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type CharacterTrackingChannelsLoader struct {
	store *sqlite.Storage
}

func New(store *sqlite.Storage) *CharacterTrackingChannelsLoader {
	return &CharacterTrackingChannelsLoader{
		store: store,
	}
}

func (l *CharacterTrackingChannelsLoader) Load(ctx context.Context, char ps2.Character) ([]string, error) {
	return l.store.TrackingChannelIdsForCharacter(ctx, char.Platform, char.Id, char.OutfitId)
}
