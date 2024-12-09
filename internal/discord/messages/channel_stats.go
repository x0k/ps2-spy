package discord_messages

import (
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	ps2_factions "github.com/x0k/ps2-spy/internal/ps2/factions"
	ps2_loadout "github.com/x0k/ps2-spy/internal/ps2/loadout"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	"golang.org/x/text/message"
)

func renderLoadoutType(p *message.Printer, lt ps2_loadout.LoadoutType) string {
	switch lt {
	case ps2_loadout.Infiltrator:
		return p.Sprintf("INF")
	case ps2_loadout.LightAssault:
		return p.Sprintf("LA")
	case ps2_loadout.Medic:
		return p.Sprintf("MED")
	case ps2_loadout.Engineer:
		return p.Sprintf("ENG")
	case ps2_loadout.HeavyAssault:
		return p.Sprintf("HA")
	case ps2_loadout.MAX:
		return p.Sprintf("MAX")
	}
	return p.Sprintf("Unknown")
}

func renderPlatformStats(p *message.Printer, sb *strings.Builder, stats stats_tracker.PlatformStats) {
	t := tablewriter.NewWriter(sb)
	t.SetHeader([]string{
		p.Sprintf("Faction"),
		p.Sprintf("Outfit"),
		p.Sprintf("Character"),
		p.Sprintf("Kills"),
		p.Sprintf("HS%%"),
		p.Sprintf("KD"),
		p.Sprintf("Loadout"),
	})
	t.SetBorder(false)
	for _, char := range stats.Characters {
		allKills := char.BodyKills + char.HeadShotsKills
		headShotsRatio := float64(0)
		if allKills > 0 {
			headShotsRatio = float64(char.HeadShotsKills) / float64(allKills)
		}
		allDeaths := char.Deaths + char.Suicides
		killDeathRatio := float64(allKills)
		if allDeaths > 0 {
			killDeathRatio = float64(allKills) / float64(allDeaths)
		}
		mainLoadoutType := ps2_loadout.HeavyAssault
		maxDuration := time.Duration(0)
		for i, d := range char.LoadoutsDistribution {
			if d > maxDuration {
				maxDuration = d
				mainLoadoutType = ps2_loadout.LoadoutType(i)
			}
		}
		t.Append([]string{
			ps2_factions.FactionNameById(char.Character.FactionId),
			char.Character.OutfitTag,
			char.Character.Name,
			strconv.FormatUint(uint64(allKills), 10),
			strconv.FormatFloat(headShotsRatio, 'f', 2, 64),
			strconv.FormatFloat(killDeathRatio, 'f', 2, 64),
			renderLoadoutType(p, mainLoadoutType),
		})
	}
	t.Render()
}
