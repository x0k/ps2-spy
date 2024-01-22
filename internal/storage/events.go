package storage

import (
	"time"

	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

const (
	ChannelOutfitSavedType      = "channel_outfit_saved"
	ChannelOutfitDeletedType    = "channel_outfit_deleted"
	ChannelCharacterSavedType   = "channel_character_saved"
	ChannelCharacterDeletedType = "channel_character_deleted"
	OutfitMemberSavedType       = "outfit_member_saved"
	OutfitMemberDeletedType     = "outfit_member_deleted"
	OutfitSynchronizedType      = "outfit_synchronized"
)

type ChannelOutfitSaved struct {
	ChannelId string
	Platform  platforms.Platform
	OutfitId  ps2.OutfitId
}

func (e ChannelOutfitSaved) Type() string {
	return ChannelOutfitSavedType
}

type ChannelOutfitDeleted struct {
	ChannelId string
	Platform  platforms.Platform
	OutfitId  ps2.OutfitId
}

func (e ChannelOutfitDeleted) Type() string {
	return ChannelOutfitDeletedType
}

type ChannelCharacterSaved struct {
	ChannelId   string
	Platform    platforms.Platform
	CharacterId ps2.CharacterId
}

func (e ChannelCharacterSaved) Type() string {
	return ChannelCharacterSavedType
}

type ChannelCharacterDeleted struct {
	ChannelId   string
	Platform    platforms.Platform
	CharacterId ps2.CharacterId
}

func (e ChannelCharacterDeleted) Type() string {
	return ChannelCharacterDeletedType
}

type OutfitMemberSaved struct {
	Platform    platforms.Platform
	OutfitId    ps2.OutfitId
	CharacterId ps2.CharacterId
}

func (e OutfitMemberSaved) Type() string {
	return OutfitMemberSavedType
}

type OutfitMemberDeleted struct {
	Platform    platforms.Platform
	OutfitId    ps2.OutfitId
	CharacterId ps2.CharacterId
}

func (e OutfitMemberDeleted) Type() string {
	return OutfitMemberDeletedType
}

type OutfitSynchronized struct {
	Platform       platforms.Platform
	OutfitId       ps2.OutfitId
	SynchronizedAt time.Time
}

func (e OutfitSynchronized) Type() string {
	return OutfitSynchronizedType
}
