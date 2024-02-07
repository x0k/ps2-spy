package ps2

import (
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/ps2/factions"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type ZoneId string

type WorldId string

type StatPerFactions struct {
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
	StatPerFactions
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
	StatPerFactions
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

type MetagameEventId string

type MetagameEvent struct {
	Id          MetagameEventId
	Name        string
	Description string
	Duration    time.Duration
}

type InstanceId string

const StartedMetagameEventStateName = "started"

type ZoneTerritoryControl struct {
	Id           ZoneId
	IsOpen       bool
	Since        time.Time
	ControlledBy factions.Id
	IsStable     bool
	HasAlerts    bool
	StatPerFactions
}

type WorldTerritoryControl struct {
	Id    WorldId
	Zones []ZoneTerritoryControl
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
	TerritoryControl StatPerFactions
}

type Alerts []Alert

type CharacterId string

type Character struct {
	Id        CharacterId
	FactionId factions.Id
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
var ZoneBenefits = map[ZoneId]string{
	"2":   "Increases heat efficiency of base Phalanx turrets",
	"4":   "Vehicle/Aircraft repair at ammo resupply towers/pads",
	"6":   "Base generators auto-repair over time",
	"8":   "Allied control points increase shield capacity",
	"344": "-20% Air Vehicle Nanite cost",
}

func ZoneBenefitById(id ZoneId) string {
	if benefit, ok := ZoneBenefits[id]; ok {
		return benefit
	}
	return "No benefit"
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
var WorldPlatforms = map[WorldId]platforms.Platform{
	"1":    platforms.PC,
	"10":   platforms.PC,
	"13":   platforms.PC,
	"17":   platforms.PC,
	"19":   platforms.PC,
	"24":   platforms.PC,
	"40":   platforms.PC,
	"1000": platforms.PS4_US,
	"2000": platforms.PS4_EU,
}
var PcPlatformWorldIds = []WorldId{"1", "10", "13", "17", "19", "24", "40"}
var Ps4euPlatformWorldIds = []WorldId{"2000"}
var Ps4usPlatformWorldIds = []WorldId{"1000"}
