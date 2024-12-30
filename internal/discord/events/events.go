package discord_events

import (
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	PlayerLoginType                        = EventType(ps2.PlayerLoginType)
	PlayerLogoutType                       = EventType(ps2.PlayerLogoutType)
	OutfitMembersUpdateType                = EventType(ps2.OutfitMembersUpdateType)
	FacilityControlType                    = EventType(worlds_tracker.FacilityControlType)
	FacilityLossType                       = EventType(worlds_tracker.FacilityLossType)
	ChannelLanguageUpdatedType             = EventType(storage.ChannelLanguageSavedType)
	ChannelCharacterNotificationsSavedType = EventType(storage.ChannelCharacterNotificationsSavedType)
	ChannelOutfitNotificationsSavedType    = EventType(storage.ChannelOutfitNotificationsSavedType)
	ChannelTitleUpdatesSavedType           = EventType(storage.ChannelTitleUpdatesSavedType)
	ChannelTrackerStartedType              = EventType(stats_tracker.ChannelTrackerStartedType)
	ChannelTrackerStoppedType              = EventType(stats_tracker.ChannelTrackerStoppedType)
	ChannelTrackingSettingsUpdatedType     = EventType(tracking.TrackingSettingsUpdatedType)
)

type channelsEvent[T pubsub.EventType, E pubsub.Event[T]] struct {
	Event    E
	Channels []discord.Channel
}

func (e channelsEvent[T, Event]) Type() EventType {
	return EventType(e.Event.Type())
}

type channelEvent[T pubsub.EventType, E pubsub.Event[T]] struct {
	Event   E
	Channel discord.Channel
}

func (e channelEvent[T, Event]) Type() EventType {
	return EventType(e.Event.Type())
}

type ChannelLanguageSaved = channelEvent[storage.EventType, storage.ChannelLanguageSaved]
type ChannelTitleUpdatesSaved = channelEvent[storage.EventType, storage.ChannelTitleUpdatesSaved]

type ChannelTrackerStarted = channelEvent[stats_tracker.EventType, stats_tracker.ChannelTrackerStarted]
type ChannelTrackerStopped = channelEvent[stats_tracker.EventType, stats_tracker.ChannelTrackerStopped]
type ChannelTrackingSettingsUpdated = channelEvent[tracking.EventType, tracking.TrackingSettingsUpdated]

type PlayerLogin = channelsEvent[ps2.EventType, ps2.PlayerLogin]
type PlayerFakeLogin = channelsEvent[ps2.EventType, ps2.PlayerFakeLogin]
type PlayerLogout = channelsEvent[ps2.EventType, ps2.PlayerLogout]
type OutfitMembersUpdate = channelsEvent[ps2.EventType, ps2.OutfitMembersUpdate]
type FacilityControl = channelsEvent[worlds_tracker.EventType, worlds_tracker.FacilityControl]
type FacilityLoss = channelsEvent[worlds_tracker.EventType, worlds_tracker.FacilityLoss]
