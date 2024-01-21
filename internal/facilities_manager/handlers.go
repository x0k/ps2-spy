package facilities_manager

import "github.com/x0k/ps2-spy/internal/publisher"

type facilityLossHandler chan<- FacilityLoss

func (h facilityLossHandler) Type() string {
	return FacilityLossType
}

func (h facilityLossHandler) Handle(e publisher.Event) {
	h <- e.(FacilityLoss)
}

func CastHandler(handler any) publisher.Handler {
	switch v := handler.(type) {
	case chan FacilityLoss:
		return facilityLossHandler(v)
	case chan<- FacilityLoss:
		return facilityLossHandler(v)
	}
	return nil
}
