package discord_events

import (
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
	"golang.org/x/text/language"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	PlayerLoginType            = EventType(characters_tracker.PlayerLoginType)
	PlayerLogoutType           = EventType(characters_tracker.PlayerLogoutType)
	OutfitMembersUpdateType    = EventType(storage.OutfitMembersUpdateType)
	FacilityControlType        = EventType(worlds_tracker.FacilityControlType)
	FacilityLossType           = EventType(worlds_tracker.FacilityLossType)
	ChannelLanguageUpdatedType = EventType(storage.ChannelLanguageUpdatedType)
	ChannelTrackerStartedType  = EventType(stats_tracker.ChannelTrackerStartedType)
	ChannelTrackerStoppedType  = EventType(stats_tracker.ChannelTrackerStoppedType)
)

type adoptedEvent[T pubsub.EventType, E pubsub.Event[T]] struct {
	Event E
}

func (e adoptedEvent[T, Event]) Type() EventType {
	return EventType(e.Event.Type())
}

type channelsEvent[T pubsub.EventType, E pubsub.Event[T]] struct {
	Event    E
	Channels []discord.Channel
}

func (e channelsEvent[T, Event]) Type() EventType {
	return EventType(e.Event.Type())
}

type localizedEvent[T pubsub.EventType, E pubsub.Event[T]] struct {
	Event    E
	Language language.Tag
}

func (e localizedEvent[T, Event]) Type() EventType {
	return EventType(e.Event.Type())
}

type ChannelLanguageUpdated = adoptedEvent[storage.EventType, storage.ChannelLanguageUpdated]

type ChannelTrackerStarted = localizedEvent[stats_tracker.EventType, stats_tracker.ChannelTrackerStarted]
type ChannelTrackerStopped = localizedEvent[stats_tracker.EventType, stats_tracker.ChannelTrackerStopped]

type PlayerLogin = channelsEvent[characters_tracker.EventType, characters_tracker.PlayerLogin]
type PlayerLogout = channelsEvent[characters_tracker.EventType, characters_tracker.PlayerLogout]
type OutfitMembersUpdate = channelsEvent[storage.EventType, storage.OutfitMembersUpdate]
type FacilityControl = channelsEvent[worlds_tracker.EventType, worlds_tracker.FacilityControl]
type FacilityLoss = channelsEvent[worlds_tracker.EventType, worlds_tracker.FacilityLoss]
