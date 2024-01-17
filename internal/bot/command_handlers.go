package bot

import (
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/handlers/command/alerts"
	channelsetup "github.com/x0k/ps2-spy/internal/bot/handlers/command/channel_setup"
	"github.com/x0k/ps2-spy/internal/bot/handlers/command/population"
	serverpopulation "github.com/x0k/ps2-spy/internal/bot/handlers/command/server_population"
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
	charNamesLoader loaders.QueriedLoader[[]string, []string],
) map[string]handlers.InteractionHandler {
	return map[string]handlers.InteractionHandler{
		"population":        population.New(popLoader),
		"server-population": serverpopulation.New(worldPopLoader),
		"alerts": alerts.New(
			alertsLoader,
			worldAlertsLoader,
		),
		"setup": channelsetup.New(settingsLoader, charNamesLoader),
	}
}
