package characters_tracker

import (
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_factions "github.com/x0k/ps2-spy/internal/ps2/factions"
)

type worldPopulationTracker struct {
	population map[ps2_factions.Id]int

	zonesPopulation map[ps2.ZoneId]map[ps2_factions.Id]int

	charactersFactions   map[ps2.CharacterId]ps2_factions.Id
	charactersLastZoneId map[ps2.CharacterId]ps2.ZoneId
}

func newWorldPopulationTracker() worldPopulationTracker {
	zonesPopulation := make(map[ps2.ZoneId]map[ps2_factions.Id]int, len(ps2.ZoneNames))
	for zoneId := range ps2.ZoneNames {
		zonesPopulation[zoneId] = make(map[ps2_factions.Id]int, len(ps2_factions.FactionNames))
	}
	return worldPopulationTracker{
		population: make(map[ps2_factions.Id]int, len(ps2_factions.FactionNames)),

		zonesPopulation: zonesPopulation,

		charactersFactions:   make(map[ps2.CharacterId]ps2_factions.Id),
		charactersLastZoneId: make(map[ps2.CharacterId]ps2.ZoneId),
	}
}

func (w *worldPopulationTracker) HandleLogin(character ps2.Character) {
	if _, ok := w.charactersFactions[character.Id]; ok {
		return
	}
	w.population[character.FactionId] += 1
	w.charactersFactions[character.Id] = character.FactionId
}

func (w *worldPopulationTracker) HandleInactive(charId ps2.CharacterId) {
	factionId, ok := w.charactersFactions[charId]
	if !ok {
		return
	}
	delete(w.charactersFactions, charId)
	w.population[factionId] -= 1
	if zoneId, ok := w.charactersLastZoneId[charId]; ok {
		w.zonesPopulation[zoneId][factionId] -= 1
		delete(w.charactersLastZoneId, charId)
	}
}

func (w *worldPopulationTracker) HandleZoneAction(charId ps2.CharacterId, strZoneId string) {
	factionId, ok := w.charactersFactions[charId]
	if !ok {
		return
	}
	zoneId := ps2.ZoneId(strZoneId)
	if lastZoneId, ok := w.charactersLastZoneId[charId]; ok {
		if lastZoneId == zoneId {
			return
		}
		w.zonesPopulation[lastZoneId][factionId] -= 1
	}
	// Non interesting zone like VR training
	if _, ok := w.zonesPopulation[zoneId]; !ok {
		delete(w.charactersLastZoneId, charId)
		return
	}
	w.zonesPopulation[zoneId][factionId] += 1
	w.charactersLastZoneId[charId] = zoneId
}

func (w *worldPopulationTracker) Population() map[ps2_factions.Id]int {
	return w.population
}

func (w *worldPopulationTracker) ZonesPopulation() map[ps2.ZoneId]map[ps2_factions.Id]int {
	return w.zonesPopulation
}
