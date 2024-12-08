package discord_messages

import (
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	"golang.org/x/text/message"
)

func renderPlatformStats(p *message.Printer, sb *strings.Builder, stats map[ps2.CharacterId]*stats_tracker.CharacterStats) string {
	t := tablewriter.NewWriter(sb)
	t.SetHeader([]string{
		p.Sprintf("Faction"),
		p.Sprintf("Outfit"),
		p.Sprintf("Character"),
		p.Sprintf("Kills"),
		p.Sprintf("HS%"),
		p.Sprintf("KD"),
		p.Sprintf("Loadout"),
	})
	for id, char := range stats {
		allKills := char.BodyKills + char.HeadShotsKills
		allDeaths := char.Deaths + char.Suicides
		t.Append([]string{
			"Unknown",
			"Unknown",
			string(id),
			strconv.FormatUint(uint64(allKills), 10),
			strconv.FormatFloat(float64(char.HeadShotsKills)/float64(allKills), 'f', 2, 64),
			strconv.FormatFloat(float64(allKills)/float64(allDeaths), 'f', 2, 64),
		})

	}
	t.Render()
	return sb.String()
}

func renderChannelStats(p *message.Printer, event stats_tracker.ChannelTrackerStopped) string {
	sb := strings.Builder{}

}
