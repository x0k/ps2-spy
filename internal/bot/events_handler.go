package bot

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/handlers/event/facility_control_event_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/event/facility_loss_event_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/event/login_event_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/event/logout_event_handler"
	"github.com/x0k/ps2-spy/internal/bot/handlers/event/outfit_members_update_event_handler"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/savers/outfit_members_saver"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type EventHandlers struct {
	charactersTrackerPublisher  *publisher.Publisher
	playerLoginHandler          handlers.Ps2EventHandler[characters_tracker.PlayerLogin]
	playerLogoutHandler         handlers.Ps2EventHandler[characters_tracker.PlayerLogout]
	outfitMembersSaverPublisher *publisher.Publisher
	outfitMembersUpdateHandler  handlers.Ps2EventHandler[outfit_members_saver.OutfitMembersUpdate]
	worldsTrackerPublisher      *publisher.Publisher
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
	playerLoginUnSub, err := eh.charactersTrackerPublisher.AddHandler(playerLogin)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	playerLogout := make(chan characters_tracker.PlayerLogout)
	playerLogoutUnSub, err := eh.charactersTrackerPublisher.AddHandler(playerLogout)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	outfitMembersUpdate := make(chan outfit_members_saver.OutfitMembersUpdate)
	outfitMembersUpdateUnSub, err := eh.outfitMembersSaverPublisher.AddHandler(outfitMembersUpdate)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	facilityControl := make(chan worlds_tracker.FacilityControl)
	facilityControlUnSub, err := eh.worldsTrackerPublisher.AddHandler(facilityControl)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	facilityLoss := make(chan worlds_tracker.FacilityLoss)
	facilityLossUnSub, err := eh.worldsTrackerPublisher.AddHandler(facilityLoss)
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
				go eh.playerLoginHandler.Run(ctx, eventHandlersConfig, e)
			case e := <-playerLogout:
				go eh.playerLogoutHandler.Run(ctx, eventHandlersConfig, e)
			case e := <-outfitMembersUpdate:
				go eh.outfitMembersUpdateHandler.Run(ctx, eventHandlersConfig, e)
			case e := <-facilityControl:
				go eh.facilityControlHandler.Run(ctx, eventHandlersConfig, e)
			case e := <-facilityLoss:
				go eh.facilityLossHandler.Run(ctx, eventHandlersConfig, e)
			}
		}
	}()
	return nil
}

func NewEventHandlers(
	charactersTrackerPublisher *publisher.Publisher,
	outfitMembersSaverPublisher *publisher.Publisher,
	worldsTrackerPublisher *publisher.Publisher,
	characterLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character],
	outfitLoader loaders.KeyedLoader[ps2.OutfitId, ps2.Outfit],
	facilityLoader loaders.KeyedLoader[ps2.FacilityId, ps2.Facility],
	charactersLoader loaders.QueriedLoader[[]ps2.CharacterId, map[ps2.CharacterId]ps2.Character],
) EventHandlers {
	return EventHandlers{
		charactersTrackerPublisher:  charactersTrackerPublisher,
		outfitMembersSaverPublisher: outfitMembersSaverPublisher,
		worldsTrackerPublisher:      worldsTrackerPublisher,
		playerLoginHandler:          login_event_handler.New(characterLoader),
		playerLogoutHandler:         logout_event_handler.New(characterLoader),
		outfitMembersUpdateHandler: outfit_members_update_event_handler.New(
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
