package stats_tracker

import (
	"time"

	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_loadout "github.com/x0k/ps2-spy/internal/ps2/loadout"
)

type characterTracker struct {
	id ps2.CharacterId
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

func newCharacterTracker(id ps2.CharacterId) *characterTracker {
	return &characterTracker{
		id: id,
	}
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

func (c *characterTracker) toStats(char ps2.Character, stoppedAt time.Time) CharacterStats {
	if !c.lastLoadoutUpdate.IsZero() {
		c.LoadoutsDistribution[c.lastLoadoutType] += stoppedAt.Sub(c.lastLoadoutUpdate)
	}
	return CharacterStats{
		Character:              char,
		BodyKills:              c.bodyKills,
		HeadShotsKills:         c.headShotsKills,
		TeamKills:              c.teamKills,
		Deaths:                 c.deaths,
		DeathsByRestrictedArea: c.deathsByRestrictedArea,
		Suicides:               c.suicides,
		LoadoutsDistribution:   c.LoadoutsDistribution,
	}
}
