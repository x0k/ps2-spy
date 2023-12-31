package ps2live

import "golang.org/x/exp/constraints"

type Number interface {
	constraints.Integer | constraints.Float
}

type Factions[T Number] struct {
	NC T `json:"nc"`
	VS T `json:"vs"`
	TR T `json:"tr"`
}

type FactionsWithTotal struct {
	NC    int `json:"nc"`
	VS    int `json:"vs"`
	TR    int `json:"tr"`
	Total int `json:"total"`
}

type PopulationServices struct {
	Honu      FactionsWithTotal `json:"honu"`
	Fisu      FactionsWithTotal `json:"fisu"`
	Saerro    FactionsWithTotal `json:"saerro"`
	Sanctuary FactionsWithTotal `json:"sanctuary"`
	VoidWell  FactionsWithTotal `json:"voidwell"`
}

type WorldPopulation struct {
	Id       int                `json:"id"`
	Average  int                `json:"average"`
	Factions Factions[int]      `json:"factions"`
	Services PopulationServices `json:"services"`
	CachedAt string             `json:"cached_at"`
}

type AlertState struct {
	Id          int               `json:"id"`
	Zone        int               `json:"zone"`
	EndTime     string            `json:"end_time"`
	StartTime   string            `json:"start_time"`
	AlertType   string            `json:"alert_type"`
	Ps2alerts   string            `json:"ps2alerts"`
	Percentages Factions[float64] `json:"percentages"`
}

type ZoneState struct {
	Id          int               `json:"id"`
	Locked      bool              `json:"locked"`
	Alert       AlertState        `json:"alert"`
	Territory   Factions[float64] `json:"territory"`
	LockedSince string            `json:"locked_since"`
}

type WorldState struct {
	Id       int         `json:"id"`
	Zones    []ZoneState `json:"zones"`
	CachedAt string      `json:"cached_at"`
}
