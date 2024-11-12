package worlds_tracker

type facilityControlHandler chan<- FacilityControl

func (h facilityControlHandler) Type() EventType {
	return FacilityControlType
}

func (h facilityControlHandler) Handle(e Event) {
	h <- e.(FacilityControl)
}

type facilityLossHandler chan<- FacilityLoss

func (h facilityLossHandler) Type() EventType {
	return FacilityLossType
}

func (h facilityLossHandler) Handle(e Event) {
	h <- e.(FacilityLoss)
}
