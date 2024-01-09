package bot

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/handlers/alerts"
	"github.com/x0k/ps2-spy/internal/bot/handlers/population"
	serverpopulation "github.com/x0k/ps2-spy/internal/bot/handlers/server_population"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type Bot struct {
	log                *slog.Logger
	session            *discordgo.Session
	registeredCommands []*discordgo.ApplicationCommand
}

type BotConfig struct {
	DiscordToken          string
	CommandHandlerTimeout time.Duration
}

func New(ctx context.Context, cfg *BotConfig, log *slog.Logger, service *ps2.Service) (*Bot, error) {
	log = log.With(slog.String("component", "bot"))
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info("logged in as", slog.String("username", s.State.User.Username), slog.String("discriminator", s.State.User.Discriminator))
		log.Info("running on", slog.Int("server_count", len(s.State.Guilds)))
	})
	commands := makeCommands(service)
	hs := map[string]handlers.InteractionHandler{
		"population":        population.New(service),
		"server-population": serverpopulation.New(service),
		"alerts":            alerts.New(service),
	}
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		var userId string
		if i.Member != nil {
			userId = i.Member.User.ID
		} else {
			userId = i.User.ID
		}
		l := log.With(
			slog.String("command", i.ApplicationCommandData().Name),
			slog.String("guild_id", i.GuildID),
			slog.String("user_id", userId),
		)
		l.Debug("command received")
		if handler, ok := hs[i.ApplicationCommandData().Name]; ok {
			go handler.Run(ctx, l, cfg.CommandHandlerTimeout, s, i)
		} else {
			l.Warn("unknown command")
		}
	})
	err = session.Open()
	if err != nil {
		return nil, err
	}
	log.Info("adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, 0, len(commands))
	for _, v := range commands {
		cmd, err := session.ApplicationCommandCreate(session.State.User.ID, "", v)
		if err != nil {
			log.Error("cannot create command", slog.String("command", v.Name), sl.Err(err))
		} else {
			registeredCommands = append(registeredCommands, cmd)
		}
	}
	return &Bot{
		log:                log,
		session:            session,
		registeredCommands: registeredCommands,
	}, nil
}

func (b *Bot) Stop() error {
	for _, v := range b.registeredCommands {
		err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", v.ID)
		if err != nil {
			b.log.Error("cannot delete command", slog.String("command", v.Name), sl.Err(err))
		}
	}
	return b.session.Close()
}
