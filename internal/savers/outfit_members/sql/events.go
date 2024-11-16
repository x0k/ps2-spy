package sql_outfit_members_saver

import (
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	OutfitMembersInitType   EventType = "outfit_members_init"
	OutfitMembersUpdateType EventType = "outfit_members_update"
)

type OutfitMembersInit struct {
	OutfitId ps2.OutfitId
	Members  []ps2.CharacterId
}

func (e OutfitMembersInit) Type() EventType {
	return OutfitMembersInitType
}

type OutfitMembersUpdate struct {
	OutfitId ps2.OutfitId
	Members  diff.Diff[ps2.CharacterId]
}

func (e OutfitMembersUpdate) Type() EventType {
	return OutfitMembersUpdateType
}
