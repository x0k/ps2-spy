package population_tracker

import (
	"log/slog"

	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/factions"
)

type worldPopulationTracker struct {
	population map[factions.Id]int

	zonesPopulation map[ps2.ZoneId]map[factions.Id]int

	charactersFactions   map[ps2.CharacterId]factions.Id
	charactersLastZoneId map[ps2.CharacterId]ps2.ZoneId
}

func newWorldPopulationTracker() *worldPopulationTracker {
	zonesPopulation := make(map[ps2.ZoneId]map[factions.Id]int, len(ps2.ZoneNames))
	for zoneId := range ps2.ZoneNames {
		zonesPopulation[zoneId] = make(map[factions.Id]int, len(factions.FactionNames))
	}
	return &worldPopulationTracker{
		population: make(map[factions.Id]int, len(factions.FactionNames)),

		zonesPopulation: zonesPopulation,

		charactersFactions:   make(map[ps2.CharacterId]factions.Id),
		charactersLastZoneId: make(map[ps2.CharacterId]ps2.ZoneId),
	}
}

func (w *worldPopulationTracker) HandleLogin(character ps2.Character) {
	w.population[character.FactionId] += 1

	w.charactersFactions[character.Id] = character.FactionId
}

func (w *worldPopulationTracker) HandleLogout(event ps2events.PlayerLogout) bool {
	charId := ps2.CharacterId(event.CharacterID)
	factionId, ok := w.charactersFactions[charId]
	if !ok {
		return false
	}
	delete(w.charactersFactions, charId)

	w.population[factionId] -= 1

	if zoneId, ok := w.charactersLastZoneId[charId]; ok {
		w.zonesPopulation[zoneId][factionId] -= 1
		delete(w.charactersLastZoneId, charId)
	}
	return true
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

func (w *worldPopulationTracker) HandleZoneIdAction(log *slog.Logger, strZoneId, strCharId string) {
	charId := ps2.CharacterId(strCharId)
	factionId, ok := w.charactersFactions[charId]
	if !ok {
		return
	}
	zoneId := ps2.ZoneId(strZoneId)
	if _, ok := w.zonesPopulation[zoneId]; !ok {
		log.Warn("zone not found", slog.String("zone_id", string(zoneId)))
		return
	}
	if lastZoneId, ok := w.charactersLastZoneId[charId]; ok {
		if lastZoneId == zoneId {
			return
		}
		w.zonesPopulation[lastZoneId][factionId] -= 1
	}
	w.zonesPopulation[zoneId][factionId] += 1
	w.charactersLastZoneId[charId] = zoneId
}

func (w *worldPopulationTracker) Population() map[factions.Id]int {
	return w.population
}

func (w *worldPopulationTracker) ZonesPopulation() map[ps2.ZoneId]map[factions.Id]int {
	return w.zonesPopulation
}
