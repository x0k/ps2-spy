package discord_commands

import (
	"context"
	"slices"
	"sync"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type commands struct {
	name                  string
	commands              []*discord.Command
	populationLoader      *populationLoader
	worldPopulationLoader *worldPopulationLoader
	alertsLoader          *alertsLoader
}

func New(
	name string,
	log *logger.Logger,
	messages discord.LocalizedMessages,
	populationLoaders map[string]loader.Simple[meta.Loaded[ps2.WorldsPopulation]],
	populationLoadersPriority []string,
	worldPopulationLoaders map[string]loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]],
	worldPopulationLoadersPriority []string,
	worldTerritoryControlLoader loader.Keyed[ps2.WorldId, meta.Loaded[ps2.WorldTerritoryControl]],
	alertsLoaders map[string]loader.Simple[meta.Loaded[ps2.Alerts]],
	alertsLoadersPriority []string,
	onlineTrackableEntitiesLoader loader.Keyed[discord.SettingsQuery, discord.TrackableEntities[
		map[ps2.OutfitId][]ps2.Character,
		[]ps2.Character,
	]],
	outfitsLoader loader.Queried[discord.PlatformQuery[[]ps2.OutfitId], map[ps2.OutfitId]ps2.Outfit],
	settingsLoader loader.Keyed[discord.SettingsQuery, discord.SubscriptionSettings],
	namesLoader loader.Queried[discord.PlatformQuery[[]ps2.CharacterId], []string],
	outfitTagsLoader loader.Queried[discord.PlatformQuery[[]ps2.OutfitId], []string],
) *commands {
	populationLoader := newPopulationLoader(
		log.With(sl.Component("population_loader")),
		populationLoaders,
		populationLoadersPriority,
	)
	worldPopulationLoader := newWorldPopulationLoader(
		log.With(sl.Component("world_population_loader")),
		worldPopulationLoaders,
		worldPopulationLoadersPriority,
	)
	alertsLoader := newAlertsLoader(
		log.With(sl.Component("alerts_loader")),
		alertsLoaders,
		alertsLoadersPriority,
	)
	return &commands{
		name:                  name,
		populationLoader:      populationLoader,
		worldPopulationLoader: worldPopulationLoader,
		commands: []*discord.Command{
			NewAbout(messages),
			NewPopulation(
				log.With(sl.Component("population_command")),
				messages,
				populationLoader.load,
				slices.Values(populationLoadersPriority),
				worldPopulationLoader.load,
				slices.Values(worldPopulationLoadersPriority),
			),
			NewTerritories(
				messages,
				worldTerritoryControlLoader,
			),
			NewAlerts(
				log.With(sl.Component("alerts_command")),
				messages,
				slices.Values(alertsLoadersPriority),
				alertsLoader.load,
				func(ctx context.Context, q query[ps2.WorldId]) (meta.Loaded[ps2.Alerts], error) {
					loaded, err := alertsLoader.load(ctx, q.Provider)
					if err != nil {
						return meta.Loaded[ps2.Alerts]{}, err
					}
					worldAlerts := make(ps2.Alerts, 0, len(loaded.Value))
					for _, alert := range loaded.Value {
						if alert.WorldId == q.Key {
							worldAlerts = append(worldAlerts, alert)
						}
					}
					loaded.Value = worldAlerts
					return loaded, nil
				},
			),
			NewOnline(
				messages,
				onlineTrackableEntitiesLoader,
				outfitsLoader,
			),
			NewSubscription(
				messages,
				settingsLoader,
				namesLoader,
				outfitTagsLoader,
			),
		},
	}
}

func (c *commands) Name() string {
	return c.name
}

func (c *commands) Commands() []*discord.Command {
	return c.commands
}

func (c *commands) Start(ctx context.Context) error {
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		c.worldPopulationLoader.Start(ctx)
	}()
	go func() {
		defer wg.Done()
		c.populationLoader.Start(ctx)
	}()
	go func() {
		defer wg.Done()
		c.alertsLoader.Start(ctx)
	}()
	<-ctx.Done()
	wg.Wait()
	return nil
}
