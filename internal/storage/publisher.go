package storage

import "github.com/x0k/ps2-spy/internal/lib/publisher"

type channelOutfitSavedHandler chan<- ChannelOutfitSaved

func (h channelOutfitSavedHandler) Type() string {
	return ChannelOutfitSavedType
}

func (h channelOutfitSavedHandler) Handle(e publisher.Event) {
	h <- e.(ChannelOutfitSaved)
}

type channelOutfitDeletedHandler chan<- ChannelOutfitDeleted

func (h channelOutfitDeletedHandler) Type() string {
	return ChannelOutfitDeletedType
}

func (h channelOutfitDeletedHandler) Handle(e publisher.Event) {
	h <- e.(ChannelOutfitDeleted)
}

type channelCharacterSavedHandler chan<- ChannelCharacterSaved

func (h channelCharacterSavedHandler) Type() string {
	return ChannelCharacterSavedType
}

func (h channelCharacterSavedHandler) Handle(e publisher.Event) {
	h <- e.(ChannelCharacterSaved)
}

type channelCharacterDeletedHandler chan<- ChannelCharacterDeleted

func (h channelCharacterDeletedHandler) Type() string {
	return ChannelCharacterDeletedType
}

func (h channelCharacterDeletedHandler) Handle(e publisher.Event) {
	h <- e.(ChannelCharacterDeleted)
}

type outfitMemberSavedHandler chan<- OutfitMemberSaved

func (h outfitMemberSavedHandler) Type() string {
	return OutfitMemberSavedType
}

func (h outfitMemberSavedHandler) Handle(e publisher.Event) {
	h <- e.(OutfitMemberSaved)
}

type outfitMemberDeletedHandler chan<- OutfitMemberDeleted

func (h outfitMemberDeletedHandler) Type() string {
	return OutfitMemberDeletedType
}

func (h outfitMemberDeletedHandler) Handle(e publisher.Event) {
	h <- e.(OutfitMemberDeleted)
}

type outfitSynchronizedHandler chan<- OutfitSynchronized

func (h outfitSynchronizedHandler) Type() string {
	return OutfitSynchronizedType
}

func (h outfitSynchronizedHandler) Handle(e publisher.Event) {
	h <- e.(OutfitSynchronized)
}

type Publisher struct {
	publisher.Publisher[publisher.Event]
}

func NewPublisher(pub publisher.Publisher[publisher.Event]) *Publisher {
	return &Publisher{pub}
}

func (p *Publisher) AddChannelOutfitSavedHandler(c chan<- ChannelOutfitSaved) func() {
	return p.AddHandler(channelOutfitSavedHandler(c))
}

func (p *Publisher) AddChannelOutfitDeletedHandler(c chan<- ChannelOutfitDeleted) func() {
	return p.AddHandler(channelOutfitDeletedHandler(c))
}

func (p *Publisher) AddChannelCharacterSavedHandler(c chan<- ChannelCharacterSaved) func() {
	return p.AddHandler(channelCharacterSavedHandler(c))
}

func (p *Publisher) AddChannelCharacterDeletedHandler(c chan<- ChannelCharacterDeleted) func() {
	return p.AddHandler(channelCharacterDeletedHandler(c))
}

func (p *Publisher) AddOutfitMemberSavedHandler(c chan<- OutfitMemberSaved) func() {
	return p.AddHandler(outfitMemberSavedHandler(c))
}

func (p *Publisher) AddOutfitMemberDeletedHandler(c chan<- OutfitMemberDeleted) func() {
	return p.AddHandler(outfitMemberDeletedHandler(c))
}

func (p *Publisher) AddOutfitSynchronizedHandler(c chan<- OutfitSynchronized) func() {
	return p.AddHandler(outfitSynchronizedHandler(c))
}
