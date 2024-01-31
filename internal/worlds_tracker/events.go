package worlds_tracker

import (
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/ps2"
)

const (
	FacilityControlType = "facility_control"
	FacilityLossType    = "facility_loss"
)

type FacilityControl struct {
	ps2events.FacilityControl
	OldOutfitId ps2.OutfitId
}

func (e FacilityControl) Type() string {
	return FacilityControlType
}

type FacilityLoss struct {
	ps2events.FacilityControl
	OldOutfitId ps2.OutfitId
}

func (e FacilityLoss) Type() string {
	return FacilityLossType
}
