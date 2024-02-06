package bot

import (
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/handlers/submit/channel_setup_submit_handler"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/loaders/character_ids_loader"
	"github.com/x0k/ps2-spy/internal/loaders/character_names_loader"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_ids_loader"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_tags_loader"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/savers/subscription_settings_saver"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

func newSetupSubmitHandler(
	platform platforms.Platform,
	sqlStorage *sqlite.Storage,
	censusClient *census2.Client,
	subscriptionSettingsLoader loaders.KeyedLoader[meta.SettingsQuery, meta.SubscriptionSettings],
) handlers.InteractionHandler {
	ns := platforms.PlatformNamespace(platform)
	return channel_setup_submit_handler.New(
		character_ids_loader.NewCensus(censusClient, ns),
		character_names_loader.NewCensus(censusClient, ns),
		outfit_ids_loader.NewCensus(censusClient, ns),
		outfit_tags_loader.NewCensus(censusClient, ns),
		subscription_settings_saver.New(
			sqlStorage,
			subscriptionSettingsLoader,
			platform,
		),
	)
}

func newSubmitHandlers(
	sqlStorage *sqlite.Storage,
	censusClient *census2.Client,
	subscriptionSettingsLoader loaders.KeyedLoader[meta.SettingsQuery, meta.SubscriptionSettings],
) map[string]handlers.InteractionHandler {
	return map[string]handlers.InteractionHandler{
		handlers.CHANNEL_SETUP_PC_MODAL: newSetupSubmitHandler(
			platforms.PC,
			sqlStorage,
			censusClient,
			subscriptionSettingsLoader,
		),
		handlers.CHANNEL_SETUP_PS4_EU_MODAL: newSetupSubmitHandler(
			platforms.PS4_EU,
			sqlStorage,
			censusClient,
			subscriptionSettingsLoader,
		),
		handlers.CHANNEL_SETUP_PS4_US_MODAL: newSetupSubmitHandler(
			platforms.PS4_US,
			sqlStorage,
			censusClient,
			subscriptionSettingsLoader,
		),
	}
}
