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
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
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
) (*module.Module, error) {
	m := module.New(log.Logger, "discord")
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	messages := discord_messages.New()
	commands := discord_commands.New(
		"discord_commands",
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
	)
	m.Append(commands)

	channelTitleUpdater := discord.NewChannelTitleUpdater(
		log.With(sl.Component("channel_title_updater")),
		session,
	)
	m.AppendServiceFn("channel_title_updater", func(ctx context.Context) error {
		channelTitleUpdater.Start(ctx)
		return nil
	})

	handlerFactories := discord_events.NewHandlers(
		log.With(sl.Component("discord_event_handlers")),
		messages,
		characterLoaders,
		outfitLoaders,
		charactersLoaders,
		facilityLoaders,
		onlineTrackableEntitiesCountLoader,
		func(_ context.Context, channelId discord.ChannelId, title string) error {
			channelTitleUpdater.UpdateTitle(channelId, title)
			return nil
		},
	)

	m.Append(NewSessionService(
		log.With(sl.Component("session")),
		m,
		session,
		commands.Commands(),
		commandHandlerTimeout,
		removeCommands,
	))

	for _, platform := range ps2_platforms.Platforms {
		handlers := make(map[discord_events.EventType][]discord_events.Handler, len(handlerFactories))
		for t, factories := range handlerFactories {
			eventHandlers := make([]discord_events.Handler, 0, len(factories))
			for _, factory := range factories {
				eventHandlers = append(eventHandlers, factory(platform))
			}
			handlers[t] = eventHandlers
		}
		handlersManager := discord_events.NewHandlersManager(
			fmt.Sprintf("discord.%s.handlers", platform),
			log.With(sl.Component("handlers_manager")),
			session,
			handlers,
			trackingManagers[platform],
			eventHandlerTimeout,
		)
		m.Append(
			handlersManager,
			newEventsSubscriptionService(
				m,
				platform,
				charactersTrackerSubsManagers[platform],
				storageSubs,
				worldTrackerSubsMangers[platform],
				handlersManager,
			),
		)
	}

	return m, nil
}
