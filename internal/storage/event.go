package storage

import (
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	ChannelOutfitSavedType      EventType = "channel_outfit_saved"
	ChannelOutfitDeletedType    EventType = "channel_outfit_deleted"
	ChannelCharacterSavedType   EventType = "channel_character_saved"
	ChannelCharacterDeletedType EventType = "channel_character_deleted"
	OutfitMemberSavedType       EventType = "outfit_member_saved"
	OutfitMemberDeletedType     EventType = "outfit_member_deleted"
	OutfitSynchronizedType      EventType = "outfit_synchronized"
)

type ChannelOutfitSaved struct {
	ChannelId discord.ChannelId
	Platform  ps2_platforms.Platform
	OutfitId  ps2.OutfitId
}

func (e ChannelOutfitSaved) Type() EventType {
	return ChannelOutfitSavedType
}

type ChannelOutfitDeleted struct {
	ChannelId discord.ChannelId
	Platform  ps2_platforms.Platform
	OutfitId  ps2.OutfitId
}

func (e ChannelOutfitDeleted) Type() EventType {
	return ChannelOutfitDeletedType
}

type ChannelCharacterSaved struct {
	ChannelId   discord.ChannelId
	Platform    ps2_platforms.Platform
	CharacterId ps2.CharacterId
}

func (e ChannelCharacterSaved) Type() EventType {
	return ChannelCharacterSavedType
}

type ChannelCharacterDeleted struct {
	ChannelId   discord.ChannelId
	Platform    ps2_platforms.Platform
	CharacterId ps2.CharacterId
}

func (e ChannelCharacterDeleted) Type() EventType {
	return ChannelCharacterDeletedType
}

type OutfitMemberSaved struct {
	Platform    ps2_platforms.Platform
	OutfitId    ps2.OutfitId
	CharacterId ps2.CharacterId
}

func (e OutfitMemberSaved) Type() EventType {
	return OutfitMemberSavedType
}

type OutfitMemberDeleted struct {
	Platform    ps2_platforms.Platform
	OutfitId    ps2.OutfitId
	CharacterId ps2.CharacterId
}

func (e OutfitMemberDeleted) Type() EventType {
	return OutfitMemberDeletedType
}

type OutfitSynchronized struct {
	Platform       ps2_platforms.Platform
	OutfitId       ps2.OutfitId
	SynchronizedAt time.Time
}

func (e OutfitSynchronized) Type() EventType {
	return OutfitSynchronizedType
}
