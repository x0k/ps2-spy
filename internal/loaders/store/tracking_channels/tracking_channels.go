package trackingchannels

import (
	"context"

	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type TrackingChannelsLoader struct {
	store *sqlite.Storage
}

func New(store *sqlite.Storage) *TrackingChannelsLoader {
	return &TrackingChannelsLoader{
		store: store,
	}
}

func (l *TrackingChannelsLoader) Load(ctx context.Context, char ps2.Character) ([]string, error) {
	return l.store.TrackingChannelIdsForCharacter(ctx, char.Id, char.OutfitTag)
}
