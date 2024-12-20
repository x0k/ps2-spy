package discord_commands

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/expirable_state_container"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
)

type commands struct {
	commands                 []*discord.Command
	populationLoader         *populationLoader
	worldPopulationLoader    *worldPopulationLoader
	alertsLoader             *alertsLoader
	createTaskStateContainer *expirable_state_container.ExpirableStateContainer[
		discord.ChannelAndUserIds,
		discord.CreateStatsTrackerTaskState,
	]
}

func New(
	log *logger.Logger,
	messages *discord_messages.Messages,
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
	settingsLoader loader.Keyed[discord.SettingsQuery, discord.TrackingSettings],
	characterNamesLoader loader.Queried[discord.PlatformQuery[[]ps2.CharacterId], []string],
	characterIdsLoader loader.Queried[discord.PlatformQuery[[]string], []ps2.CharacterId],
	outfitTagsLoader loader.Queried[discord.PlatformQuery[[]ps2.OutfitId], []string],
	outfitIdsLoader loader.Queried[discord.PlatformQuery[[]string], []ps2.OutfitId],
	channelTrackingSettingsSaver ChannelTrackingSettingsSaver,
	statsTracker *stats_tracker.StatsTracker,
	channelLoader ChannelLoader,
	channelLanguageSaver ChannelLanguageSaver,
	channelCharacterNotificationsSaver ChannelCharacterNotificationsSaver,
	channelOutfitNotificationsSaver ChannelOutfitNotificationsSaver,
	channelTitleUpdatesSaver ChannelTitleUpdatesSaver,
	channelDefaultTimezoneSaver ChannelDefaultTimezoneSaver,
	channelStatsTrackerTasksLoader ChannelStatsTrackerTasksLoader,
	statsTrackerTaskCreator ChannelStatsTrackerTaskCreator,
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
	createTaskStateContainer := expirable_state_container.New[discord.ChannelAndUserIds, discord.CreateStatsTrackerTaskState](
		10 * time.Minute,
	)
	return &commands{
		populationLoader:         populationLoader,
		worldPopulationLoader:    worldPopulationLoader,
		alertsLoader:             alertsLoader,
		createTaskStateContainer: createTaskStateContainer,
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
			NewTracking(
				messages,
				settingsLoader,
				characterNamesLoader,
				characterIdsLoader,
				outfitTagsLoader,
				outfitIdsLoader,
				channelTrackingSettingsSaver,
			),
			NewChannelSettings(
				messages,
				channelLoader,
				channelLanguageSaver,
				channelCharacterNotificationsSaver,
				channelOutfitNotificationsSaver,
				channelTitleUpdatesSaver,
				channelDefaultTimezoneSaver,
			),
			NewStatsTracker(
				log.With(sl.Component("stats_tracker_command")),
				messages,
				statsTracker,
				channelStatsTrackerTasksLoader,
				channelLoader,
				createTaskStateContainer,
				statsTrackerTaskCreator,
			),
		},
	}
}

func (c *commands) Commands() []*discord.Command {
	return c.commands
}

func (c *commands) Start(ctx context.Context) error {
	wg := sync.WaitGroup{}
	wg.Add(4)
	go func() {
		defer wg.Done()
		c.createTaskStateContainer.Start(ctx)
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
