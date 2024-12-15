package discord_module

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
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
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func New(
	log *logger.Logger,
	token string,
	commandHandlerTimeout time.Duration,
	eventHandlerTimeout time.Duration,
	removeCommands bool,
	charactersTrackerSubsManagers map[ps2_platforms.Platform]pubsub.SubscriptionsManager[characters_tracker.EventType],
	trackingManagers map[ps2_platforms.Platform]*tracking_manager.TrackingManager,
	storageSubs pubsub.SubscriptionsManager[storage.EventType],
	worldTrackerSubsMangers map[ps2_platforms.Platform]pubsub.SubscriptionsManager[worlds_tracker.EventType],
	populationLoaders map[string]loader.Simple[meta.Loaded[ps2.WorldsPopulation]],
	worldPopulationLoaders map[string]loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]],
	worldTerritoryControlLoader loader.Keyed[ps2.WorldId, meta.Loaded[ps2.WorldTerritoryControl]],
	alertsLoaders map[string]loader.Simple[meta.Loaded[ps2.Alerts]],
	onlineTrackableEntitiesLoader loader.Keyed[discord.SettingsQuery, discord.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character]],
	outfitsLoader loader.Queried[discord.PlatformQuery[[]ps2.OutfitId], map[ps2.OutfitId]ps2.Outfit],
	trackingSettingsLoader loader.Keyed[discord.SettingsQuery, discord.TrackingSettings],
	characterNamesLoader loader.Queried[discord.PlatformQuery[[]ps2.CharacterId], []string],
	characterIdsLoader loader.Queried[discord.PlatformQuery[[]string], []ps2.CharacterId],
	outfitTagsLoader loader.Queried[discord.PlatformQuery[[]ps2.OutfitId], []string],
	outfitIdsLoader loader.Queried[discord.PlatformQuery[[]string], []ps2.OutfitId],
	saveChannelTrackingSettings discord_commands.ChannelTrackingSettingsSaver,
	saveChannelLanguage discord_commands.ChannelLanguageSaver,
	characterLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.CharacterId, ps2.Character],
	outfitLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.OutfitId, ps2.Outfit],
	charactersLoaders map[ps2_platforms.Platform]loader.Multi[ps2.CharacterId, ps2.Character],
	facilityLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.FacilityId, ps2.Facility],
	onlineTrackableEntitiesCountLoader loader.Keyed[discord.ChannelId, int],
	statsTracker *stats_tracker.StatsTracker,
	statsTrackerSubs pubsub.SubscriptionsManager[stats_tracker.EventType],
	channelLoader discord_events.ChannelLoader,
) (*module.Module, error) {
	m := module.New(log.Logger, "discord")
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	messages := discord_messages.New()
	commands := discord_commands.New(
		log.With(sl.Component("commands")),
		messages,
		populationLoaders,
		[]string{"spy", "honu", "ps2live", "saerro", "fisu", "sanctuary", "voidwell"},
		worldPopulationLoaders,
		[]string{"spy", "honu", "saerro", "voidwell"},
		worldTerritoryControlLoader,
		alertsLoaders,
		[]string{"spy", "ps2alerts", "honu", "census", "voidwell"},
		onlineTrackableEntitiesLoader,
		outfitsLoader,
		trackingSettingsLoader,
		characterNamesLoader,
		characterIdsLoader,
		outfitTagsLoader,
		outfitIdsLoader,
		saveChannelTrackingSettings,
		saveChannelLanguage,
		statsTracker,
		channelLoader,
	)
	m.AppendR("discord.commands", commands.Start)

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
	) {
		eventsPubSub.AddHandler(handler)
	}
	eventsPublisher := discord_events.NewEventsPublisher(
		log.With(sl.Component("events_publisher")),
		eventsPubSub,
		channelLoader,
	)
	m.AppendVR("discord.events_publisher", eventsPublisher.Start)
	channelLanguageUpdate := storage.Subscribe[storage.ChannelSaved](m, storageSubs)
	channelTrackerStarted := stats_tracker.Subscribe[stats_tracker.ChannelTrackerStarted](m, statsTrackerSubs)
	channelTrackerStopped := stats_tracker.Subscribe[stats_tracker.ChannelTrackerStopped](m, statsTrackerSubs)
	m.AppendVR("discord.events_subscription", func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-channelLanguageUpdate:
				eventsPublisher.PublishChannelLanguageUpdated(ctx, e)
			case e := <-channelTrackerStarted:
				eventsPublisher.PublishChannelTrackerStarted(ctx, e)
			case e := <-channelTrackerStopped:
				eventsPublisher.PublishChannelTrackerStopped(ctx, e)
			}
		}
	})

	for _, platform := range ps2_platforms.Platforms {

		platformEventsPubSub := pubsub.New[discord_events.EventType]()

		for _, handler := range discord_event_handlers.NewPlatform(
			handlersManager,
			messages,
			platform,
			outfitLoaders[platform],
			facilityLoaders[platform],
			charactersLoaders[platform],
			characterLoaders[platform],
			onlineTrackableEntitiesCountLoader,
			handlersChannelTitleUpdater,
		) {
			platformEventsPubSub.AddHandler(handler)
		}
		platformEventsPublisher := discord_events.NewPlatformEventsPublisher(
			log.With(sl.Component("platform_events_publisher")),
			platformEventsPubSub,
			func(ctx context.Context, oi ps2.CharacterId) ([]discord.Channel, error) {
				return trackingManagers[platform].ChannelIdsForCharacter(ctx, oi)
			},
			func(ctx context.Context, oi ps2.OutfitId) ([]discord.Channel, error) {
				return trackingManagers[platform].ChannelIdsForOutfit(ctx, oi)
			},
		)
		m.AppendVR(
			fmt.Sprintf("discord.%s.events_subscription", platform),
			platformEventsPublisher.Start,
		)
		playerLogin := characters_tracker.Subscribe[characters_tracker.PlayerLogin](m, charactersTrackerSubsManagers[platform])
		playerLogout := characters_tracker.Subscribe[characters_tracker.PlayerLogout](m, charactersTrackerSubsManagers[platform])
		facilityControl := worlds_tracker.Subscribe[worlds_tracker.FacilityControl](m, worldTrackerSubsMangers[platform])
		facilityLoss := worlds_tracker.Subscribe[worlds_tracker.FacilityLoss](m, worldTrackerSubsMangers[platform])
		outfitMembersUpdate := storage.Subscribe[storage.OutfitMembersUpdate](m, storageSubs)
		m.AppendVR(
			fmt.Sprintf("discord.%s.events_subscription", platform),
			func(ctx context.Context) {
				for {
					select {
					case <-ctx.Done():
						return
					case e := <-playerLogin:
						platformEventsPublisher.PublishPlayerLogin(ctx, e)
					case e := <-playerLogout:
						platformEventsPublisher.PublishPlayerLogout(ctx, e)
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
