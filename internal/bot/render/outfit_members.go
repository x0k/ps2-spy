package render

import (
	"strings"

	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func RenderOutfitMembersUpdate(outfit ps2.Outfit, change diff.Diff[ps2.Character]) string {
	builder := strings.Builder{}
	builder.WriteString("Update of ")
	builder.WriteString(outfit.Name)
	builder.WriteString(" [")
	builder.WriteString(outfit.Tag)
	builder.WriteString("] outfit members:")
	if len(change.ToAdd) > 0 {
		builder.WriteString("\n**Welcome to the outfit:**")
		for i := range change.ToAdd {
			builder.WriteString("\n- ")
			builder.WriteString(change.ToAdd[i].Name)
		}
	}
	if len(change.ToDel) > 0 {
		builder.WriteString("\n**Left the outfit:**")
		for i := range change.ToDel {
			builder.WriteString("\n- ")
			builder.WriteString(change.ToDel[i].Name)
		}
	}
	return builder.String()
}
