package discord_module

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/module"
)

type Config struct {
	Token                 string
	RemoveCommands        bool
	CommandHandlerTimeout time.Duration
	EventHandlerTimeout   time.Duration
	Commands              []*Command
}

func New(
	log *logger.Logger,
	cfg *Config,
) (*module.Module, error) {
	m := module.New(log.Logger, "discord")
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, err
	}
	m.Append(NewSessionService(log, cfg, m, session))
	return m, nil
}
