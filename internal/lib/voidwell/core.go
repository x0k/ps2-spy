package voidwell

type Population struct {
	VS int `json:"vs"`
	NC int `json:"nc"`
	TR int `json:"tr"`
	NS int `json:"ns"`
}

type LockState struct {
	// LOCKED, UNLOCKED
	State             string `json:"state"`
	Timestamp         string `json:"timestamp"`
	MetagameEventID   int    `json:"metagameEventId"`
	TriggeringFaction int    `json:"triggeringFaction"`
}

type MetagameEvent struct {
	Id          int    `json:"id"`
	TypeId      int    `json:"typeId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ZoneId      int    `json:"zoneId"`
	// Example "01:30:00"
	Duration string `json:"duration"`
}

type AlertState struct {
	Timestamp     string        `json:"timestamp"`
	InstanceId    int           `json:"instanceId"`
	MetagameEvent MetagameEvent `json:"metagameEvent"`
}

type ZoneState struct {
	Id         int        `json:"id"`
	Name       string     `json:"name"`
	IsTracking bool       `json:"isTracking"`
	LockState  LockState  `json:"lockState"`
	AlertState AlertState `json:"alertState"`
	Population Population `json:"population"`
}

type World struct {
	Id               int         `json:"id"`
	Name             string      `json:"name"`
	IsOnline         bool        `json:"isOnline"`
	OnlineCharacters int         `json:"onlineCharacters"`
	ZoneStates       []ZoneState `json:"zoneStates"`
}
