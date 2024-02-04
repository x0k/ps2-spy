package worlds_tracker

import "github.com/x0k/ps2-spy/internal/lib/publisher"

type facilityControlHandler chan<- FacilityControl

func (h facilityControlHandler) Type() string {
	return FacilityControlType
}

func (h facilityControlHandler) Handle(e publisher.Event) {
	h <- e.(FacilityControl)
}

type facilityLossHandler chan<- FacilityLoss

func (h facilityLossHandler) Type() string {
	return FacilityLossType
}

func (h facilityLossHandler) Handle(e publisher.Event) {
	h <- e.(FacilityLoss)
}

type Publisher struct {
	publisher.Publisher[publisher.Event]
}

func NewPublisher(pub publisher.Publisher[publisher.Event]) *Publisher {
	return &Publisher{
		Publisher: pub,
	}
}

func (p *Publisher) AddFacilityControlHandler(c chan<- FacilityControl) func() {
	return p.AddHandler(facilityControlHandler(c))
}

func (p *Publisher) AddFacilityLossHandler(c chan<- FacilityLoss) func() {
	return p.AddHandler(facilityLossHandler(c))
}
