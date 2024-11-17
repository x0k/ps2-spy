package discord_module

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
)

type EventsHandler struct {
	log             *logger.Logger
	session         *discordgo.Session
	trackingManager *tracking_manager.TrackingManager
}

func (h *EventsHandler) HandlePlayerLogin(ctx context.Context, e characters_tracker.PlayerLogin) {
	channels, err := h.trackingManager.ChannelIdsForCharacter(ctx, e.CharacterId)
	if err != nil {
		h.log.Error(ctx, "failed to get channels for character", sl.Err(err))
		return
	}
	if len(channels) == 0 {
		return
	}

}
