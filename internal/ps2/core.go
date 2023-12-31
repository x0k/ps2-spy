package ps2

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type AlertInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

//go:embed data/alerts.json
var alertsFile []byte
var alertsMap = func() map[int]AlertInfo {
	var rawInfo map[string]AlertInfo
	if err := json.Unmarshal(alertsFile, &rawInfo); err != nil {
		panic(err)
	}
	alerts := make(map[int]AlertInfo, len(rawInfo))
	for k, v := range rawInfo {
		id, err := strconv.Atoi(k)
		if err != nil {
			panic(err)
		}
		alerts[id] = v
	}
	return alerts
}()

type ZoneId int

type WorldId int

type StatsByFactions struct {
	All   int
	VS    int
	NC    int
	TR    int
	NS    int
	Other int
}

type ZonePopulation struct {
	StatsByFactions
	Id     ZoneId
	Name   string
	IsOpen bool
}

type Zones map[ZoneId]ZonePopulation

type WorldPopulation struct {
	Id    WorldId
	Name  string
	Total StatsByFactions
	Zones Zones
}

type Worlds map[WorldId]WorldPopulation

type WorldsPopulation struct {
	Total  StatsByFactions
	Worlds Worlds
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
