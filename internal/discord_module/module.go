package discord_module

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/module"
)

func New(
	log *logger.Logger,
	cfg *Config,
) (*module.Module, error) {
	m := module.New(log.Logger, "discord")
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, err
	}
	m.PostStart(NewSessionService(log, cfg, m, session))
	return m, nil
}
