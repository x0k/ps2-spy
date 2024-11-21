package discord_events

import (
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	sql_outfit_members_saver "github.com/x0k/ps2-spy/internal/savers/outfit_members/sql"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	PlayerLoginType         EventType = "player_login"
	PlayerLogoutType        EventType = "player_logout"
	OutfitMembersUpdateType EventType = "outfit_members_update"
	FacilityControlType     EventType = "facility_control"
	FacilityLossType        EventType = "facility_loss"
)

type PlayerLogin characters_tracker.PlayerLogin

func (e PlayerLogin) Type() EventType {
	return PlayerLoginType
}

type PlayerLogout characters_tracker.PlayerLogout

func (e PlayerLogout) Type() EventType {
	return PlayerLoginType
}

type OutfitMembersUpdate sql_outfit_members_saver.OutfitMembersUpdate

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
