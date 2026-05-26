package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_commands "github.com/x0k/ps2-spy/internal/discord/commands"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/meta"
	discord_module "github.com/x0k/ps2-spy/internal/modules/discord"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/shared"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	stats_tracker_tasks_creator "github.com/x0k/ps2-spy/internal/stats_tracker/tasks_creator"
	"github.com/x0k/ps2-spy/internal/storage"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
	"github.com/x0k/ps2-spy/internal/tracking"
	tracking_settings "github.com/x0k/ps2-spy/internal/tracking/settings"
)

func newDiscordModule(
	log *logger.Logger,
	cfg *Config,
	m *module.Module,
	store *sql_storage.Storage,
	storePubSub pubsub.SubscriptionsManager[storage.EventType],
	ps2PubSub pubsub.SubscriptionsManager[ps2.EventType],
	trackingPubSub pubsub.SubscriptionsManager[tracking.EventType],
	charactersTrackerPubSub pubsub.SubscriptionsManager[characters_tracker.EventType],
	statsTrackerPubSub pubsub.SubscriptionsManager[stats_tracker.EventType],
	trackers *trackerDeps,
	loaders *loaderDeps,
	infra *infrastructureDeps,
	settingsService *tracking_settings.Service,
) error {
	statsTrackerTasksCreator := stats_tracker_tasks_creator.New(
		trackers.storageTasksRepo,
		cfg.StatsTracker.MaxTrackingDuration,
		cfg.StatsTracker.MaxNumberOfTasksPerChannel,
	)

	discordMessages := discord_messages.New(
		shared.Timezones,
		cfg.StatsTracker.MaxTrackingDuration,
		cfg.Tracking.MaxNumberTrackedCharacters,
		cfg.Tracking.MaxNumberTrackedOutfits,
		cfg.StatsTracker.MaxNumberOfTasksPerChannel,
	)

	discordCommands := discord_commands.New(
		log.With(sl.Component("commands")),
		discordMessages,
		discord_commands.PopulationProviders{
			ProviderRegistry: discord_commands.ProviderRegistry[loader.Simple[meta.Loaded[ps2.WorldsPopulation]]]{
				Loaders:  loaders.population,
				Priority: []string{"spy", "honu", "ps2live", "saerro", "fisu", "sanctuary", "voidwell"},
			},
		},
		discord_commands.WorldPopulationProviders{
			ProviderRegistry: discord_commands.ProviderRegistry[loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]]]{
				Loaders:  loaders.worldPopulation,
				Priority: []string{"spy", "honu", "saerro", "voidwell"},
			},
		},
		func(ctx context.Context, worldId ps2.WorldId) (meta.Loaded[ps2.WorldTerritoryControl], error) {
			platform, ok := ps2.WorldPlatforms[worldId]
			if !ok {
				return meta.Loaded[ps2.WorldTerritoryControl]{}, fmt.Errorf("unknown world %q", worldId)
			}
			control, err := infra.platformServices.WorldTracker(platform).WorldTerritoryControl(ctx, worldId)
			if err != nil {
				return meta.Loaded[ps2.WorldTerritoryControl]{}, err
			}
			return meta.LoadedNow(cfg.AppName, control), nil
		},
		discord_commands.AlertsProviders{
			ProviderRegistry: discord_commands.ProviderRegistry[loader.Simple[meta.Loaded[ps2.Alerts]]]{
				Loaders:  loaders.alerts,
				Priority: []string{"spy", "ps2alerts", "honu", "census", "voidwell"},
			},
		},
		settingsService.Load,
		infra.platformServices.LoadOutfits,
		settingsService.LoadView,
		settingsService.Update,
		trackers.statsTracker,
		discord_commands.ChannelStore{
			Loader:                      store.Channel,
			LanguageSaver:               store.SaveChannelLanguage,
			CharacterNotificationsSaver: store.SaveChannelCharacterNotifications,
			OutfitNotificationsSaver:    store.SaveChannelOutfitNotifications,
			TitleUpdatesSaver:           store.SaveChannelTitleUpdates,
			DefaultTimezoneSaver:        store.SaveChannelDefaultTimezone,
		},
		discord_commands.StatsTaskStore{
			Loader:  trackers.storageTasksRepo.ByChannelId,
			Creator: statsTrackerTasksCreator.Create,
			Remover: trackers.storageTasksRepo.Delete,
			ById:    trackers.storageTasksRepo.ById,
			Updater: statsTrackerTasksCreator.Update,
		},
	)
	m.Go(module.NewRun("discord.commands", discordCommands.Start))

	discordRunnables, err := discord_module.New(
		m,
		log.With(sl.Module("discord")),
		cfg.Discord.Token,
		cfg.Discord.CommandHandlerTimeout,
		cfg.Discord.EventHandlerTimeout,
		cfg.Discord.RemoveCommands,
		discordMessages,
		discordCommands,
		trackers.trackingManager,
		storePubSub,
		ps2PubSub,
		trackingPubSub,
		charactersTrackerPubSub,
		infra.platformServices,
		func(ctx context.Context, channelId discord.ChannelId) (int, error) {
			count := 0
			errs := make([]error, 0, len(ps2_platforms.Platforms))
			for _, platform := range ps2_platforms.Platforms {
				settings, err := trackers.settingsRepo.Get(ctx, channelId, platform)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				outfits, err := trackers.charactersTracker.OnlineOutfitMembers(ctx, platform, settings.Outfits)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				for _, outfit := range outfits {
					count += len(outfit)
				}
				characters, err := trackers.charactersTracker.OnlineCharacters(ctx, platform, settings.Characters)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				count += len(characters)
			}
			return count, errors.Join(errs...)
		},
		statsTrackerPubSub,
		store.Channel,
		settingsService.LoadDiffView,
	)
	if err != nil {
		return err
	}
	for _, r := range discordRunnables {
		m.Go(r)
	}
	return nil
}
