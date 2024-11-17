package discord_module

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
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
	commands []*Command,
	commandHandlerTimeout time.Duration,
	eventHandlerTimeout time.Duration,
	removeCommands bool,
	charactersTrackerSubsManagers map[ps2_platforms.Platform]pubsub.SubscriptionsManager[characters_tracker.EventType],
	trackingManagers map[ps2_platforms.Platform]*tracking_manager.TrackingManager,
	handlers []Handler,
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

	unsubs := make([]func(), 0, len(handlers)*len(ps2_platforms.Platforms))

	for _, platform := range ps2_platforms.Platforms {
		eventsPubSub := pubsub.New[EventType]()
		eventsPublisher := NewPublisher(eventsPubSub, trackingManagers[platform])
		m.Append(newEventsSubscriptionService(
			log.With(sl.Component("events_subscription")),
			m,
			platform,
			charactersTrackerSubsManagers[platform],
			eventsPublisher,
		))

		for _, handler := range handlers {
			unsubs = append(unsubs, eventsPubSub.AddHandler(handler.ForPlatform(platform)))
		}
	}

	m.PreStop(module.NewHook("handlers_unsubscribe", func(_ context.Context) error {
		for _, unsub := range unsubs {
			unsub()
		}
		return nil
	}))

	return m, nil
}
