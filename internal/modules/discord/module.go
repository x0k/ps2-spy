package discord_module

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	pubsub_adapters "github.com/x0k/ps2-spy/internal/adapters/pubsub"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_commands "github.com/x0k/ps2-spy/internal/discord/commands"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_event_handlers "github.com/x0k/ps2-spy/internal/discord/events/handlers"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type PlatformServicesProvider interface {
	WorldTrackerSubsManager(platform ps2_platforms.Platform) pubsub.SubscriptionsManager[worlds_tracker.EventType]
	CharacterLoader(platform ps2_platforms.Platform) loader.Keyed[ps2.CharacterId, ps2.Character]
	CharactersLoader(platform ps2_platforms.Platform) loader.Multi[ps2.CharacterId, ps2.Character]
	OutfitLoader(platform ps2_platforms.Platform) loader.Keyed[ps2.OutfitId, ps2.Outfit]
	OutfitsLoader(platform ps2_platforms.Platform) loader.Multi[ps2.OutfitId, ps2.Outfit]
	FacilityLoader(platform ps2_platforms.Platform) loader.Keyed[ps2.FacilityId, ps2.Facility]
}

func New(
	stopper module.Stopper,
	log *logger.Logger,
	token string,
	commandHandlerTimeout time.Duration,
	eventHandlerTimeout time.Duration,
	removeCommands bool,
	messages *discord_messages.Messages,
	commands *discord_commands.Commands,
	trackingManager *tracking.Manager,
	storageSubs pubsub.SubscriptionsManager[storage.EventType],
	ps2Subs pubsub.SubscriptionsManager[ps2.EventType],
	trackingSubs pubsub.SubscriptionsManager[tracking.EventType],
	charactersTrackerSubs pubsub.SubscriptionsManager[characters_tracker.EventType],
	platformServicesProvider PlatformServicesProvider,
	onlineTrackableEntitiesCountLoader loader.Keyed[discord.ChannelId, int],
	statsTrackerSubs pubsub.SubscriptionsManager[stats_tracker.EventType],
	channelLoader discord_events.ChannelLoader,
	trackingSettingsDiffViewLoader discord_event_handlers.TrackingSettingsDiffViewLoader,
) ([]module.Runnable, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	channelTitleUpdater := discord.NewChannelTitleUpdater(
		log.With(sl.Component("channel_title_updater")),
		session,
	)

	var runnables []module.Runnable
	runnables = append(runnables, module.NewRun("discord.channel_title_updater", func(ctx context.Context) error {
		channelTitleUpdater.Start(ctx)
		return nil
	}))
	handlersChannelTitleUpdater := func(ctx context.Context, channelId discord.ChannelId, title string) error {
		channelTitleUpdater.UpdateTitle(channelId, title)
		return nil
	}

	runnables = append(runnables, module.NewRun("discord.session", sessionStart(
		log.With(sl.Component("session")),
		session,
		commands.Commands(),
		commandHandlerTimeout,
		removeCommands,
	)))

	handlersManager := discord_event_handlers.NewHandlersManager(
		log.With(sl.Component("handlers_manager")),
		session,
		eventHandlerTimeout,
	)
	runnables = append(runnables, module.NewRun("discord.handlers_manager", func(ctx context.Context) error {
		handlersManager.Start(ctx)
		return nil
	}))

	eventsPubSub := pubsub.New[discord_events.EventType]()
	for _, handler := range discord_event_handlers.New(
		handlersManager,
		messages,
		onlineTrackableEntitiesCountLoader,
		handlersChannelTitleUpdater,
		trackingSettingsDiffViewLoader,
	) {
		eventsPubSub.AddHandler(handler)
	}
	eventsPublisher := discord_events.NewEventsPublisher(
		log.With(sl.Component("events_publisher")),
		eventsPubSub,
		channelLoader,
	)
	runnables = append(runnables, module.NewRun("discord.events_publisher", func(ctx context.Context) error {
		eventsPublisher.Start(ctx)
		return nil
	}))
	runnables = append(runnables, pubsub_adapters.Listen(stopper, storageSubs, "discord.channel_language_update", func(ctx context.Context, e storage.ChannelLanguageSaved) {
		eventsPublisher.PublishChannelLanguageUpdated(ctx, e)
	}))
	runnables = append(runnables, pubsub_adapters.Listen(stopper, storageSubs, "discord.channel_title_updates", func(ctx context.Context, e storage.ChannelTitleUpdatesSaved) {
		eventsPublisher.PublishChannelTitleUpdates(ctx, e)
	}))
	runnables = append(runnables, pubsub_adapters.Listen(stopper, statsTrackerSubs, "discord.channel_tracker_started", func(ctx context.Context, e stats_tracker.ChannelTrackerStarted) {
		eventsPublisher.PublishChannelTrackerStarted(ctx, e)
	}))
	runnables = append(runnables, pubsub_adapters.Listen(stopper, statsTrackerSubs, "discord.channel_tracker_stopped", func(ctx context.Context, e stats_tracker.ChannelTrackerStopped) {
		eventsPublisher.PublishChannelTrackerStopped(ctx, e)
	}))
	runnables = append(runnables, pubsub_adapters.Listen(stopper, trackingSubs, "discord.tracking_settings_updated", func(ctx context.Context, e tracking.TrackingSettingsUpdated) {
		eventsPublisher.PublishChannelTrackingSettingsUpdated(ctx, e)
	}))

	for _, platform := range ps2_platforms.Platforms {

		platformEventsPubSub := pubsub.New[discord_events.EventType]()

		for _, handler := range discord_event_handlers.NewPlatform(
			handlersManager,
			messages,
			platform,
			platformServicesProvider.OutfitLoader(platform),
			platformServicesProvider.FacilityLoader(platform),
			platformServicesProvider.CharactersLoader(platform),
			platformServicesProvider.CharacterLoader(platform),
			onlineTrackableEntitiesCountLoader,
			handlersChannelTitleUpdater,
		) {
			platformEventsPubSub.AddHandler(handler)
		}
		platformEventsPublisher := discord_events.NewPlatformEventsPublisher(
			log.With(sl.Component("platform_events_publisher")),
			platformEventsPubSub,
			func(ctx context.Context, charId ps2.CharacterId) ([]discord.Channel, error) {
				return trackingManager.CharacterChannels(ctx, platform, charId)
			},
			func(ctx context.Context, outfitId ps2.OutfitId) ([]discord.Channel, error) {
				return trackingManager.OutfitChannels(ctx, platform, outfitId)
			},
		)
		runnables = append(runnables, module.NewRun(
			fmt.Sprintf("discord.%s.events_subscription", platform),
			func(ctx context.Context) error {
				platformEventsPublisher.Start(ctx)
				return nil
			},
		))
		worldTrackerSubsManager := platformServicesProvider.WorldTrackerSubsManager(platform)
		runnables = append(runnables, pubsub_adapters.Listen(stopper, charactersTrackerSubs, fmt.Sprintf("discord.%s.player_login", platform), func(ctx context.Context, e characters_tracker.PlayerLogin) {
			if e.Platform == platform {
				platformEventsPublisher.PublishPlayerLogin(ctx, e)
			}
		}))
		runnables = append(runnables, pubsub_adapters.Listen(stopper, charactersTrackerSubs, fmt.Sprintf("discord.%s.player_fake_login", platform), func(ctx context.Context, e characters_tracker.PlayerFakeLogin) {
			if e.Platform == platform {
				platformEventsPublisher.PublishPlayerFakeLogin(ctx, e)
			}
		}))
		runnables = append(runnables, pubsub_adapters.Listen(stopper, charactersTrackerSubs, fmt.Sprintf("discord.%s.player_logout", platform), func(ctx context.Context, e characters_tracker.PlayerLogout) {
			if e.Platform == platform {
				platformEventsPublisher.PublishPlayerLogout(ctx, e)
			}
		}))
		runnables = append(runnables, pubsub_adapters.Listen(stopper, worldTrackerSubsManager, fmt.Sprintf("discord.%s.facility_control", platform), func(ctx context.Context, e worlds_tracker.FacilityControl) {
			platformEventsPublisher.PublishFacilityControl(ctx, e)
		}))
		runnables = append(runnables, pubsub_adapters.Listen(stopper, worldTrackerSubsManager, fmt.Sprintf("discord.%s.facility_loss", platform), func(ctx context.Context, e worlds_tracker.FacilityLoss) {
			platformEventsPublisher.PublishFacilityLoss(ctx, e)
		}))
		runnables = append(runnables, pubsub_adapters.Listen(stopper, ps2Subs, fmt.Sprintf("discord.%s.outfit_members_update", platform), func(ctx context.Context, e ps2.OutfitMembersUpdate) {
			if e.Platform == platform {
				platformEventsPublisher.PublishOutfitMembersUpdate(ctx, e)
			}
		}))
	}

	return runnables, nil
}
