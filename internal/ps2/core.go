package ps2

import (
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type ZoneId string

type WorldId string

type StatsByFactions struct {
	All   int
	VS    int
	NC    int
	TR    int
	NS    int
	Other int
}

type ZonePopulation struct {
	Id     ZoneId
	Name   string
	IsOpen bool
	StatsByFactions
}

type DetailedWorldPopulation struct {
	Id    WorldId
	Name  string
	Total int
	Zones []ZonePopulation
}

func WorldNameById(id WorldId) string {
	if name, ok := WorldNames[id]; ok {
		return name
	}
	return fmt.Sprintf("World %s", id)
}

func ZoneNameById(id ZoneId) string {
	if name, ok := ZoneNames[id]; ok {
		return name
	}
	return fmt.Sprintf("Zone %s", id)
}

type WorldPopulation struct {
	Id   WorldId
	Name string
	StatsByFactions
}

func NewWorldPopulation(id WorldId, name string) WorldPopulation {
	if name == "" {
		return WorldPopulation{
			Id:   id,
			Name: WorldNameById(id),
		}
	}
	return WorldPopulation{
		Id:   id,
		Name: name,
	}
}

type WorldsPopulation struct {
	Total  int
	Worlds []WorldPopulation
}

type Alert struct {
	WorldId          WorldId
	WorldName        string
	ZoneId           ZoneId
	ZoneName         string
	AlertName        string
	AlertDescription string
	StartedAt        time.Time
	Duration         time.Duration
	TerritoryControl StatsByFactions
}

type Alerts []Alert

type CharacterId string

type Character struct {
	Id        CharacterId
	FactionId string
	Name      string
	OutfitId  OutfitId
	OutfitTag string
	WorldId   WorldId
	Platform  platforms.Platform
}

type OutfitId string

type Outfit struct {
	Id       OutfitId
	Name     string
	Tag      string
	Platform platforms.Platform
}

type FacilityId string

type Facility struct {
	Id     FacilityId
	Name   string
	Type   string
	ZoneId ZoneId
}

var ErrWorldNotFound = fmt.Errorf("world not found")
var ZoneNames = map[ZoneId]string{
	"2":   "Indar",
	"4":   "Hossin",
	"6":   "Amerish",
	"8":   "Esamir",
	"344": "Oshur",
}
var ZoneFacilitiesCount = map[ZoneId]int{
	"2":   89,
	"4":   88,
	"6":   81,
	"8":   51,
	"344": 75,
}
var WorldNames = map[WorldId]string{
	"1":    "Connery",
	"10":   "Miller",
	"13":   "Cobalt",
	"17":   "Emerald",
	"19":   "Jaeger",
	"24":   "Apex",
	"40":   "SolTech",
	"1000": "Genudine",
	"2000": "Ceres",
}
