package app

import (
	"context"

	pubsub_adapters "github.com/x0k/ps2-spy/internal/adapters/pubsub"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/metrics"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_outfit_members_synchronizer "github.com/x0k/ps2-spy/internal/ps2/outfit_members_synchronizer"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	ps2_storage_outfits_repo "github.com/x0k/ps2-spy/internal/ps2/storage_outfits_repo"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	stats_tracker_storage_tasks_repo "github.com/x0k/ps2-spy/internal/stats_tracker/storage_tasks_repo"
	"github.com/x0k/ps2-spy/internal/storage"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
	"github.com/x0k/ps2-spy/internal/tracking"
	tracking_settings "github.com/x0k/ps2-spy/internal/tracking/settings"
)

type trackerDeps struct {
	settingsRepo      *tracking_settings.Repository
	trackingManager   *tracking.Manager
	storageTasksRepo  *stats_tracker_storage_tasks_repo.Repository
	statsTracker      *stats_tracker.StatsTracker
	charactersTracker *characters_tracker.Tracker
}

func newTrackers(
	log *logger.Logger,
	cfg *Config,
	m *module.Module,
	store *sql_storage.Storage,
	infra *infrastructureDeps,
	statsTrackerPubSub pubsub.Publisher[stats_tracker.Event],
	charactersTrackerPubSub pubsub.Publisher[characters_tracker.Event],
	mt *metrics.Metrics,
) *trackerDeps {
	settingsRepo := tracking_settings.NewRepository(store, log.With(sl.Component("tracking_repo")))

	trackingManager := tracking.New(
		log.With(sl.Component("tracking_manager")),
		func(ctx context.Context, platform ps2_platforms.Platform, characterId ps2.CharacterId) (ps2.Character, error) {
			return infra.platformServices.CharacterLoader(platform)(ctx, characterId)
		},
		func(ctx context.Context, platform ps2_platforms.Platform, c ps2.Character) ([]discord.Channel, error) {
			return settingsRepo.TrackingChannelsForCharacter(ctx, platform, c.Id, c.OutfitId)
		},
		settingsRepo.AllTrackableCharacterIdsWithDuplicationsForPlatform,
		infra.storageOutfitsRepo.MemberIds,
		settingsRepo.TrackingChannelsForOutfit,
		settingsRepo.AllTrackableOutfitIdsWithDuplicationsForPlatform,
	)
	m.Go(module.NewRun("tracking_manager", func(ctx context.Context) error {
		trackingManager.Start(ctx)
		return nil
	}))

	storageTasksRepo := stats_tracker_storage_tasks_repo.New(store)
	statsTracker := stats_tracker.New(
		log.With(sl.Component("stats_tracker")),
		statsTrackerPubSub,
		settingsRepo.PlatformsByChannelId,
		storageTasksRepo.ChannelsWithActiveTasks,
		func(ctx context.Context, platform ps2_platforms.Platform, charId ps2.CharacterId) ([]discord.ChannelId, error) {
			channels, err := trackingManager.CharacterChannels(ctx, platform, charId)
			if err != nil {
				return nil, err
			}
			channelIds := make([]discord.ChannelId, 0, len(channels))
			for _, channel := range channels {
				channelIds = append(channelIds, channel.Id)
			}
			return channelIds, nil
		},
		func(
			ctx context.Context, platform ps2_platforms.Platform, characterIds []ps2.CharacterId,
		) (map[ps2.CharacterId]ps2.Character, error) {
			return infra.platformServices.CharactersLoader(platform)(ctx, characterIds)
		},
		cfg.StatsTracker.MaxTrackingDuration,
	)
	m.Go(module.NewRun("stats_tracker", func(ctx context.Context) error {
		statsTracker.Start(ctx)
		return nil
	}))

	charactersTracker := characters_tracker.New(
		log.With(sl.Component("platforms_characters_tracker")),
		func(ctx context.Context, platform ps2_platforms.Platform, characterId ps2.CharacterId) (ps2.Character, error) {
			return infra.platformServices.CharacterLoader(platform)(ctx, characterId)
		},
		charactersTrackerPubSub,
		mt,
	)
	m.Go(module.NewRun("characters_tracker", func(ctx context.Context) error {
		charactersTracker.Start(ctx)
		return nil
	}))

	return &trackerDeps{
		settingsRepo:      settingsRepo,
		trackingManager:   trackingManager,
		storageTasksRepo:  storageTasksRepo,
		statsTracker:      statsTracker,
		charactersTracker: charactersTracker,
	}
}

func newSettingsService(
	log *logger.Logger,
	cfg *Config,
	storePubSub pubsub.Publisher[storage.Event],
	infra *infrastructureDeps,
	trackers *trackerDeps,
	trackingPubSub pubsub.Publisher[tracking.Event],
) *tracking_settings.Service {
	return tracking_settings.New(tracking_settings.Opts{
		SettingsRepo:         trackers.settingsRepo,
		OutfitsRepo:          infra.censusOutfitsRepo,
		CharactersRepo:       infra.censusCharactersRepo,
		TrackingRepo:         trackers.charactersTracker,
		MaxTrackedOutfits:    cfg.Tracking.MaxNumberTrackedOutfits,
		MaxTrackedCharacters: cfg.Tracking.MaxNumberTrackedCharacters,
		Publisher:            trackingPubSub,
		StoragePublisher:     storePubSub,
	})
}

func newOutfitSync(
	log *logger.Logger,
	cfg *Config,
	m *module.Module,
	ps2PubSub pubsub.Publisher[ps2.Event],
	infra *infrastructureDeps,
) *ps2_outfit_members_synchronizer.OutfitMembersSynchronizer[*ps2_storage_outfits_repo.Repository] {
	outfitSync := ps2_outfit_members_synchronizer.New(
		log.With(sl.Component("outfit_members_synchronizer")),
		infra.storageOutfitsRepo,
		infra.censusOutfitsRepo,
		cfg.Ps2.OutfitsSynchronizeInterval,
		ps2PubSub,
	)
	m.Go(module.NewRun("outfit_members_synchronizer", func(ctx context.Context) error {
		outfitSync.Start(ctx)
		return nil
	}))
	return outfitSync
}

func newStorageEventsSubscription(
	log *logger.Logger,
	m *module.Module,
	ps2PubSub pubsub.SubscriptionsManager[ps2.EventType],
	trackers *trackerDeps,
) {
	m.Go(pubsub_adapters.Listen(m, ps2PubSub, "storage.outfit_members_added", func(ctx context.Context, e ps2.OutfitMembersAdded) {
		if err := trackers.trackingManager.TrackOutfitMembers(e.OutfitId, e.Platform, e.CharacterIds); err != nil {
			log.Error(ctx, "failed to track outfit members", sl.Err(err))
		}
	}))
	m.Go(pubsub_adapters.Listen(m, ps2PubSub, "storage.outfit_members_removed", func(ctx context.Context, e ps2.OutfitMembersRemoved) {
		if err := trackers.trackingManager.UntrackOutfitMembers(e.OutfitId, e.Platform, e.CharacterIds); err != nil {
			log.Error(ctx, "failed to untrack outfit members", sl.Err(err))
		}
	}))
}

func newTrackingSettingsSubscription(
	log *logger.Logger,
	m *module.Module,
	trackingPubSub pubsub.SubscriptionsManager[tracking.EventType],
	trackers *trackerDeps,
	outfitSync *ps2_outfit_members_synchronizer.OutfitMembersSynchronizer[*ps2_storage_outfits_repo.Repository],
) {
	m.Go(pubsub_adapters.Listen(m, trackingPubSub, "tracking.settings_updated", func(ctx context.Context, e tracking.TrackingSettingsUpdated) {
		if err := trackers.trackingManager.HandleTrackingSettingsUpdate(ctx, e.Platform, e); err != nil {
			log.Error(ctx, "failed to handle tracking settings update", sl.Err(err))
		}
		for _, oId := range e.Diff.Outfits.ToAdd {
			outfitSync.SyncOutfit(ctx, e.Platform, oId)
		}
	}))
}
