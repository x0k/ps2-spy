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
) (*module.Module, error) {
	m := module.New(log.Logger, "discord")
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	channelTitleUpdater := discord.NewChannelTitleUpdater(
		log.With(sl.Component("channel_title_updater")),
		session,
	)
	m.AppendVR("discord.channel_title_updater", channelTitleUpdater.Start)
	handlersChannelTitleUpdater := func(ctx context.Context, channelId discord.ChannelId, title string) error {
		channelTitleUpdater.UpdateTitle(channelId, title)
		return nil
	}

	m.AppendR("discord.session", sessionStart(
		log.With(sl.Component("session")),
		session,
		commands.Commands(),
		commandHandlerTimeout,
		removeCommands,
	))

	handlersManager := discord_event_handlers.NewHandlersManager(
		log.With(sl.Component("handlers_manager")),
		session,
		eventHandlerTimeout,
	)
	m.AppendVR("discord.handlers_manager", handlersManager.Start)

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
	m.AppendVR("discord.events_publisher", eventsPublisher.Start)
	channelLanguageUpdate := pubsub_adapters.SubscribeTo[storage.EventType, storage.ChannelLanguageSaved](m, storageSubs)
	channelTitleUpdates := pubsub_adapters.SubscribeTo[storage.EventType, storage.ChannelTitleUpdatesSaved](m, storageSubs)
	channelTrackerStarted := pubsub_adapters.SubscribeTo[stats_tracker.EventType, stats_tracker.ChannelTrackerStarted](m, statsTrackerSubs)
	channelTrackerStopped := pubsub_adapters.SubscribeTo[stats_tracker.EventType, stats_tracker.ChannelTrackerStopped](m, statsTrackerSubs)
	trackingSettingsUpdated := pubsub_adapters.SubscribeTo[tracking.EventType, tracking.TrackingSettingsUpdated](m, trackingSubs)
	m.AppendVR("discord.events_subscription", func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-channelLanguageUpdate:
				eventsPublisher.PublishChannelLanguageUpdated(ctx, e)
			case e := <-channelTitleUpdates:
				eventsPublisher.PublishChannelTitleUpdates(ctx, e)
			case e := <-channelTrackerStarted:
				eventsPublisher.PublishChannelTrackerStarted(ctx, e)
			case e := <-channelTrackerStopped:
				eventsPublisher.PublishChannelTrackerStopped(ctx, e)
			case e := <-trackingSettingsUpdated:
				eventsPublisher.PublishChannelTrackingSettingsUpdated(ctx, e)
			}
		}
	})

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
		m.AppendVR(
			fmt.Sprintf("discord.%s.events_subscription", platform),
			platformEventsPublisher.Start,
		)
		worldTrackerSubsManager := platformServicesProvider.WorldTrackerSubsManager(platform)
		playerLogin := pubsub_adapters.SubscribeTo[characters_tracker.EventType, characters_tracker.PlayerLogin](m, charactersTrackerSubs)
		playerFakeLogin := pubsub_adapters.SubscribeTo[characters_tracker.EventType, characters_tracker.PlayerFakeLogin](m, charactersTrackerSubs)
		playerLogout := pubsub_adapters.SubscribeTo[characters_tracker.EventType, characters_tracker.PlayerLogout](m, charactersTrackerSubs)
		facilityControl := pubsub_adapters.SubscribeTo[worlds_tracker.EventType, worlds_tracker.FacilityControl](m, worldTrackerSubsManager)
		facilityLoss := pubsub_adapters.SubscribeTo[worlds_tracker.EventType, worlds_tracker.FacilityLoss](m, worldTrackerSubsManager)
		outfitMembersUpdate := pubsub_adapters.SubscribeTo[ps2.EventType, ps2.OutfitMembersUpdate](m, ps2Subs)
		m.AppendVR(
			fmt.Sprintf("discord.%s.events_subscription", platform),
			func(ctx context.Context) {
				for {
					select {
					case <-ctx.Done():
						return
					case e := <-playerLogin:
						if e.Platform == platform {
							platformEventsPublisher.PublishPlayerLogin(ctx, e)
						}
					case e := <-playerFakeLogin:
						if e.Platform == platform {
							platformEventsPublisher.PublishPlayerFakeLogin(ctx, e)
						}
					case e := <-playerLogout:
						if e.Platform == platform {
							platformEventsPublisher.PublishPlayerLogout(ctx, e)
						}
					case e := <-facilityControl:
						platformEventsPublisher.PublishFacilityControl(ctx, e)
					case e := <-facilityLoss:
						platformEventsPublisher.PublishFacilityLoss(ctx, e)
					case e := <-outfitMembersUpdate:
						if e.Platform == platform {
							platformEventsPublisher.PublishOutfitMembersUpdate(ctx, e)
						}
					}
				}
			},
		)
	}

	return m, nil
}
