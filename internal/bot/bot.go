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
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
)

var ErrEventTrackingChannelsLoaderNotFound = fmt.Errorf("event tracking channels loader not found")
var ErrEventsPublisherNotFound = fmt.Errorf("events publisher not found")
var ErrEventHandlerNotFound = fmt.Errorf("event handler not found")

type Bot struct {
	log                 *logger.Logger
	session             *discordgo.Session
	eventHandlerTimeout time.Duration
	removeCommands      bool
	registeredCommands  []*discordgo.ApplicationCommand
}

func New(
	ctx context.Context,
	log *logger.Logger,
	cfg *BotConfig,
) (*Bot, error) {
	const op = "bot.Bot.New"
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("%s creating Discord session: %w", op, err)
	}
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info(ctx, "logged in as", slog.String("username", s.State.User.Username), slog.String("discriminator", s.State.User.Discriminator))
		log.Info(ctx, "running on", slog.Int("serverCount", len(s.State.Guilds)))
	})
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		const op = "bot.Bot.InteractionCreateHandler"
		var userId string
		if i.Member != nil {
			userId = i.Member.User.ID
		} else {
			userId = i.User.ID
		}
		hLog := log.With(
			infra.Op(op),
			slog.String("guild_id", i.GuildID),
			slog.String("channel_id", i.ChannelID),
			slog.String("user_id", userId),
		)
		hLog.Debug(ctx, "interaction received", slog.String("type", i.Type.String()))
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			cLog := hLog.With(slog.String("command_name", i.ApplicationCommandData().Name))
			cLog.Debug(ctx, "command received")
			if handler, ok := cfg.CommandHandlers[i.ApplicationCommandData().Name]; ok {
				go handler.Run(ctx, cLog, cfg.CommandHandlerTimeout, s, i)
			} else {
				hLog.Warn(ctx, "unknown command")
			}
		case discordgo.InteractionMessageComponent:
			hLog.Debug(ctx, "component invoked")
		case discordgo.InteractionModalSubmit:
			data := i.ModalSubmitData()
			smLog := hLog.With(slog.Any("modal_data", data))
			smLog.Debug(ctx, "modal submitted")
			if handler, ok := cfg.SubmitHandlers[data.CustomID]; ok {
				go handler.Run(ctx, smLog, cfg.CommandHandlerTimeout, s, i)
			} else {
				smLog.Warn(ctx, "unknown modal")
			}
		}
	})

	err = session.Open()
	if err != nil {
		return nil, fmt.Errorf("%s session open: %w", op, err)
	}
	log.Info(ctx, "adding commands")
	registeredCommands, err := session.ApplicationCommandBulkOverwrite(session.State.User.ID, "", cfg.Commands)
	if err != nil {
		return nil, fmt.Errorf("%s registering commands: %w", op, err)
	}
	return &Bot{
		log:                 log,
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
	b.log.Info(ctx, "stopping bot")
	if b.removeCommands {
		for _, v := range b.registeredCommands {
			if err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", v.ID); err != nil {
				b.log.Error(ctx, "cannot delete command", slog.String("command", v.Name), sl.Err(err))
			}
		}
	}
	return b.session.Close()
}
