package bot

import (
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/handlers/alerts"
	"github.com/x0k/ps2-spy/internal/bot/handlers/population"
	serverpopulation "github.com/x0k/ps2-spy/internal/bot/handlers/server_population"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func NewHandlers(
	popLoader loaders.KeyedLoader[string, ps2.WorldsPopulation],
	worldPopLoader loaders.QueriedLoader[loaders.MultiLoaderQuery[ps2.WorldId], ps2.DetailedWorldPopulation],
	alertsLoader loaders.KeyedLoader[string, ps2.Alerts],
	worldAlertsLoader loaders.QueriedLoader[loaders.MultiLoaderQuery[ps2.WorldId], ps2.Alerts],
) map[string]handlers.InteractionHandler {
	return map[string]handlers.InteractionHandler{
		"population":        population.New(popLoader),
		"server-population": serverpopulation.New(worldPopLoader),
		"alerts": alerts.New(
			alertsLoader,
			worldAlertsLoader,
		),
	}
}
