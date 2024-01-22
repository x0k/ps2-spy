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
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/publisher"
	"github.com/x0k/ps2-spy/internal/savers/outfit_members_saver"
)

var ErrEventTrackingChannelsLoaderNotFound = fmt.Errorf("event tracking channels loader not found")
var ErrEventsPublisherNotFound = fmt.Errorf("events publisher not found")
var ErrEventHandlerNotFound = fmt.Errorf("event handler not found")

type Bot struct {
	session            *discordgo.Session
	registeredCommands []*discordgo.ApplicationCommand
}

type BotConfig struct {
	DiscordToken                 string
	CommandHandlerTimeout        time.Duration
	Ps2EventHandlerTimeout       time.Duration
	Commands                     []*discordgo.ApplicationCommand
	CommandHandlers              map[string]handlers.InteractionHandler
	SubmitHandlers               map[string]handlers.InteractionHandler
	EventTrackingChannelsLoaders map[string]loaders.QueriedLoader[any, []string]
	// Raw PS2 events
	Ps2EventsPublishers  map[string]*ps2events.Publisher
	PlayerLoginHandlers  map[string]handlers.Ps2EventHandler[ps2events.PlayerLogin]
	PlayerLogoutHandlers map[string]handlers.Ps2EventHandler[ps2events.PlayerLogout]
	// Outfit events
	OutfitMembersSaverPublishers map[string]*publisher.Publisher
	OutfitMembersUpdateHandlers  map[string]handlers.Ps2EventHandler[outfit_members_saver.OutfitMembersUpdate]
	// Facility events
	FacilitiesManagerPublishers map[string]*publisher.Publisher
	FacilityControlHandlers     map[string]handlers.Ps2EventHandler[facilities_manager.FacilityControl]
	FacilityLossHandlers        map[string]handlers.Ps2EventHandler[facilities_manager.FacilityLoss]
}

func startEventHandlersForPlatform(
	ctx context.Context,
	session *discordgo.Session,
	cfg *BotConfig,
	platform string,
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
				go handler.Run(ctx, cfg.CommandHandlerTimeout, s, i)
			} else {
				l.Warn("unknown command")
			}
		case discordgo.InteractionMessageComponent:
			l.Debug("component invoked")
		case discordgo.InteractionModalSubmit:
			data := i.ModalSubmitData()
			l.Debug("modal submitted", slog.Any("data", data))
			if handler, ok := cfg.SubmitHandlers[data.CustomID]; ok {
				go handler.Run(ctx, cfg.CommandHandlerTimeout, s, i)
			} else {
				l.Warn("unknown modal")
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
		session:            session,
		registeredCommands: registeredCommands,
	}, nil
}

func (b *Bot) Stop(ctx context.Context) error {
	const op = "bot.Bot.Stop"
	log := infra.OpLogger(ctx, op)
	log.Info("stopping bot")
	for _, v := range b.registeredCommands {
		err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", v.ID)
		if err != nil {
			log.Error("cannot delete command", slog.String("command", v.Name), sl.Err(err))
		}
	}
	return b.session.Close()
}
