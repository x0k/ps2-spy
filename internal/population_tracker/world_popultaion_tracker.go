package population_tracker

import (
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type worldPopulationTracker struct {
	population           int
	zonesPopulation      map[ps2.ZoneId]int
	charactersLastZoneId map[ps2.CharacterId]ps2.ZoneId
}

func newWorldPopulationTracker() *worldPopulationTracker {
	return &worldPopulationTracker{
		zonesPopulation:      make(map[ps2.ZoneId]int, len(ps2.ZoneNames)),
		charactersLastZoneId: make(map[ps2.CharacterId]ps2.ZoneId),
	}
}

func (w *worldPopulationTracker) HandleLogin(event ps2events.PlayerLogin) {
	w.population += 1
}

func (w *worldPopulationTracker) HandleLogout(event ps2events.PlayerLogout) {
	w.population -= 1
	cId := ps2.CharacterId(event.CharacterID)
	if zId, ok := w.charactersLastZoneId[cId]; ok {
		w.zonesPopulation[zId] -= 1
		delete(w.charactersLastZoneId, cId)
	}
}

func (w *worldPopulationTracker) HandleZoneIdAction(strZoneId, strCharId string) {
	zoneId := ps2.ZoneId(strZoneId)
	charId := ps2.CharacterId(strCharId)
	if lastZoneId, ok := w.charactersLastZoneId[charId]; ok {
		if lastZoneId == zoneId {
			return
		}
		w.zonesPopulation[lastZoneId] -= 1
	}
	w.zonesPopulation[zoneId] += 1
	w.charactersLastZoneId[charId] = zoneId
}
