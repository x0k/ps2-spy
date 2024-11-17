package discord_module

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func New(
	log *logger.Logger,
	token string,
	commands []*Command,
	commandHandlerTimeout time.Duration,
	eventHandlerTimeout time.Duration,
	removeCommands bool,
	charactersTrackerSubsManagers map[ps2_platforms.Platform]pubsub.SubscriptionsManager[characters_tracker.EventType],
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
	return m, nil
}
