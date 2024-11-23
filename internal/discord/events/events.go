package discord_events

import (
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	PlayerLoginType            EventType = "player_login"
	PlayerLogoutType           EventType = "player_logout"
	OutfitMembersUpdateType    EventType = "outfit_members_update"
	FacilityControlType        EventType = "facility_control"
	FacilityLossType           EventType = "facility_loss"
	ChannelLanguageUpdatedType EventType = "channel_language_updated"
)

type PlayerLogin characters_tracker.PlayerLogin

func (e PlayerLogin) Type() EventType {
	return PlayerLoginType
}

type PlayerLogout characters_tracker.PlayerLogout

func (e PlayerLogout) Type() EventType {
	return PlayerLogoutType
}

type OutfitMembersUpdate storage.OutfitMembersUpdate

func (e OutfitMembersUpdate) Type() EventType {
	return OutfitMembersUpdateType
}

type FacilityControl worlds_tracker.FacilityControl

func (e FacilityControl) Type() EventType {
	return FacilityControlType
}

type FacilityLoss worlds_tracker.FacilityLoss

func (e FacilityLoss) Type() EventType {
	return FacilityLossType
}

type ChannelLanguageUpdated storage.ChannelLanguageUpdated

func (e ChannelLanguageUpdated) Type() EventType {
	return ChannelLanguageUpdatedType
}
