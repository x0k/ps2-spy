package storage

import (
	"time"

	"github.com/x0k/ps2-spy/internal/event"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

const (
	ChannelOutfitSavedType      event.Type = "channel_outfit_saved"
	ChannelOutfitDeletedType    event.Type = "channel_outfit_deleted"
	ChannelCharacterSavedType   event.Type = "channel_character_saved"
	ChannelCharacterDeletedType event.Type = "channel_character_deleted"
	OutfitMemberSavedType       event.Type = "outfit_member_saved"
	OutfitMemberDeletedType     event.Type = "outfit_member_deleted"
	OutfitSynchronizedType      event.Type = "outfit_synchronized"
)

type ChannelOutfitSaved struct {
	ChannelId meta.ChannelId
	Platform  platforms.Platform
	OutfitId  ps2.OutfitId
}

func (e ChannelOutfitSaved) Type() event.Type {
	return ChannelOutfitSavedType
}

type ChannelOutfitDeleted struct {
	ChannelId meta.ChannelId
	Platform  platforms.Platform
	OutfitId  ps2.OutfitId
}

func (e ChannelOutfitDeleted) Type() event.Type {
	return ChannelOutfitDeletedType
}

type ChannelCharacterSaved struct {
	ChannelId   meta.ChannelId
	Platform    platforms.Platform
	CharacterId ps2.CharacterId
}

func (e ChannelCharacterSaved) Type() event.Type {
	return ChannelCharacterSavedType
}

type ChannelCharacterDeleted struct {
	ChannelId   meta.ChannelId
	Platform    platforms.Platform
	CharacterId ps2.CharacterId
}

func (e ChannelCharacterDeleted) Type() event.Type {
	return ChannelCharacterDeletedType
}

type OutfitMemberSaved struct {
	Platform    platforms.Platform
	OutfitId    ps2.OutfitId
	CharacterId ps2.CharacterId
}

func (e OutfitMemberSaved) Type() event.Type {
	return OutfitMemberSavedType
}

type OutfitMemberDeleted struct {
	Platform    platforms.Platform
	OutfitId    ps2.OutfitId
	CharacterId ps2.CharacterId
}

func (e OutfitMemberDeleted) Type() event.Type {
	return OutfitMemberDeletedType
}

type OutfitSynchronized struct {
	Platform       platforms.Platform
	OutfitId       ps2.OutfitId
	SynchronizedAt time.Time
}

func (e OutfitSynchronized) Type() event.Type {
	return OutfitSynchronizedType
}
