package storage

type channelOutfitSavedHandler chan<- ChannelOutfitSaved

func (h channelOutfitSavedHandler) Type() EventType {
	return ChannelOutfitSavedType
}

func (h channelOutfitSavedHandler) Handle(e Event) {
	h <- e.(ChannelOutfitSaved)
}

type channelOutfitDeletedHandler chan<- ChannelOutfitDeleted

func (h channelOutfitDeletedHandler) Type() EventType {
	return ChannelOutfitDeletedType
}

func (h channelOutfitDeletedHandler) Handle(e Event) {
	h <- e.(ChannelOutfitDeleted)
}

type channelCharacterSavedHandler chan<- ChannelCharacterSaved

func (h channelCharacterSavedHandler) Type() EventType {
	return ChannelCharacterSavedType
}

func (h channelCharacterSavedHandler) Handle(e Event) {
	h <- e.(ChannelCharacterSaved)
}

type channelCharacterDeletedHandler chan<- ChannelCharacterDeleted

func (h channelCharacterDeletedHandler) Type() EventType {
	return ChannelCharacterDeletedType
}

func (h channelCharacterDeletedHandler) Handle(e Event) {
	h <- e.(ChannelCharacterDeleted)
}

type outfitMemberSavedHandler chan<- OutfitMemberSaved

func (h outfitMemberSavedHandler) Type() EventType {
	return OutfitMemberSavedType
}

func (h outfitMemberSavedHandler) Handle(e Event) {
	h <- e.(OutfitMemberSaved)
}

type outfitMemberDeletedHandler chan<- OutfitMemberDeleted

func (h outfitMemberDeletedHandler) Type() EventType {
	return OutfitMemberDeletedType
}

func (h outfitMemberDeletedHandler) Handle(e Event) {
	h <- e.(OutfitMemberDeleted)
}

type outfitSynchronizedHandler chan<- OutfitSynchronized

func (h outfitSynchronizedHandler) Type() EventType {
	return OutfitSynchronizedType
}

func (h outfitSynchronizedHandler) Handle(e Event) {
	h <- e.(OutfitSynchronized)
}
