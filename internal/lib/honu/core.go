package honu

type ZoneAlert struct {
	Id                int    `json:"id"`
	Timestamp         string `json:"timestamp"`
	Duration          int    `json:"duration"`
	ZoneId            int    `json:"zoneID"`
	WorldId           int    `json:"worldID"`
	AlertId           int    `json:"alertID"`
	InstanceId        int    `json:"instanceID"`
	Name              string `json:"name"`
	VictorFactionId   int    `json:"victorFactionID"`
	WarpgateVS        int    `json:"warpgateVS"`
	WarpgateNC        int    `json:"warpgateNC"`
	WarpgateTR        int    `json:"warpgateTR"`
	ZoneFacilityCount int    `json:"zoneFacilityCount"`
	CountVS           int    `json:"countVS"`
	CountNC           int    `json:"countNC"`
	CountTR           int    `json:"countTR"`
	Participants      int    `json:"participants"`
}

type ZoneAlertInfo struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	TypeId          int    `json:"typeID"`
	DurationMinutes int    `json:"durationMinutes"`
}

type ZonePlayers struct {
	All     int `json:"all"`
	VS      int `json:"vs"`
	NC      int `json:"nc"`
	TR      int `json:"tr"`
	Unknown int `json:"unknown"`
}

type ZoneTerritoryControl struct {
	VS    int `json:"vs"`
	NC    int `json:"nc"`
	TR    int `json:"tr"`
	Total int `json:"total"`
}

type WorldZone struct {
	ZoneId           int                  `json:"zoneID"`
	WorldId          int                  `json:"worldID"`
	IsOpened         bool                 `json:"isOpened"`
	UnstableState    int                  `json:"unstableState"`
	Alert            ZoneAlert            `json:"alert"`
	AlertInfo        ZoneAlertInfo        `json:"alertInfo"`
	AlertStart       string               `json:"alertStart"`
	AlertEnd         string               `json:"alertEnd"`
	LastLocked       string               `json:"lastLocked"`
	PlayerCount      int                  `json:"playerCount"`
	Players          ZonePlayers          `json:"players"`
	TerritoryControl ZoneTerritoryControl `json:"territoryControl"`
}

type World struct {
	WorldId       int         `json:"worldID"`
	WorldName     string      `json:"worldName"`
	PlayersOnline int         `json:"playersOnline"`
	Zones         []WorldZone `json:"zones"`
}
