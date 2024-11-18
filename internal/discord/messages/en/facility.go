package render

import (
	"fmt"

	"github.com/x0k/ps2-spy/internal/ps2"
)

func RenderFacilityControl(
	worldId ps2.WorldId,
	outfit ps2.Outfit,
	facility ps2.Facility,
) string {
	return fmt.Sprintf(
		"%s [%s] captured %s (%s) on %s (%s)",
		outfit.Name,
		outfit.Tag,
		facility.Name,
		facility.Type,
		ps2.ZoneNameById(facility.ZoneId),
		ps2.WorldNameById(worldId),
	)
}

func RenderFacilityLoss(
	worldId ps2.WorldId,
	outfit ps2.Outfit,
	facility ps2.Facility,
) string {
	return fmt.Sprintf(
		"%s [%s] lost %s (%s) on %s (%s)",
		outfit.Name,
		outfit.Tag,
		facility.Name,
		facility.Type,
		ps2.ZoneNameById(facility.ZoneId),
		ps2.WorldNameById(worldId),
	)
}
