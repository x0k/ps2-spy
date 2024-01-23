package bot

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/facilities_manager"
	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/publisher"
	"github.com/x0k/ps2-spy/internal/savers/outfit_members_saver"
)

var ErrEventTrackingChannelsLoaderNotFound = fmt.Errorf("event tracking channels loader not found")
var ErrEventsPublisherNotFound = fmt.Errorf("events publisher not found")
var ErrEventHandlerNotFound = fmt.Errorf("event handler not found")

type Bot struct {
	session            *discordgo.Session
	removeCommands     bool
	registeredCommands []*discordgo.ApplicationCommand
}

type BotConfig struct {
	DiscordToken                 string
	RemoveCommands               bool
	CommandHandlerTimeout        time.Duration
	Ps2EventHandlerTimeout       time.Duration
	Commands                     []*discordgo.ApplicationCommand
	CommandHandlers              map[string]handlers.InteractionHandler
	SubmitHandlers               map[string]handlers.InteractionHandler
	EventTrackingChannelsLoaders map[platforms.Platform]loaders.QueriedLoader[any, []meta.ChannelId]
	// Raw PS2 events
	Ps2EventsPublishers  map[platforms.Platform]*ps2events.Publisher
	PlayerLoginHandlers  map[platforms.Platform]handlers.Ps2EventHandler[ps2events.PlayerLogin]
	PlayerLogoutHandlers map[platforms.Platform]handlers.Ps2EventHandler[ps2events.PlayerLogout]
	// Outfit events
	OutfitMembersSaverPublishers map[platforms.Platform]*publisher.Publisher
	OutfitMembersUpdateHandlers  map[platforms.Platform]handlers.Ps2EventHandler[outfit_members_saver.OutfitMembersUpdate]
	// Facility events
	FacilitiesManagerPublishers map[platforms.Platform]*publisher.Publisher
	FacilityControlHandlers     map[platforms.Platform]handlers.Ps2EventHandler[facilities_manager.FacilityControl]
	FacilityLossHandlers        map[platforms.Platform]handlers.Ps2EventHandler[facilities_manager.FacilityLoss]
}

func startEventHandlersForPlatform(
	ctx context.Context,
	session *discordgo.Session,
	cfg *BotConfig,
	platform platforms.Platform,
) error {
	const op = "bot.startEventHandlersForPlatform"
	eventTrackingChannelsLoader, ok := cfg.EventTrackingChannelsLoaders[platform]
	if !ok {
		return fmt.Errorf("%s get event tracking channels loader: %w", platform, ErrEventsPublisherNotFound)
	}
	eventHandlersConfig := &handlers.Ps2EventHandlerConfig{
		Session:                     session,
		Timeout:                     cfg.Ps2EventHandlerTimeout,
		EventTrackingChannelsLoader: eventTrackingChannelsLoader,
	}
	// PS2 Events
	eventsPublisher, ok := cfg.Ps2EventsPublishers[platform]
	if !ok {
		return fmt.Errorf("%s get events publisher: %w", platform, ErrEventsPublisherNotFound)
	}
	playerLoginHandler, ok := cfg.PlayerLoginHandlers[platform]
	if !ok {
		return fmt.Errorf("%s get player login handler: %w", platform, ErrEventHandlerNotFound)
	}
	playerLogoutHandler, ok := cfg.PlayerLogoutHandlers[platform]
	if !ok {
		return fmt.Errorf("%s get player logout handler: %w", platform, ErrEventHandlerNotFound)
	}
	// Outfits
	outfitMembersUpdateHandler, ok := cfg.OutfitMembersUpdateHandlers[platform]
	if !ok {
		return fmt.Errorf("%s get outfit member join handler: %w", platform, ErrEventHandlerNotFound)
	}
	outfitMembersSaverPublisher, ok := cfg.OutfitMembersSaverPublishers[platform]
	if !ok {
		return fmt.Errorf("%s get outfit members saver publisher: %w", platform, ErrEventHandlerNotFound)
	}
	// Facilities
	facilitiesManagerPublisher, ok := cfg.FacilitiesManagerPublishers[platform]
	if !ok {
		return fmt.Errorf("%s get facilities manager publisher: %w", platform, ErrEventHandlerNotFound)
	}
	facilityControlHandler, ok := cfg.FacilityControlHandlers[platform]
	if !ok {
		return fmt.Errorf("%s get facility control handler: %w", platform, ErrEventHandlerNotFound)
	}
	facilitiesLossHandler, ok := cfg.FacilityLossHandlers[platform]
	if !ok {
		return fmt.Errorf("%s get facility loss handler: %w", platform, ErrEventHandlerNotFound)
	}
	// Register event handlers
	playerLogin := make(chan ps2events.PlayerLogin)
	playerLoginUnSub, err := eventsPublisher.AddHandler(playerLogin)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	playerLogout := make(chan ps2events.PlayerLogout)
	playerLogoutUnSub, err := eventsPublisher.AddHandler(playerLogout)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	outfitMembersUpdate := make(chan outfit_members_saver.OutfitMembersUpdate)
	outfitMembersUpdateUnSub, err := outfitMembersSaverPublisher.AddHandler(outfitMembersUpdate)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	facilityControl := make(chan facilities_manager.FacilityControl)
	facilityControlUnSub, err := facilitiesManagerPublisher.AddHandler(facilityControl)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	facilityLoss := make(chan facilities_manager.FacilityLoss)
	facilityLossUnSub, err := facilitiesManagerPublisher.AddHandler(facilityLoss)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	wg := infra.Wg(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer playerLoginUnSub()
		defer playerLogoutUnSub()
		defer facilityControlUnSub()
		defer outfitMembersUpdateUnSub()
		defer facilityLossUnSub()
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-playerLogin:
				// TODO: add handlers to wait group
				go playerLoginHandler.Run(ctx, eventHandlersConfig, e)
			case e := <-playerLogout:
				go playerLogoutHandler.Run(ctx, eventHandlersConfig, e)
			case e := <-outfitMembersUpdate:
				go outfitMembersUpdateHandler.Run(ctx, eventHandlersConfig, e)
			case e := <-facilityControl:
				go facilityControlHandler.Run(ctx, eventHandlersConfig, e)
			case e := <-facilityLoss:
				go facilitiesLossHandler.Run(ctx, eventHandlersConfig, e)
			}
		}
	}()
	return nil
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

	for _, p := range platforms.Platforms {
		err = startEventHandlersForPlatform(ctx, session, cfg, p)
		if err != nil {
			return nil, fmt.Errorf("%s start event handlers for %q: %w", op, p, err)
		}
	}
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
		session:            session,
		removeCommands:     cfg.RemoveCommands,
		registeredCommands: registeredCommands,
	}, nil
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
