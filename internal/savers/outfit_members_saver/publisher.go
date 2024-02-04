package outfit_members_saver

import "github.com/x0k/ps2-spy/internal/lib/publisher"

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

type Publisher struct {
	publisher.Publisher[publisher.Event]
}

func NewPublisher(pub publisher.Publisher[publisher.Event]) *Publisher {
	return &Publisher{
		Publisher: pub,
	}
}

func (p *Publisher) AddOutfitMembersUpdateHandler(c chan<- OutfitMembersUpdate) func() {
	return p.AddHandler(outfitMembersUpdateHandler(c))
}
