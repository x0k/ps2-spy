package render

import (
	"fmt"

	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/factions"
)

func RenderCharacterLogin(char ps2.Character) string {
	if char.OutfitId != "" {
		return fmt.Sprintf(
			"[%s] %s (%s) is now on %s",
			char.OutfitId,
			char.Name,
			factions.FactionNameById(char.FactionId),
			ps2.WorldNameById(char.WorldId),
		)
	}
	return fmt.Sprintf(
		"%s (%s) is now on %s",
		char.Name,
		factions.FactionNameById(char.FactionId),
		ps2.WorldNameById(char.WorldId),
	)
}
