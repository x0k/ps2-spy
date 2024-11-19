package sql_subscription_settings_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
)

func New(
	storage *sql_storage.Storage,
) loader.Keyed[discord.SettingsQuery, discord.SubscriptionSettings] {
	return func(ctx context.Context, query discord.SettingsQuery) (discord.SubscriptionSettings, error) {
		outfits, err := storage.TrackingOutfitIdsForPlatform(ctx, query.ChannelId, query.Platform)
		if err != nil {
			return discord.SubscriptionSettings{}, err
		}
		characters, err := storage.TrackingCharacterIdsForPlatform(ctx, query.ChannelId, query.Platform)
		if err != nil {
			return discord.SubscriptionSettings{}, err
		}
		return discord.SubscriptionSettings{
			Outfits:    outfits,
			Characters: characters,
		}, nil
	}
}
