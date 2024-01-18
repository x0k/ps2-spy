package bot

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/loaders"
)

type Bot struct {
	wg                 *sync.WaitGroup
	log                *slog.Logger
	session            *discordgo.Session
	registeredCommands []*discordgo.ApplicationCommand
}

type BotConfig struct {
	DiscordToken                string
	CommandHandlerTimeout       time.Duration
	Ps2EventHandlerTimeout      time.Duration
	Commands                    []*discordgo.ApplicationCommand
	CommandHandlers             map[string]handlers.InteractionHandler
	SubmitHandlers              map[string]handlers.InteractionHandler
	PlayerLoginHandler          handlers.Ps2EventHandler[ps2events.PlayerLogin]
	EventTrackingChannelsLoader loaders.QueriedLoader[any, []string]
	EventsPublisher             *ps2events.Publisher
}

func New(
	ctx context.Context,
	log *slog.Logger,
	cfg *BotConfig,
) (*Bot, error) {
	log = log.With(slog.String("component", "bot"))
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info("logged in as", slog.String("username", s.State.User.Username), slog.String("discriminator", s.State.User.Discriminator))
		log.Info("running on", slog.Int("server_count", len(s.State.Guilds)))
	})
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		var userId string
		if i.Member != nil {
			userId = i.Member.User.ID
		} else {
			userId = i.User.ID
		}
		l := log.With(
			slog.String("guild_id", i.GuildID),
			slog.String("user_id", userId),
		)
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			l.Debug("command received", slog.String("command", i.ApplicationCommandData().Name))
			if handler, ok := cfg.CommandHandlers[i.ApplicationCommandData().Name]; ok {
				go handler.Run(ctx, l, cfg.CommandHandlerTimeout, s, i)
			} else {
				l.Warn("unknown command")
			}
		case discordgo.InteractionMessageComponent:
			l.Debug("component invoked")
		case discordgo.InteractionModalSubmit:
			data := i.ModalSubmitData()
			l.Debug("modal submitted", slog.Any("data", data))
			if handler, ok := cfg.SubmitHandlers[data.CustomID]; ok {
				go handler.Run(ctx, l, cfg.CommandHandlerTimeout, s, i)
			} else {
				l.Warn("unknown modal")
			}
		}
	})
	eventHandlersConfig := &handlers.Ps2EventHandlerConfig{
		Session:                     session,
		Timeout:                     cfg.Ps2EventHandlerTimeout,
		EventTrackingChannelsLoader: cfg.EventTrackingChannelsLoader,
	}
	wg := &sync.WaitGroup{}
	if cfg.PlayerLoginHandler != nil {
		playerLogin := make(chan ps2events.PlayerLogin)
		playerLoginUnSub, err := cfg.EventsPublisher.AddHandler(playerLogin)
		if err != nil {
			return nil, err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer playerLoginUnSub()
			for {
				select {
				case <-ctx.Done():
					return
				case pl := <-playerLogin:
					go cfg.PlayerLoginHandler.Run(
						ctx,
						log.With(slog.Any("event", pl)),
						eventHandlersConfig,
						pl,
					)
				}
			}
		}()
	} else {
		log.Warn("no player login handler")
	}

	err = session.Open()
	if err != nil {
		return nil, err
	}
	log.Info("adding commands")
	registeredCommands := make([]*discordgo.ApplicationCommand, 0, len(cfg.Commands))
	for _, v := range cfg.Commands {
		cmd, err := session.ApplicationCommandCreate(session.State.User.ID, "", v)
		if err != nil {
			log.Error("cannot create command", slog.String("command", v.Name), sl.Err(err))
		} else {
			registeredCommands = append(registeredCommands, cmd)
		}
	}
	return &Bot{
		wg:                 wg,
		log:                log,
		session:            session,
		registeredCommands: registeredCommands,
	}, nil
}

func (b *Bot) Stop() error {
	b.log.Info("stopping bot")
	for _, v := range b.registeredCommands {
		err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", v.ID)
		if err != nil {
			b.log.Error("cannot delete command", slog.String("command", v.Name), sl.Err(err))
		}
	}
	return b.session.Close()
}
