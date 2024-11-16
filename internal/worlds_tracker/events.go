package worlds_tracker

import (
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	FacilityControlType EventType = "facility_control"
	FacilityLossType    EventType = "facility_loss"
)

type FacilityControl struct {
	ps2events.FacilityControl
	OldOutfitId ps2.OutfitId
}

func (e FacilityControl) Type() EventType {
	return FacilityControlType
}

type FacilityLoss struct {
	ps2events.FacilityControl
	OldOutfitId ps2.OutfitId
}

func (e FacilityLoss) Type() EventType {
	return FacilityLossType
}
