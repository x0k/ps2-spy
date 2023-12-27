package ps2

import "fmt"

type ZoneId int

type WorldId int

type ZonePopulation struct {
	ZoneId ZoneId
	IsOpen bool
	All    int
	VS     int
	NC     int
	TR     int
	Other  int
}

type Zones map[ZoneId]ZonePopulation

type WorldPopulation struct {
	WorldId WorldId
	Total   ZonePopulation
	Zones   Zones
}

type Population map[WorldId]WorldPopulation

var ErrWorldNotFound = fmt.Errorf("world not found")
