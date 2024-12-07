package stats_tracker

import (
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_loadout "github.com/x0k/ps2-spy/internal/ps2/loadout"
)

type CharacterStats struct {
	// Kills
	HeadShotsKills uint
	BodyKills      uint
	TeamKills      uint
	// Deaths
	Deaths uint
	// With deaths by restricted area
	Suicides uint

	LoadoutsDistribution [ps2_loadout.LoadoutTypeCount]time.Duration
	lastLoadoutType      ps2_loadout.LoadoutType
	lastLoadoutUpdate    time.Time
}

func (c *CharacterStats) updateLoadout(loadout ps2_loadout.LoadoutType) {
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

type ChannelTracker struct {
	startedAt  time.Time
	mu         sync.Mutex
	characters map[ps2.CharacterId]*CharacterStats
}

func newChannelTracker() *ChannelTracker {
	return &ChannelTracker{
		startedAt:  time.Now(),
		characters: make(map[ps2.CharacterId]*CharacterStats),
	}
}

func (c *ChannelTracker) getCharacterStats(characterId ps2.CharacterId) *CharacterStats {
	character, ok := c.characters[characterId]
	if !ok {
		character = &CharacterStats{}
		c.characters[characterId] = character
	}
	return character
}

func (c *ChannelTracker) handleCharacterEvent(
	characterId ps2.CharacterId,
	loadout ps2_loadout.LoadoutType,
	update func(*CharacterStats),
) {
	c.mu.Lock()
	defer c.mu.Unlock()
	character := c.getCharacterStats(characterId)
	update(character)
	character.updateLoadout(loadout)
}

func addDeath(character *CharacterStats) {
	character.Deaths++
}

func addSuicide(character *CharacterStats) {
	character.Suicides++
}

func addBodyKill(character *CharacterStats) {
	character.BodyKills++
}

func addHeadShotKill(character *CharacterStats) {
	character.HeadShotsKills++
}

func addTeamKill(character *CharacterStats) {
	character.TeamKills++
}

func updateLoadout(character *CharacterStats) {}
