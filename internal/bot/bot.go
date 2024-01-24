package bot

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
)

var ErrEventTrackingChannelsLoaderNotFound = fmt.Errorf("event tracking channels loader not found")
var ErrEventsPublisherNotFound = fmt.Errorf("events publisher not found")
var ErrEventHandlerNotFound = fmt.Errorf("event handler not found")

type Bot struct {
	session             *discordgo.Session
	eventHandlerTimeout time.Duration
	removeCommands      bool
	registeredCommands  []*discordgo.ApplicationCommand
}

type BotConfig struct {
	DiscordToken           string
	RemoveCommands         bool
	CommandHandlerTimeout  time.Duration
	Ps2EventHandlerTimeout time.Duration
	Commands               []*discordgo.ApplicationCommand
	CommandHandlers        map[string]handlers.InteractionHandler
	SubmitHandlers         map[string]handlers.InteractionHandler
}

func New(
	ctx context.Context,
	cfg *BotConfig,
) (*Bot, error) {
	const op = "bot.Bot.New"
	log := infra.OpLogger(ctx, op)
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("%s creating Discord session: %w", op, err)
	}
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info("logged in as", slog.String("username", s.State.User.Username), slog.String("discriminator", s.State.User.Discriminator))
		log.Info("running on", slog.Int("serverCount", len(s.State.Guilds)))
	})
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		const op = "bot.Bot.InteractionCreateHandler"
		var userId string
		if i.Member != nil {
			userId = i.Member.User.ID
		} else {
			userId = i.User.ID
		}
		log := infra.Logger(ctx).With(
			infra.Op(op),
			slog.String("guildId", i.GuildID),
			slog.String("channelId", i.ChannelID),
			slog.String("userId", userId),
		)
		log.Debug("interaction received", slog.String("type", i.Type.String()))
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			log.Debug("command received", slog.String("command", i.ApplicationCommandData().Name))
			if handler, ok := cfg.CommandHandlers[i.ApplicationCommandData().Name]; ok {
				go handler.Run(ctx, cfg.CommandHandlerTimeout, s, i)
			} else {
				log.Warn("unknown command")
			}
		case discordgo.InteractionMessageComponent:
			log.Debug("component invoked")
		case discordgo.InteractionModalSubmit:
			data := i.ModalSubmitData()
			log.Debug("modal submitted", slog.Any("data", data))
			if handler, ok := cfg.SubmitHandlers[data.CustomID]; ok {
				go handler.Run(ctx, cfg.CommandHandlerTimeout, s, i)
			} else {
				log.Warn("unknown modal")
			}
		}
	})

	err = session.Open()
	if err != nil {
		return nil, fmt.Errorf("%s session open: %w", op, err)
	}
	log.Info("adding commands")
	registeredCommands, err := session.ApplicationCommandBulkOverwrite(session.State.User.ID, "", cfg.Commands)
	if err != nil {
		return nil, fmt.Errorf("%s registering commands: %w", op, err)
	}
	return &Bot{
		session:             session,
		eventHandlerTimeout: cfg.Ps2EventHandlerTimeout,
		removeCommands:      cfg.RemoveCommands,
		registeredCommands:  registeredCommands,
	}, nil
}

func (b *Bot) StartEventHandlers(
	ctx context.Context,
	eventTrackingChannelsLoader loaders.QueriedLoader[any, []meta.ChannelId],
	eventHandlers EventHandlers,
) error {
	eventHandlersConfig := &handlers.Ps2EventHandlerConfig{
		Session:                     b.session,
		Timeout:                     b.eventHandlerTimeout,
		EventTrackingChannelsLoader: eventTrackingChannelsLoader,
	}
	return eventHandlers.Start(ctx, eventHandlersConfig)
}

func (b *Bot) Stop(ctx context.Context) error {
	const op = "bot.Bot.Stop"
	log := infra.OpLogger(ctx, op)
	log.Info("stopping bot")
	if b.removeCommands {
		for _, v := range b.registeredCommands {
			if err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", v.ID); err != nil {
				log.Error("cannot delete command", slog.String("command", v.Name), sl.Err(err))
			}
		}
	}
	return b.session.Close()
}
