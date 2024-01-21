package facilities_manager

import ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"

const (
	FacilityLossType = "facility_loss"
)

type FacilityLoss struct {
	ps2events.FacilityControl
	OldOutfitId string
}

func (e FacilityLoss) Type() string {
	return FacilityLossType
}
