package ps2

import "fmt"

type ZoneId int

type WorldId int

type CommonPopulation struct {
	All   int
	VS    int
	NC    int
	TR    int
	Other int
}

type ZonePopulation struct {
	CommonPopulation
	Id     ZoneId
	Name   string
	IsOpen bool
}

type Zones map[ZoneId]ZonePopulation

type WorldPopulation struct {
	Id    WorldId
	Name  string
	Total CommonPopulation
	Zones Zones
}

type Worlds map[WorldId]WorldPopulation

type Population struct {
	Total  CommonPopulation
	Worlds Worlds
}

var ErrWorldNotFound = fmt.Errorf("world not found")
var ZoneNames = map[ZoneId]string{
	2:   "Indar",
	4:   "Hossin",
	6:   "Amerish",
	8:   "Esamir",
	344: "Oshur",
	14:  "Koltyr",
}
var WorldNames = map[WorldId]string{
	1:    "Connery",
	10:   "Miller",
	13:   "Cobalt",
	17:   "Emerald",
	19:   "Jaeger",
	24:   "Apex",
	40:   "SolTech",
	1000: "Genudine",
	2000: "Ceres",
}
