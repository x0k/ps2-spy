package storage

type eventHandler interface {
	Type() string
	Handle(e any)
}

type channelOutfitSavedHandler chan<- ChannelOutfitSaved

func (h channelOutfitSavedHandler) Type() string {
	return ChannelOutfitSavedType
}

func (h channelOutfitSavedHandler) Handle(e any) {
	h <- e.(ChannelOutfitSaved)
}

type channelOutfitDeletedHandler chan<- ChannelOutfitDeleted

func (h channelOutfitDeletedHandler) Type() string {
	return ChannelOutfitDeletedType
}

func (h channelOutfitDeletedHandler) Handle(e any) {
	h <- e.(ChannelOutfitDeleted)
}

type channelCharacterSavedHandler chan<- ChannelCharacterSaved

func (h channelCharacterSavedHandler) Type() string {
	return ChannelCharacterSavedType
}

func (h channelCharacterSavedHandler) Handle(e any) {
	h <- e.(ChannelCharacterSaved)
}

type channelCharacterDeletedHandler chan<- ChannelCharacterDeleted

func (h channelCharacterDeletedHandler) Type() string {
	return ChannelCharacterDeletedType
}

func (h channelCharacterDeletedHandler) Handle(e any) {
	h <- e.(ChannelCharacterDeleted)
}

type outfitMemberSavedHandler chan<- OutfitMemberSaved

func (h outfitMemberSavedHandler) Type() string {
	return OutfitMemberSavedType
}

func (h outfitMemberSavedHandler) Handle(e any) {
	h <- e.(OutfitMemberSaved)
}

type outfitMemberDeletedHandler chan<- OutfitMemberDeleted

func (h outfitMemberDeletedHandler) Type() string {
	return OutfitMemberDeletedType
}

func (h outfitMemberDeletedHandler) Handle(e any) {
	h <- e.(OutfitMemberDeleted)
}

type outfitSynchronizedHandler chan<- OutfitSynchronized

func (h outfitSynchronizedHandler) Type() string {
	return OutfitSynchronizedType
}

func (h outfitSynchronizedHandler) Handle(e any) {
	h <- e.(OutfitSynchronized)
}

func handlerForInterface(handler any) eventHandler {
	switch v := handler.(type) {
	case chan ChannelOutfitSaved:
		return channelOutfitSavedHandler(v)
	case chan<- ChannelOutfitSaved:
		return channelOutfitSavedHandler(v)
	case chan ChannelOutfitDeleted:
		return channelOutfitDeletedHandler(v)
	case chan<- ChannelOutfitDeleted:
		return channelOutfitDeletedHandler(v)
	case chan ChannelCharacterSaved:
		return channelCharacterSavedHandler(v)
	case chan<- ChannelCharacterSaved:
		return channelCharacterSavedHandler(v)
	case chan ChannelCharacterDeleted:
		return channelCharacterDeletedHandler(v)
	case chan<- ChannelCharacterDeleted:
		return channelCharacterDeletedHandler(v)
	case chan OutfitMemberSaved:
		return outfitMemberSavedHandler(v)
	case chan<- OutfitMemberSaved:
		return outfitMemberSavedHandler(v)
	case chan OutfitMemberDeleted:
		return outfitMemberDeletedHandler(v)
	case chan<- OutfitMemberDeleted:
		return outfitMemberDeletedHandler(v)
	case chan OutfitSynchronized:
		return outfitSynchronizedHandler(v)
	case chan<- OutfitSynchronized:
		return outfitSynchronizedHandler(v)
	}
	return nil
}
