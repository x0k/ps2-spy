package subscriptionsettings

import (
	"context"

	channelsetup "github.com/x0k/ps2-spy/internal/bot/handlers/command/channel_setup"
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

func (l *SubscriptionSettingsLoader) Load(ctx context.Context, key [2]string) (channelsetup.SubscriptionSettings, error) {
	outfits, err := l.storage.TrackingOutfitsForPlatform(ctx, key[0], key[1])
	if err != nil {
		return channelsetup.SubscriptionSettings{}, err
	}
	characters, err := l.storage.TrackingCharactersForPlatform(ctx, key[0], key[1])
	if err != nil {
		return channelsetup.SubscriptionSettings{}, err
	}
	return channelsetup.SubscriptionSettings{
		Outfits:    outfits,
		Characters: characters,
	}, nil
}
