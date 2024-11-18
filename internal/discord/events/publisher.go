package discord_events

import (
	"context"
	"log/slog"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
)

type HandlersManager struct {
	name            string
	log             *logger.Logger
	handlers        map[discord.EventType]discord.Handler
	session         *discordgo.Session
	wg              sync.WaitGroup
	trackingManager *tracking_manager.TrackingManager
}

func NewHandlersManager(
	name string,
	log *logger.Logger,
	handlers map[discord.EventType]discord.Handler,
	trackingManager *tracking_manager.TrackingManager,
) *HandlersManager {
	return &HandlersManager{
		name:            name,
		log:             log,
		handlers:        handlers,
		trackingManager: trackingManager,
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

func (h *HandlersManager) handleCharacterEventTask(
	ctx context.Context,
	characterId ps2.CharacterId,
	e discord.Event,
) {
	defer h.wg.Done()
	t := e.Type()
	handler, ok := h.handlers[t]
	if !ok {
		return
	}
	channels, err := h.trackingManager.ChannelIdsForCharacter(ctx, characterId)
	if err != nil {
		h.log.Error(ctx, "cannot get channels for character", slog.String("character_id", string(characterId)), sl.Err(err))
	}
	if len(channels) == 0 {
		return
	}
	if err := handler(ctx, h.session, channels, e); err != nil {
		h.log.Error(ctx, "cannot handle event", sl.Err(err))
	}
}
