package storage

const (
	ChannelOutfitSavedType      = "channel_outfit_saved"
	ChannelOutfitDeletedType    = "channel_outfit_deleted"
	ChannelCharacterSavedType   = "channel_character_saved"
	ChannelCharacterDeletedType = "channel_character_deleted"
)

type Event interface {
	Type() string
}

type ChannelOutfitSaved struct {
	ChannelId string
	Platform  string
	OutfitId  string
}

func (e ChannelOutfitSaved) Type() string {
	return ChannelOutfitSavedType
}

type ChannelOutfitDeleted struct {
	ChannelId string
	Platform  string
	OutfitId  string
}

func (e ChannelOutfitDeleted) Type() string {
	return ChannelOutfitDeletedType
}

type ChannelCharacterSaved struct {
	ChannelId   string
	Platform    string
	CharacterId string
}

func (e ChannelCharacterSaved) Type() string {
	return ChannelCharacterSavedType
}

type ChannelCharacterDeleted struct {
	ChannelId   string
	Platform    string
	CharacterId string
}

func (e ChannelCharacterDeleted) Type() string {
	return ChannelCharacterDeletedType
}
