package storage

import (
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"golang.org/x/text/language"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	ChannelOutfitSavedType                 EventType = "channel_outfit_saved"
	ChannelOutfitDeletedType               EventType = "channel_outfit_deleted"
	ChannelCharacterSavedType              EventType = "channel_character_saved"
	ChannelCharacterDeletedType            EventType = "channel_character_deleted"
	OutfitMembersInitType                  EventType = "outfit_members_init"
	OutfitMembersUpdateType                EventType = "outfit_members_update"
	OutfitMemberSavedType                  EventType = "outfit_member_saved"
	OutfitMemberDeletedType                EventType = "outfit_member_deleted"
	OutfitSynchronizedType                 EventType = "outfit_synchronized"
	ChannelLanguageSavedType               EventType = "channel_language_saved"
	ChannelCharacterNotificationsSavedType EventType = "channel_character_notifications_saved"
	ChannelOutfitNotificationsSavedType    EventType = "channel_outfit_notifications_saved"
	ChannelTitleUpdatesSavedType           EventType = "channel_title_updates_saved"
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

type OutfitMembersInit struct {
	Platform ps2_platforms.Platform
	OutfitId ps2.OutfitId
	Members  []ps2.CharacterId
}

func (e OutfitMembersInit) Type() EventType {
	return OutfitMembersInitType
}

type OutfitMembersUpdate struct {
	Platform ps2_platforms.Platform
	OutfitId ps2.OutfitId
	Members  diff.Diff[ps2.CharacterId]
}

func (e OutfitMembersUpdate) Type() EventType {
	return OutfitMembersUpdateType
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

type ChannelLanguageSaved struct {
	ChannelId discord.ChannelId
	Language  language.Tag
}

func (e ChannelLanguageSaved) Type() EventType {
	return ChannelLanguageSavedType
}

type ChannelCharacterNotificationsSaved struct {
	ChannelId discord.ChannelId
	Enabled   bool
}

func (e ChannelCharacterNotificationsSaved) Type() EventType {
	return ChannelCharacterNotificationsSavedType
}

type ChannelOutfitNotificationsSaved struct {
	ChannelId discord.ChannelId
	Enabled   bool
}

func (e ChannelOutfitNotificationsSaved) Type() EventType {
	return ChannelOutfitNotificationsSavedType
}

type ChannelTitleUpdatesSaved struct {
	ChannelId discord.ChannelId
	Enabled   bool
}

func (e ChannelTitleUpdatesSaved) Type() EventType {
	return ChannelTitleUpdatesSavedType
}
