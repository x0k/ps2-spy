package discord_events

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type HandlersManager struct {
	name            string
	log             *logger.Logger
	handlers        map[EventType][]Handler
	session         *discordgo.Session
	wg              sync.WaitGroup
	trackingManager *tracking_manager.TrackingManager
	handlersTimeout time.Duration
}

func NewHandlersManager(
	name string,
	log *logger.Logger,
	session *discordgo.Session,
	handlers map[EventType][]Handler,
	trackingManager *tracking_manager.TrackingManager,
	handlersTimeout time.Duration,
) *HandlersManager {
	return &HandlersManager{
		name:            name,
		log:             log,
		session:         session,
		handlers:        handlers,
		trackingManager: trackingManager,
		handlersTimeout: handlersTimeout,
	}
}

func (h *HandlersManager) Name() string {
	return h.name
}

func (h *HandlersManager) Start(ctx context.Context) error {
	<-ctx.Done()
	h.wg.Wait()
	return nil
}

func (h *HandlersManager) HandlePlayerLogin(ctx context.Context, e characters_tracker.PlayerLogin) {
	h.wg.Add(1)
	go h.handleCharacterEventTask(ctx, e.CharacterId, PlayerLogin(e))
}

func (h *HandlersManager) HandlePlayerLogout(ctx context.Context, e characters_tracker.PlayerLogout) {
	h.wg.Add(1)
	go h.handleCharacterEventTask(ctx, e.CharacterId, PlayerLogout(e))
}

func (h *HandlersManager) HandleOutfitMembersUpdate(ctx context.Context, e storage.OutfitMembersUpdate) {
	h.wg.Add(1)
	go h.handleOutfitEventTask(ctx, e.OutfitId, OutfitMembersUpdate(e))
}

func (h *HandlersManager) HandleFacilityControl(ctx context.Context, e worlds_tracker.FacilityControl) {
	h.wg.Add(1)
	go h.handleOutfitEventTask(ctx, ps2.OutfitId(e.OutfitID), FacilityControl(e))
}

func (h *HandlersManager) HandleFacilityLoss(ctx context.Context, e worlds_tracker.FacilityLoss) {
	h.wg.Add(1)
	go h.handleOutfitEventTask(ctx, e.OldOutfitId, FacilityLoss(e))
}

func (h *HandlersManager) HandleChannelLanguageUpdate(ctx context.Context, e storage.ChannelLanguageUpdated) {
	handlers, ok := h.handlers[ChannelLanguageUpdatedType]
	if !ok {
		h.log.Debug(ctx, "no handler for event", slog.String("event_type", string(ChannelLanguageUpdatedType)))
		return
	}
	h.runHandlers(ctx, handlers, []discord.Channel{discord.NewChannel(e.ChannelId, e.Language)}, ChannelLanguageUpdated(e))
}

func (h *HandlersManager) handleCharacterEventTask(
	ctx context.Context,
	characterId ps2.CharacterId,
	e Event,
) {
	defer h.wg.Done()
	handlers, ok := h.handlers[e.Type()]
	if !ok {
		h.log.Debug(ctx, "no handler for event", slog.String("event_type", string(e.Type())))
		return
	}
	channels, err := h.trackingManager.ChannelIdsForCharacter(ctx, characterId)
	if err != nil {
		h.log.Error(ctx, "cannot get channels for character", slog.String("character_id", string(characterId)), sl.Err(err))
	}
	h.runHandlers(ctx, handlers, channels, e)
}

func (h *HandlersManager) handleOutfitEventTask(
	ctx context.Context,
	outfitId ps2.OutfitId,
	e Event,
) {
	defer h.wg.Done()
	handlers, ok := h.handlers[e.Type()]
	if !ok {
		h.log.Debug(ctx, "no handler for event", slog.String("event_type", string(e.Type())))
		return
	}
	channels, err := h.trackingManager.ChannelIdsForOutfit(ctx, outfitId)
	if err != nil {
		h.log.Error(ctx, "cannot get channels for outfit", slog.String("outfit_id", string(outfitId)), sl.Err(err))
	}
	h.runHandlers(ctx, handlers, channels, e)
}

func (h *HandlersManager) runHandlers(
	ctx context.Context,
	handlers []Handler,
	channels []discord.Channel,
	e Event,
) {
	if len(channels) == 0 {
		return
	}
	for _, handler := range handlers {
		h.wg.Add(1)
		go func(handler Handler) {
			defer h.wg.Done()
			ctx, cancel := context.WithTimeout(ctx, h.handlersTimeout)
			defer cancel()
			if err := handler(ctx, h.session, channels, e); err != nil {
				h.log.Error(ctx, "cannot handle event", sl.Err(err))
			}
		}(handler)
	}
}
