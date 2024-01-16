package trackingchannels

import (
	"context"

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

func (l *TrackingChannelsLoader) Load(ctx context.Context, key [2]string) ([]string, error) {
	return l.store.TrackingChannelIdsForCharacter(ctx, key[0], key[1])
}
