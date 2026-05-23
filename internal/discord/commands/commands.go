package discord_commands

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
)

type Commands struct {
	commands               []*discord.Command
	populationLoader       *populationLoader
	worldPopulationLoader  *worldPopulationLoader
	alertsLoader           *alertsLoader
	taskFormStateContainer *containers.ExpirableState[
		discord.ChannelAndUserIds,
		discord.FormState[stats_tracker.CreateOrUpdateTask],
	]
}

func New(
	log *logger.Logger,
	messages *discord_messages.Messages,
	populationProviders PopulationProviders,
	worldPopulationProviders WorldPopulationProviders,
	worldTerritoryControlLoader loader.Keyed[ps2.WorldId, meta.Loaded[ps2.WorldTerritoryControl]],
	alertsProviders AlertsProviders,
	trackingSettingsDataLoader TrackingSettingsDataLoader,
	outfitsLoader OutfitsLoader,
	trackingSettingsLoader TrackingSettingsLoader,
	trackingSettingsUpdater TrackingSettingsUpdater,
	statsTracker *stats_tracker.StatsTracker,
	channelStore ChannelStore,
	statsTaskStore StatsTaskStore,
) *Commands {
	populationLoader := newPopulationLoader(
		log.With(sl.Component("population_loader")),
		populationProviders.Loaders,
		populationProviders.Priority,
	)
	worldPopulationLoader := newWorldPopulationLoader(
		log.With(sl.Component("world_population_loader")),
		worldPopulationProviders.Loaders,
		worldPopulationProviders.Priority,
	)
	alertsLoader := newAlertsLoader(
		log.With(sl.Component("alerts_loader")),
		alertsProviders.Loaders,
		alertsProviders.Priority,
	)
	taskFormStateContainer := containers.NewExpirableState[
		discord.ChannelAndUserIds,
		discord.FormState[stats_tracker.CreateOrUpdateTask],
	](10 * time.Minute)
	return &Commands{
		populationLoader:       populationLoader,
		worldPopulationLoader:  worldPopulationLoader,
		alertsLoader:           alertsLoader,
		taskFormStateContainer: taskFormStateContainer,
		commands: []*discord.Command{
			NewAbout(messages),
			NewPopulation(
				log.With(sl.Component("population_command")),
				messages,
				populationLoader.load,
				slices.Values(populationProviders.Priority),
				worldPopulationLoader.load,
				slices.Values(worldPopulationProviders.Priority),
			),
			NewTerritories(
				messages,
				worldTerritoryControlLoader,
			),
			NewAlerts(
				log.With(sl.Component("alerts_command")),
				messages,
				slices.Values(alertsProviders.Priority),
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
				trackingSettingsDataLoader,
				outfitsLoader,
			),
			NewTracking(
				messages,
				trackingSettingsLoader,
				trackingSettingsUpdater,
			),
			NewChannelSettings(
				messages,
				channelStore,
			),
			NewStatsTracker(
				log.With(sl.Component("stats_tracker_command")),
				messages,
				statsTracker,
				statsTaskStore,
				channelStore.Loader,
				taskFormStateContainer,
			),
		},
	}
}

func (c *Commands) Commands() []*discord.Command {
	return c.commands
}

func (c *Commands) Start(ctx context.Context) error {
	wg := sync.WaitGroup{}
	wg.Add(4)
	go func() {
		defer wg.Done()
		c.taskFormStateContainer.Start(ctx)
	}()
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
