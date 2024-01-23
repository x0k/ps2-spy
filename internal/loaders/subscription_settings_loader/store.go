package subscription_settings_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type SubscriptionSettingsLoader struct {
	storage *sqlite.Storage
}

func New(storage *sqlite.Storage) *SubscriptionSettingsLoader {
	return &SubscriptionSettingsLoader{
		storage: storage,
	}
}

func (l *SubscriptionSettingsLoader) Load(ctx context.Context, query meta.SettingsQuery) (meta.SubscriptionSettings, error) {
	outfits, err := l.storage.TrackingOutfitsForPlatform(ctx, query.ChannelId, query.Platform)
	if err != nil {
		return meta.SubscriptionSettings{}, err
	}
	characters, err := l.storage.TrackingCharactersForPlatform(ctx, query.ChannelId, query.Platform)
	if err != nil {
		return meta.SubscriptionSettings{}, err
	}
	return meta.SubscriptionSettings{
		Outfits:    outfits,
		Characters: characters,
	}, nil
}
