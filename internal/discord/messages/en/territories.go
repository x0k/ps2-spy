package en_messages

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	messages_shared "github.com/x0k/ps2-spy/internal/discord/messages/shared"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	factions "github.com/x0k/ps2-spy/internal/ps2/factions"
)

func renderZoneTerritoryControl(zone ps2.ZoneTerritoryControl) string {
	b := strings.Builder{}
	if zone.IsOpen {
		b.WriteString("Unlocked ")
	} else {
		b.WriteString("Locked ")
	}
	b.WriteString(messages_shared.RenderTime(zone.Since))
	b.WriteString(" (")
	b.WriteString(messages_shared.RenderRelativeTime(zone.Since))
	if zone.ControlledBy == factions.None {
		b.WriteByte(')')
	} else {
		b.WriteString(") by `")
		b.WriteString(factions.FactionNameById(zone.ControlledBy))
		b.WriteByte('`')
	}
	if !zone.IsOpen {
		return b.String()
	}
	b.WriteString("\nStatus: _")
	if zone.IsStable {
		b.WriteString("Stable")
	} else {
		b.WriteString("Unstable")
	}
	b.WriteString("_\nAlerts: _")
	if zone.HasAlerts {
		b.WriteString("Yes")
	} else {
		b.WriteString("No")
	}
	b.WriteString("_\n")
	RenderStatPerFactions(&b, zone.StatPerFactions)
	return b.String()
}

func RenderZoneTerritoryControl(zone ps2.ZoneTerritoryControl) *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Name:  ps2.ZoneNameById(zone.Id),
		Value: renderZoneTerritoryControl(zone),
	}
}

func RenderZoneTerritoryControls(zones []ps2.ZoneTerritoryControl) []*discordgo.MessageEmbedField {
	fields := make([]*discordgo.MessageEmbedField, 0, len(zones))
	for _, zone := range zones {
		fields = append(fields, RenderZoneTerritoryControl(zone))
	}
	return fields
}

func RenderWorldTerritoriesControl(loaded meta.Loaded[ps2.WorldTerritoryControl]) *discordgo.MessageEmbed {
	world := loaded.Value
	return &discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Title:  ps2.WorldNameById(world.Id),
		Fields: RenderZoneTerritoryControls(world.Zones),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Source: %s", loaded.Source),
		},
		Timestamp: loaded.UpdatedAt.Format(time.RFC3339),
	}
}