package render

import (
	"fmt"

	"github.com/x0k/ps2-spy/internal/ps2"
	factions "github.com/x0k/ps2-spy/internal/ps2/factions"
)

func RenderCharacterLogout(char ps2.Character) string {
	if char.OutfitTag != "" {
		return fmt.Sprintf(
			"[%s] %s (%s) is now offline (%s)",
			char.OutfitTag,
			char.Name,
			factions.FactionNameById(char.FactionId),
			ps2.WorldNameById(char.WorldId),
		)
	}
	return fmt.Sprintf(
		"%s (%s) is now offline (%s)",
		char.Name,
		factions.FactionNameById(char.FactionId),
		ps2.WorldNameById(char.WorldId),
	)
}
