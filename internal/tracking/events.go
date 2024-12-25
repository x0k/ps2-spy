package tracking

import (
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	TrackingSettingsUpdatedType EventType = "tracking_settings_updated"
)

type TrackingSettingsUpdated struct {
	ChannelId  discord.ChannelId
	Platform   ps2_platforms.Platform
	Outfits    diff.Diff[ps2.OutfitId]
	Characters diff.Diff[ps2.CharacterId]
}

func (e TrackingSettingsUpdated) Type() EventType {
	return TrackingSettingsUpdatedType
}
