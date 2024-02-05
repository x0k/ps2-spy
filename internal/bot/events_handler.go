package bot

import (
	"context"

	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/handlers/event/facility_control_event_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/event/facility_loss_event_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/event/login_event_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/event/logout_event_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/event/outfit_members_update_event_handler"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/savers/outfit_members_saver"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type EventHandlers struct {
	log                         *logger.Logger
	charactersTrackerPublisher  *characters_tracker.Publisher
	playerLoginHandler          handlers.Ps2EventHandler[characters_tracker.PlayerLogin]
	playerLogoutHandler         handlers.Ps2EventHandler[characters_tracker.PlayerLogout]
	outfitMembersSaverPublisher *outfit_members_saver.Publisher
	outfitMembersUpdateHandler  handlers.Ps2EventHandler[outfit_members_saver.OutfitMembersUpdate]
	worldsTrackerPublisher      *worlds_tracker.Publisher
	facilityControlHandler      handlers.Ps2EventHandler[worlds_tracker.FacilityControl]
	facilityLossHandler         handlers.Ps2EventHandler[worlds_tracker.FacilityLoss]
}

func (eh *EventHandlers) Start(
	ctx context.Context,
	eventHandlersConfig *handlers.Ps2EventHandlerConfig,
) error {
	const op = "bot.EventHandlers.Start"
	// Register event handlers
	playerLogin := make(chan characters_tracker.PlayerLogin)
	playerLoginUnSub := eh.charactersTrackerPublisher.AddPlayerLoginHandler(playerLogin)
	playerLogout := make(chan characters_tracker.PlayerLogout)
	playerLogoutUnSub := eh.charactersTrackerPublisher.AddPlayerLogoutHandler(playerLogout)
	outfitMembersUpdate := make(chan outfit_members_saver.OutfitMembersUpdate)
	outfitMembersUpdateUnSub := eh.outfitMembersSaverPublisher.AddOutfitMembersUpdateHandler(outfitMembersUpdate)
	facilityControl := make(chan worlds_tracker.FacilityControl)
	facilityControlUnSub := eh.worldsTrackerPublisher.AddFacilityControlHandler(facilityControl)
	facilityLoss := make(chan worlds_tracker.FacilityLoss)
	facilityLossUnSub := eh.worldsTrackerPublisher.AddFacilityLossHandler(facilityLoss)
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
				go eh.playerLoginHandler.Run(ctx, eh.log, eventHandlersConfig, e)
			case e := <-playerLogout:
				go eh.playerLogoutHandler.Run(ctx, eh.log, eventHandlersConfig, e)
			case e := <-outfitMembersUpdate:
				go eh.outfitMembersUpdateHandler.Run(ctx, eh.log, eventHandlersConfig, e)
			case e := <-facilityControl:
				go eh.facilityControlHandler.Run(ctx, eh.log, eventHandlersConfig, e)
			case e := <-facilityLoss:
				go eh.facilityLossHandler.Run(ctx, eh.log, eventHandlersConfig, e)
			}
		}
	}()
	return nil
}

func NewEventHandlers(
	log *logger.Logger,
	charactersTrackerPublisher *characters_tracker.Publisher,
	outfitMembersSaverPublisher *outfit_members_saver.Publisher,
	worldsTrackerPublisher *worlds_tracker.Publisher,
	characterLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character],
	outfitLoader loaders.KeyedLoader[ps2.OutfitId, ps2.Outfit],
	facilityLoader loaders.KeyedLoader[ps2.FacilityId, ps2.Facility],
	charactersLoader loaders.QueriedLoader[[]ps2.CharacterId, map[ps2.CharacterId]ps2.Character],
) EventHandlers {
	return EventHandlers{
		log:                         log,
		charactersTrackerPublisher:  charactersTrackerPublisher,
		outfitMembersSaverPublisher: outfitMembersSaverPublisher,
		worldsTrackerPublisher:      worldsTrackerPublisher,
		playerLoginHandler:          login_event_handler.New(characterLoader),
		playerLogoutHandler:         logout_event_handler.New(characterLoader),
		outfitMembersUpdateHandler: outfit_members_update_event_handler.New(
			log,
			outfitLoader,
			charactersLoader,
		),
		facilityControlHandler: facility_control_event_handler.New(
			outfitLoader,
			facilityLoader,
		),
		facilityLossHandler: facility_loss_event_handler.New(
			outfitLoader,
			facilityLoader,
		),
	}
}
