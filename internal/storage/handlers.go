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
	if t, ok := e.(ChannelOutfitSaved); ok {
		h <- t
	}
}

type channelOutfitDeletedHandler chan<- ChannelOutfitDeleted

func (h channelOutfitDeletedHandler) Type() string {
	return ChannelOutfitDeletedType
}

func (h channelOutfitDeletedHandler) Handle(e any) {
	if t, ok := e.(ChannelOutfitDeleted); ok {
		h <- t
	}
}

type channelCharacterSavedHandler chan<- ChannelCharacterSaved

func (h channelCharacterSavedHandler) Type() string {
	return ChannelCharacterSavedType
}

func (h channelCharacterSavedHandler) Handle(e any) {
	if t, ok := e.(ChannelCharacterSaved); ok {
		h <- t
	}
}

type channelCharacterDeletedHandler chan<- ChannelCharacterDeleted

func (h channelCharacterDeletedHandler) Type() string {
	return ChannelCharacterDeletedType
}

func (h channelCharacterDeletedHandler) Handle(e any) {
	if t, ok := e.(ChannelCharacterDeleted); ok {
		h <- t
	}
}

type outfitMemberSavedHandler chan<- OutfitMemberSaved

func (h outfitMemberSavedHandler) Type() string {
	return OutfitMemberSavedType
}

func (h outfitMemberSavedHandler) Handle(e any) {
	if t, ok := e.(OutfitMemberSaved); ok {
		h <- t
	}
}

type outfitMemberDeletedHandler chan<- OutfitMemberDeleted

func (h outfitMemberDeletedHandler) Type() string {
	return OutfitMemberDeletedType
}

func (h outfitMemberDeletedHandler) Handle(e any) {
	if t, ok := e.(OutfitMemberDeleted); ok {
		h <- t
	}
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
	}
	return nil
}
