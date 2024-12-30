package stats_tracker

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_factions "github.com/x0k/ps2-spy/internal/ps2/factions"
	ps2_loadout "github.com/x0k/ps2-spy/internal/ps2/loadout"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type CharactersLoader = func(
	context.Context, ps2_platforms.Platform, []ps2.CharacterId,
) (map[ps2.CharacterId]ps2.Character, error)

type platformTracker struct {
	platform         ps2_platforms.Platform
	mu               sync.Mutex
	characters       map[ps2.CharacterId]*characterTracker
	charactersLoader CharactersLoader
}

func newPlatformTracker(
	platform ps2_platforms.Platform,
	charactersLoader CharactersLoader,
) *platformTracker {
	return &platformTracker{
		platform:         platform,
		characters:       make(map[ps2.CharacterId]*characterTracker),
		charactersLoader: charactersLoader,
	}
}

func (c *platformTracker) characterTracker(characterId ps2.CharacterId) *characterTracker {
	character, ok := c.characters[characterId]
	if !ok {
		character = newCharacterTracker(characterId)
		c.characters[characterId] = character
	}
	return character
}

func (c *platformTracker) handleCharacterEvent(
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

func (c *platformTracker) toStats(
	ctx context.Context,
	stoppedAt time.Time,
) (PlatformStats, error) {
	characters, err := c.charactersLoader(
		ctx, c.platform, slices.Collect(maps.Keys(c.characters)),
	)
	if len(characters) == 0 && err != nil {
		return PlatformStats{}, fmt.Errorf("failed to get any character: %w", err)
	}
	stats := make([]CharacterStats, 0, len(c.characters))
	for characterId, tracker := range c.characters {
		char, ok := characters[characterId]
		if !ok {
			char = ps2.Character{
				Id:        characterId,
				FactionId: ps2_factions.None,
				Name:      string(characterId),
				OutfitId:  ps2.OutfitId(""),
				OutfitTag: "Unknown",
				WorldId:   ps2.WorldId(""),
				Platform:  c.platform,
			}
		}
		stats = append(stats, tracker.toStats(char, stoppedAt))
	}
	slices.SortFunc(stats, func(a, b CharacterStats) int {
		return int(b.BodyKills+b.HeadShotsKills) - int(a.BodyKills+a.HeadShotsKills)
	})
	return PlatformStats{
		Platform:   c.platform,
		Characters: stats,
	}, nil
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
