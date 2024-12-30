package ps2

import (
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	OutfitMembersUpdateType  EventType = "outfit_members_update"
	OutfitMembersAddedType   EventType = "outfit_members_added"
	OutfitMembersRemovedType EventType = "outfit_members_removed"
)

type OutfitMembersUpdate struct {
	Platform ps2_platforms.Platform
	OutfitId OutfitId
	Members  diff.Diff[CharacterId]
}

func (e OutfitMembersUpdate) Type() EventType {
	return OutfitMembersUpdateType
}

type OutfitMembersAdded struct {
	Platform     ps2_platforms.Platform
	OutfitId     OutfitId
	CharacterIds []CharacterId
}

func (e OutfitMembersAdded) Type() EventType {
	return OutfitMembersAddedType
}

type OutfitMembersRemoved struct {
	Platform     ps2_platforms.Platform
	OutfitId     OutfitId
	CharacterIds []CharacterId
}

func (e OutfitMembersRemoved) Type() EventType {
	return OutfitMembersRemovedType
}
