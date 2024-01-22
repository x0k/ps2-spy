package facilities_manager

import "github.com/x0k/ps2-spy/internal/publisher"

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

func CastHandler(handler any) publisher.Handler {
	switch v := handler.(type) {
	case chan FacilityControl:
		return facilityControlHandler(v)
	case chan<- FacilityControl:
		return facilityControlHandler(v)
	case chan FacilityLoss:
		return facilityLossHandler(v)
	case chan<- FacilityLoss:
		return facilityLossHandler(v)
	}
	return nil
}
