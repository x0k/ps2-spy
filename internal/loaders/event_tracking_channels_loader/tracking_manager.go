package event_tracking_channels_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
)

type EventTrackingChannelsLoader struct {
	trackingManager *tracking_manager.TrackingManager
}

func New(trackingManager *tracking_manager.TrackingManager) *EventTrackingChannelsLoader {
	return &EventTrackingChannelsLoader{
		trackingManager: trackingManager,
	}
}

func (l *EventTrackingChannelsLoader) Load(ctx context.Context, event any) ([]meta.ChannelId, error) {
	return l.trackingManager.ChannelIds(ctx, event)
}
