package stats_tracker

import (
	"sync"
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

func (c *characterTracker) toStats() {
	if !c.lastLoadoutUpdate.IsZero() {
		c.LoadoutsDistribution[c.lastLoadoutType] += time.Since(c.lastLoadoutUpdate)
	}
}

type ChannelTracker struct {
	mu         sync.Mutex
	characters map[ps2.CharacterId]*characterTracker
}

func newChannelTracker() *ChannelTracker {
	return &ChannelTracker{
		characters: make(map[ps2.CharacterId]*characterTracker),
	}
}

func (c *ChannelTracker) characterTracker(characterId ps2.CharacterId) *characterTracker {
	character, ok := c.characters[characterId]
	if !ok {
		character = &characterTracker{}
		c.characters[characterId] = character
	}
	return character
}

func (c *ChannelTracker) handleCharacterEvent(
	characterId ps2.CharacterId,
	loadout ps2_loadout.LoadoutType,
	update func(*characterTracker),
) {
	c.mu.Lock()
	defer c.mu.Unlock()
	character := c.characterTracker(characterId)
	update(character)
	character.updateLoadout(loadout)
}

func addDeath(character *characterTracker) {
	character.deaths++
}

func addDeathByRestrictedArea(character *characterTracker) {
	character.deathsByRestrictedArea++
}

func addSuicide(character *characterTracker) {
	character.suicides++
}

func addBodyKill(character *characterTracker) {
	character.bodyKills++
}

func addHeadShotKill(character *characterTracker) {
	character.headShotsKills++
}

func addTeamKill(character *characterTracker) {
	character.teamKills++
}

func updateLoadout(character *characterTracker) {}
