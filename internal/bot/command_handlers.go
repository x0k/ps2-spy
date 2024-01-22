package bot

import (
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/handlers/command/about_command_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/command/alerts_command_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/command/channel_setup_command_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/command/population_command_handler"
	serverpopulation "github.com/x0k/ps2-spy/internal/bot/handlers/command/server_population_command_handler"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func NewCommandHandlers(
	popLoader loaders.KeyedLoader[string, loaders.Loaded[ps2.WorldsPopulation]],
	worldPopLoader loaders.QueriedLoader[loaders.MultiLoaderQuery[ps2.WorldId], loaders.Loaded[ps2.DetailedWorldPopulation]],
	alertsLoader loaders.KeyedLoader[string, loaders.Loaded[ps2.Alerts]],
	worldAlertsLoader loaders.QueriedLoader[loaders.MultiLoaderQuery[ps2.WorldId], loaders.Loaded[ps2.Alerts]],
	settingsLoader loaders.KeyedLoader[[2]string, meta.SubscriptionSettings],
	charNamesLoader loaders.QueriedLoader[channel_setup_command_handler.PlatformQuery, []string],
	outfitTagsLoader loaders.QueriedLoader[channel_setup_command_handler.PlatformQuery, []string],
) map[string]handlers.InteractionHandler {
	return map[string]handlers.InteractionHandler{
		"population":        population_command_handler.New(popLoader),
		"server-population": serverpopulation.New(worldPopLoader),
		"alerts": alerts_command_handler.New(
			alertsLoader,
			worldAlertsLoader,
		),
		"setup": channel_setup_command_handler.New(settingsLoader, charNamesLoader, outfitTagsLoader),
		"about": about_command_handler.New(),
	}
}
