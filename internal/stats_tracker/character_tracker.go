package stats_tracker

import (
	"time"

	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_loadout "github.com/x0k/ps2-spy/internal/ps2/loadout"
)

type characterTracker struct {
	// Kills
	bodyKills      uint
	headShotsKills uint
	teamKills      uint
	// deaths
	deaths                 uint
	deathsByRestrictedArea uint
	suicides               uint

	LoadoutsDistribution [ps2_loadout.LoadoutTypeCount]time.Duration
	lastLoadoutType      ps2_loadout.LoadoutType
	lastLoadoutUpdate    time.Time
}

func (c *characterTracker) updateLoadout(loadout ps2_loadout.LoadoutType) {
	now := time.Now()
	if c.lastLoadoutUpdate.IsZero() {
		c.lastLoadoutUpdate = now
		c.lastLoadoutType = loadout
		return
	}
	if c.lastLoadoutType == loadout {
		return
	}
	c.LoadoutsDistribution[c.lastLoadoutType] += now.Sub(c.lastLoadoutUpdate)
	c.lastLoadoutUpdate = now
	c.lastLoadoutType = loadout
}

func (c *characterTracker) toStats(stoppedAt time.Time, character ps2.Character) CharacterStats {
	if !c.lastLoadoutUpdate.IsZero() {
		c.LoadoutsDistribution[c.lastLoadoutType] += stoppedAt.Sub(c.lastLoadoutUpdate)
	}
	return CharacterStats{
		Character:              character,
		BodyKills:              c.bodyKills,
		HeadShotsKills:         c.headShotsKills,
		TeamKills:              c.teamKills,
		Deaths:                 c.deaths,
		DeathsByRestrictedArea: c.deathsByRestrictedArea,
		Suicides:               c.suicides,
		LoadoutsDistribution:   c.LoadoutsDistribution,
	}
}
