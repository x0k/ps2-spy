package outfit_members_saver

import "github.com/x0k/ps2-spy/internal/publisher"

type outfitMembersInitHandler chan<- OutfitMembersInit

func (h outfitMembersInitHandler) Type() string {
	return OutfitMembersInitType
}

func (h outfitMembersInitHandler) Handle(e publisher.Event) {
	h <- e.(OutfitMembersInit)
}

type outfitMembersUpdateHandler chan<- OutfitMembersUpdate

func (h outfitMembersUpdateHandler) Type() string {
	return OutfitMembersUpdateType
}

func (h outfitMembersUpdateHandler) Handle(e publisher.Event) {
	h <- e.(OutfitMembersUpdate)
}

func CastHandler(h any) publisher.Handler {
	switch v := h.(type) {
	case chan OutfitMembersInit:
		return outfitMembersInitHandler(v)
	case chan<- OutfitMembersInit:
		return outfitMembersInitHandler(v)
	case chan OutfitMembersUpdate:
		return outfitMembersUpdateHandler(v)
	case chan<- OutfitMembersUpdate:
		return outfitMembersUpdateHandler(v)
	}
	return nil
}
