package discord_module

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
)

func New(
	log *logger.Logger,
	token string,
	commands []*discord.Command,
	commandHandlerTimeout time.Duration,
	eventHandlerTimeout time.Duration,
	removeCommands bool,
	charactersTrackerSubsManagers map[ps2_platforms.Platform]pubsub.SubscriptionsManager[characters_tracker.EventType],
	trackingManagers map[ps2_platforms.Platform]*tracking_manager.TrackingManager,
	handlerFactories map[discord_events.EventType]discord_events.HandlerFactory,
) (*module.Module, error) {
	m := module.New(log.Logger, "discord")
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	m.Append(NewSessionService(
		log.With(sl.Component("session")),
		m,
		session,
		commands,
		commandHandlerTimeout,
		removeCommands,
	))

	for _, platform := range ps2_platforms.Platforms {
		handlers := make(map[discord_events.EventType]discord_events.Handler, len(handlerFactories))
		for t, factory := range handlerFactories {
			handlers[t] = factory(platform)
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
				handlersManager,
			),
		)
	}

	return m, nil
}
