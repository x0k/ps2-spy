package bot

import (
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/handlers/command/about_command_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/command/alerts_command_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/command/channel_setup_command_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/command/online_command_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/command/population_command_handler"
	serverpopulation "github.com/x0k/ps2-spy/internal/bot/handlers/command/server_population_command_handler"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func newCommandHandlers(
	log *logger.Logger,
	popLoader loaders.KeyedLoader[string, loaders.Loaded[ps2.WorldsPopulation]],
	worldPopLoader loaders.QueriedLoader[loaders.MultiLoaderQuery[ps2.WorldId], loaders.Loaded[ps2.DetailedWorldPopulation]],
	alertsLoader loaders.KeyedLoader[string, loaders.Loaded[ps2.Alerts]],
	worldAlertsLoader loaders.QueriedLoader[loaders.MultiLoaderQuery[ps2.WorldId], loaders.Loaded[ps2.Alerts]],
	settingsLoader loaders.KeyedLoader[meta.SettingsQuery, meta.SubscriptionSettings],
	charNamesLoader loaders.QueriedLoader[meta.PlatformQuery[ps2.CharacterId], []string],
	outfitTagsLoader loaders.QueriedLoader[meta.PlatformQuery[ps2.OutfitId], []string],
	trackableOnlineEntitiesLoader loaders.KeyedLoader[meta.SettingsQuery, meta.TrackableEntities[
		map[ps2.OutfitId][]ps2.Character,
		[]ps2.Character,
	]],
	outfitsLoader loaders.QueriedLoader[meta.PlatformQuery[ps2.OutfitId], map[ps2.OutfitId]ps2.Outfit],
) map[string]handlers.InteractionHandler {
	return map[string]handlers.InteractionHandler{
		"population":        population_command_handler.New(log, popLoader),
		"server-population": serverpopulation.New(log, worldPopLoader),
		"alerts": alerts_command_handler.New(
			log,
			alertsLoader,
			worldAlertsLoader,
		),
		"setup":  channel_setup_command_handler.New(settingsLoader, charNamesLoader, outfitTagsLoader),
		"online": online_command_handler.New(trackableOnlineEntitiesLoader, outfitsLoader),
		"about":  about_command_handler.New(),
	}
}
