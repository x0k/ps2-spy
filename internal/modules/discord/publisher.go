package discord_module

import (
	"context"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
)

type Publisher struct {
	pubsub.Publisher[Event]
	trackingManager *tracking_manager.TrackingManager
}

func NewPublisher(publisher pubsub.Publisher[Event], trackingManager *tracking_manager.TrackingManager) *Publisher {
	return &Publisher{
		Publisher:       publisher,
		trackingManager: trackingManager,
	}
}

func (h *Publisher) PublishPlayerLogin(ctx context.Context, e characters_tracker.PlayerLogin) error {
	channels, err := h.trackingManager.ChannelIdsForCharacter(ctx, e.CharacterId)
	if err != nil {
		return err
	}
	if len(channels) == 0 {
		return nil
	}
	return h.Publish(PlayerLogin(e))
}
